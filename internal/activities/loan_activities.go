package activities

import (
	"context"
	"fmt"
	"math/rand"
	"time"
	"go.temporal.io/sdk/activity"
)

type GenerateLoanAgreementInput struct {
	LoanApplicationID string `json:"loan_application_id"`
}

type ProcessFundingInput struct {
	LoanApplicationID string `json:"loan_application_id"`
}

type CreditScoreCheckInput struct {
	LoanApplicationID string `json:"loan_application_id"`
	BorrowerName      string `json:"borrower_name"`
}

type CreditScoreCheckResult struct {
	CreditScore int    `json:"credit_score"`
	Status      string `json:"status"`
}

// Activities
func GenerateLoanAgreement(ctx context.Context, input GenerateLoanAgreementInput) error {
	// In a real system, this would integrate with banking systems
	// For demo purposes, we'll just simulate the generating an agreement
	time.Sleep(2 * time.Second) // Simulate processing time
	return nil
}

func ProcessFunding(ctx context.Context, input ProcessFundingInput) error {
	// In a real system, this would integrate with banking systems
	// For demo purposes, we'll just simulate the funding process
	time.Sleep(2 * time.Second) // Simulate processing time
	return nil
}

func CreditScoreCheck(ctx context.Context, input CreditScoreCheckInput) (*CreditScoreCheckResult, error) {
	// Simulate API call processing time
	time.Sleep(1 * time.Second)
	
	// Simulate failure for first 2 attempts, succeed on 3rd attempt
	if activity.GetInfo(ctx).Attempt < 3 {
		return nil, fmt.Errorf("credit score API temporarily unavailable (attempt %d/3)", activity.GetInfo(ctx).Attempt)
	}
	
	// Generate random credit score between 0-1000
	creditScore := rand.Intn(1001)
	
	return &CreditScoreCheckResult{
		CreditScore: creditScore,
		Status:      "completed",
	}, nil
}
