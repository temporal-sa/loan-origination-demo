package workflows

import (
	"fmt"
	"loan-origination-system/internal/activities"
	"time"

	"go.temporal.io/sdk/workflow"
)

// Workflow signals
type DocumentUploadedSignal struct {
	DocumentID   string `json:"document_id"`
	DocumentType string `json:"document_type"`
}

type DocumentVerificationSignal struct {
	DocumentID          string                 `json:"document_id"`
	VerificationStatus  string                 `json:"verification_status"`
	VerificationDetails map[string]interface{} `json:"verification_details"`
}

type UnderwritingDecisionSignal struct {
	Decision      string `json:"decision"`
	Comments      string `json:"comments"`
	UnderwriterID string `json:"underwriter_id"`
}

type AppraisalCompletedSignal struct {
	PropertyValue  float64 `json:"property_value"`
	AppraisalNotes string  `json:"appraisal_notes"`
	AppraiserID    string  `json:"appraiser_id"`
}

type FundingCompletedSignal struct {
	FundManagerID string  `json:"fund_manager_id"`
	FundingAmount float64 `json:"funding_amount"`
	FundingNotes  string  `json:"funding_notes"`
}

// Workflow input
type LoanOriginationWorkflowInput struct {
	LoanApplication LoanApplication `json:"loan_application"`
}

// Loan application data structure
type LoanApplication struct {
	ID            string    `json:"id"`
	BorrowerName  string    `json:"borrower_name"`
	BorrowerEmail string    `json:"borrower_email"`
	BorrowerPhone string    `json:"borrower_phone"`
	LoanAmount    float64   `json:"loan_amount"`
	LoanPurpose   string    `json:"loan_purpose"`
	Status        string    `json:"status"`
	NextStep      string    `json:"next_step"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	WorkflowID    string    `json:"workflow_id"`
}

// Document data structure
type Document struct {
	ID                  string                 `json:"id"`
	DocumentType        string                 `json:"document_type"`
	FileName            string                 `json:"file_name"`
	FilePath            string                 `json:"file_path"`
	VerificationStatus  string                 `json:"verification_status"`
	VerificationDetails map[string]interface{} `json:"verification_details"`
	UploadedAt          time.Time              `json:"uploaded_at"`
	VerifiedAt          *time.Time             `json:"verified_at"`
}

// Appraisal data structure
type Appraisal struct {
	ID             string     `json:"id"`
	PropertyValue  float64    `json:"property_value"`
	AppraisalNotes string     `json:"appraisal_notes"`
	AppraiserID    string     `json:"appraiser_id"`
	Status         string     `json:"status"`
	CompletedAt    *time.Time `json:"completed_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

// Underwriting decision data structure
type UnderwritingDecision struct {
	ID            string    `json:"id"`
	Decision      string    `json:"decision"`
	Comments      string    `json:"comments"`
	UnderwriterID string    `json:"underwriter_id"`
	DecisionDate  time.Time `json:"decision_date"`
}

// Workflow state
type LoanOriginationState struct {
	LoanApplication      LoanApplication       `json:"loan_application"`
	Documents            []Document            `json:"documents"`
	Appraisal            *Appraisal            `json:"appraisal"`
	UnderwritingDecision *UnderwritingDecision `json:"underwriting_decision"`
	Status               string                `json:"status"`
	NextStep             string                `json:"next_step"`
}

func LoanOriginationWorkflow(ctx workflow.Context, input LoanOriginationWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting loan origination workflow", "loanApplicationID", input.LoanApplication.ID)

	// Set up activity options with timeouts
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Initialize workflow state with loan application data
	state := &LoanOriginationState{
		LoanApplication: input.LoanApplication,
		Documents:       []Document{},
		Status:          "processing",
	}

	// Update loan status to processing
	state.LoanApplication.Status = "processing"
	state.Status = "processing"

	// Set up query handlers
	err := workflow.SetQueryHandler(ctx, "getLoanApplication", func() (LoanOriginationState, error) {
		return *state, nil
	})
	if err != nil {
		return err
	}

	// Generate loan agreement
	workflow.ExecuteActivity(ctx, activities.GenerateLoanAgreement, activities.GenerateLoanAgreementInput{
		LoanApplicationID: state.LoanApplication.ID,
	})

	err = runWorkflowSteps(ctx, state)
	if err != nil {
		return err
	}

	// Process based on underwriting decision
	if state.UnderwritingDecision != nil {
		if state.UnderwritingDecision.Decision == "approved" {
			state.LoanApplication.Status = "approved"
			state.Status = "approved"

			// Wait for funding completion
			err = waitForFunding(ctx, state)
			if err != nil {
				return err
			}

			// Release funds
			workflow.ExecuteActivity(ctx, activities.ProcessFunding, activities.ProcessFundingInput{
				LoanApplicationID: state.LoanApplication.ID,
			}).Get(ctx, nil)
		} else {
			state.LoanApplication.Status = "rejected"
			state.Status = "rejected"
		}
	} else {
		state.LoanApplication.Status = "incomplete"
		state.Status = "incomplete"
	}
	state.NextStep = "n/a"

	logger.Info("Loan origination workflow completed", "loanApplicationID", input.LoanApplication.ID, "status", state.Status)
	return nil
}

func waitForFunding(ctx workflow.Context, state *LoanOriginationState) error {
	logger := workflow.GetLogger(ctx)

	state.NextStep = "Waiting for funding"

	// Set up signal channel for funding completion
	fundingChannel := workflow.GetSignalChannel(ctx, "funding-completed")

	selector := workflow.NewSelector(ctx)

	// Listen for funding completion signal
	selector.AddReceive(fundingChannel, func(c workflow.ReceiveChannel, more bool) {
		var signal FundingCompletedSignal
		c.Receive(ctx, &signal)

		state.LoanApplication.Status = "funded"
		state.Status = "funded"
		logger.Info("Funding completed", "fundManagerID", signal.FundManagerID, "amount", signal.FundingAmount)
	})

	// Add timeout for funding (7 days)
	selector.AddFuture(workflow.NewTimer(ctx, 7*24*time.Hour), func(f workflow.Future) {
		logger.Error("Timeout waiting for funding completion")
		state.LoanApplication.Status = "funding_timeout"
		state.Status = "funding_timeout"
	})

	selector.Select(ctx)

	return nil
}

func runWorkflowSteps(ctx workflow.Context, state *LoanOriginationState) error {
	logger := workflow.GetLogger(ctx)

	// Set up signal channels
	documentUploadChannel := workflow.GetSignalChannel(ctx, "document-uploaded")
	verificationChannel := workflow.GetSignalChannel(ctx, "document-verified")
	appraisalChannel := workflow.GetSignalChannel(ctx, "appraisal-completed")
	underwritingChannel := workflow.GetSignalChannel(ctx, "underwriting-decision")

	// Track completion status
	requiredDocuments := 2
	appraisalCompleted := state.Appraisal != nil
	underwritingCompleted := false

	timerCtx, timerCancel := workflow.WithCancel(ctx)
	timer := workflow.NewTimer(timerCtx, 30*24*time.Hour)

	// Main workflow loop - listen for all signals
	for !underwritingCompleted {
		verifiedDocCount := 0
		// Count already verified documents
		for _, doc := range state.Documents {
			if doc.VerificationStatus == "verified" {
				verifiedDocCount++
			}
		}

		// Count rejected documents
		rejectedDocCount := 0
		for _, doc := range state.Documents {
			if doc.VerificationStatus == "rejected" {
				rejectedDocCount++
			}
		}

		moreDocsRequired := (len(state.Documents) - rejectedDocCount) < requiredDocuments

		selector := workflow.NewSelector(ctx)

		switch {
		// Listen for document verification (only if we have uploaded documents)
		case len(state.Documents) > 0 && verifiedDocCount+rejectedDocCount < len(state.Documents):
			state.NextStep = "Waiting for document verification"

			selector.AddReceive(verificationChannel, func(c workflow.ReceiveChannel, more bool) {
				var signal DocumentVerificationSignal
				c.Receive(ctx, &signal)

				// Update document verification status
				for i, doc := range state.Documents {
					if doc.ID == signal.DocumentID {
						now := workflow.Now(ctx)
						state.Documents[i].VerificationStatus = signal.VerificationStatus
						state.Documents[i].VerificationDetails = signal.VerificationDetails
						state.Documents[i].VerifiedAt = &now

						switch signal.VerificationStatus {
						case "verified":
							verifiedDocCount++
						case "rejected":
							rejectedDocCount++
						}

						logger.Info("Document", signal.VerificationStatus, "documentID", signal.DocumentID, "status", signal.VerificationStatus, "verified", verifiedDocCount, "rejected", rejectedDocCount)
						break
					}
				}
			})

			fallthrough

		// Listen for document uploads
		case moreDocsRequired:
			if moreDocsRequired {
				state.NextStep = fmt.Sprintf("Waiting for customer documents: %d more required", requiredDocuments-len(state.Documents)+rejectedDocCount)
			}

			selector.AddReceive(documentUploadChannel, func(c workflow.ReceiveChannel, more bool) {
				var signal DocumentUploadedSignal
				c.Receive(ctx, &signal)

				// Create document record
				doc := Document{
					ID:                 signal.DocumentID,
					DocumentType:       signal.DocumentType,
					FileName:           signal.DocumentType + "_document.pdf", // Simplified
					FilePath:           "/uploads/" + signal.DocumentID,
					VerificationStatus: "pending",
					UploadedAt:         workflow.Now(ctx),
				}
				state.Documents = append(state.Documents, doc)
				logger.Info("Document uploaded", "documentID", signal.DocumentID, "type", signal.DocumentType, "count", len(state.Documents))
			})

		// Listen for appraisal completion
		case !appraisalCompleted:

			state.NextStep = "Waiting for appraisal"

			selector.AddReceive(appraisalChannel, func(c workflow.ReceiveChannel, more bool) {
				var signal AppraisalCompletedSignal
				c.Receive(ctx, &signal)

				now := workflow.Now(ctx)
				state.Appraisal = &Appraisal{
					ID:             "appraisal-" + state.LoanApplication.ID,
					PropertyValue:  signal.PropertyValue,
					AppraisalNotes: signal.AppraisalNotes,
					AppraiserID:    signal.AppraiserID,
					Status:         "completed",
					CompletedAt:    &now,
					CreatedAt:      now,
				}
				appraisalCompleted = true
				logger.Info("Appraisal completed", "propertyValue", signal.PropertyValue)
			})

		// Listen for underwriting decision (only if we have enough documents and appraisal is done)
		// Allow underwriting even if some documents are rejected - underwriter can decide
		case len(state.Documents) >= requiredDocuments && appraisalCompleted:

			state.NextStep = "Waiting for underwriting decision"

			selector.AddReceive(underwritingChannel, func(c workflow.ReceiveChannel, more bool) {
				var signal UnderwritingDecisionSignal
				c.Receive(ctx, &signal)

				state.UnderwritingDecision = &UnderwritingDecision{
					ID:            "decision-" + state.LoanApplication.ID,
					Decision:      signal.Decision,
					Comments:      signal.Comments,
					UnderwriterID: signal.UnderwriterID,
					DecisionDate:  workflow.Now(ctx),
				}
				if signal.Decision != "needs_more_info" {
					underwritingCompleted = true
				} else {
					requiredDocuments++
				}

				logger.Info("Underwriting decision received", "decision", signal.Decision)
			})

		}

		// Add timeout to prevent infinite waiting
		selector.AddFuture(timer, func(f workflow.Future) {
			logger.Error("Workflow timeout - completing with current state")
			underwritingCompleted = true
		})

		selector.Select(ctx)
	}

	timerCancel()

	return nil
}
