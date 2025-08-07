package api

import (
	"loan-origination-system/internal/api/handlers"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
)

func SetupRoutes(router *gin.Engine, temporalClient client.Client) {
	loanHandler := handlers.NewLoanHandler(temporalClient)

	// API routes
	api := router.Group("/api/v1")
	{
		// Loan application routes
		api.POST("/loans", loanHandler.CreateLoanApplication)
		api.GET("/loans", loanHandler.GetLoanApplications)
		api.GET("/loans/:id", loanHandler.GetLoanApplication)
		api.GET("/loans/:id/status", loanHandler.GetWorkflowStatus)

		// Document routes
		api.POST("/loans/:id/documents", loanHandler.UploadDocument)
		api.POST("/loans/:id/verify-documents", loanHandler.VerifyDocument)

		// Appraisal routes
		api.POST("/loans/:id/appraisal", loanHandler.CompleteAppraisal)

		// Underwriting routes
		api.POST("/loans/:id/underwriting", loanHandler.MakeUnderwritingDecision)

		// Funding routes
		api.POST("/loans/:id/funding", loanHandler.ProcessFunding)
	}

	// Serve static files for frontend
	router.Static("/css", "./web/css")
	router.Static("/js", "./web/js")
	router.StaticFile("/", "./web/index.html")
}
