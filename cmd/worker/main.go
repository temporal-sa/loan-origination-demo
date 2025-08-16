package main

import (
	"log"

	"loan-origination-system/internal/activities"
	"loan-origination-system/internal/workflows"
	"loan-origination-system/pkg/temporal"
)

func main() {
	// Create Temporal client
	c, err := temporal.NewClient()
	if err != nil {
		log.Fatal("Unable to create Temporal client:", err)
	}
	defer c.Close()

	// Create worker
	w := temporal.NewWorker(c)

	// Register workflows
	w.RegisterWorkflow(workflows.LoanOriginationWorkflow)

	// Register activities
	w.RegisterActivity(activities.GenerateLoanAgreement)
	w.RegisterActivity(activities.ProcessFunding)
	w.RegisterActivity(activities.CreditScoreCheck)

	log.Println("Starting Temporal worker...")
	err = w.Run(nil)
	if err != nil {
		log.Fatal("Unable to start worker:", err)
	}
}
