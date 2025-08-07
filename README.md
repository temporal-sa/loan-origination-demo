# Loan Origination System

A complete loan origination system built with Go, Temporal, and a simple SPA frontend. This system demonstrates human-in-the-loop workflows with Temporal signals for document verification, appraisal, and underwriting processes. All state is managed within Temporal workflows using queries for data retrieval.

## Features

- **Backend**: Go with Gin framework
- **Workflow Engine**: Temporal for reliable, durable execution
- **State Management**: All data stored in Temporal workflow state (no database required)
- **Frontend**: Vanilla JavaScript SPA with role-based interface
- **Human-in-the-Loop**: Manual processes with signals for document verification and underwriting
- **Queries**: Temporal workflow queries for data retrieval

## Architecture

The system includes 6 personas:
1. **Loan Officer** - Create loan applications
2. **Customer** - Upload documents
3. **Loan Processor** - Verify documents
4. **Appraiser** - Complete property appraisals
5. **Underwriter** - Make loan decisions
6. **Fund Manager** - Process funding

## Prerequisites

- Go 1.21 or later
- Temporal CLI

Install the Temporal CLI from: https://docs.temporal.io/cli#installation

## How to Run

### 1. Start Temporal Server
```bash
temporal server start-dev
```

### 2. Start the Temporal Worker (in another terminal)
```bash
go run cmd/worker/main.go
```

### 3. Start the API Server (in another terminal)
```bash
go run cmd/server/main.go
```

Open your browser and navigate to:
- **Frontend**: http://localhost:8082
- **Temporal Web UI**: http://localhost:8233

## How to Use

### Complete End-to-End Demo

1. **Loan Officer**: 
   - Switch to "Loan Officer" role
   - Create a new loan application using the form
   - View your created applications

2. **Customer**: 
   - Switch to "Customer" role
   - Upload documents for pending applications
   - Upload at least 2 documents (income statement, bank statement)

3. **Loan Processor**: 
   - Switch to "Loan Processor" role
   - Verify uploaded documents by clicking "Verify Documents"
   - Approve or reject each document

4. **Appraiser**: 
   - Switch to "Appraiser" role
   - Complete property appraisal for applications
   - Enter property value and notes

5. **Underwriter**: 
   - Switch to "Underwriter" role
   - Review applications with completed appraisals
   - Make approve/reject decisions

6. **Fund Manager**: 
   - Switch to "Fund Manager" role
   - Process funding for approved loans

### Testing Third-Party Integration

You can simulate third-party document verification using curl:

```bash
# Get the loan ID and document ID from the frontend first
curl -X POST http://localhost:8082/api/v1/loans/{loan-id}/verify-documents \
  -H "Content-Type: application/json" \
  -d '{
    "document_id": "{document-id}",
    "verification_status": "verified",
    "verification_details": {
      "verified_by": "third-party-service",
      "confidence_score": 0.95
    }
  }'
```

## API Endpoints

- `POST /api/v1/loans` - Create loan application
- `GET /api/v1/loans` - Get all loan applications
- `GET /api/v1/loans/:id` - Get specific loan application
- `GET /api/v1/loans/:id/status` - Get workflow status
- `POST /api/v1/loans/:id/documents` - Upload document
- `POST /api/v1/loans/:id/verify-documents` - Verify document
- `POST /api/v1/loans/:id/appraisal` - Complete appraisal
- `POST /api/v1/loans/:id/underwriting` - Make underwriting decision
- `POST /api/v1/loans/:id/funding` - Process funding

## Temporal Features Demonstrated

- **Long-running workflows** - Loan origination process can take days/weeks
- **Human-in-the-loop** - Manual steps for document upload, verification, and underwriting
- **Signal handling** - External events trigger workflow progression
- **Workflow state management** - All data stored in workflow state
- **Workflow queries** - Real-time data retrieval from running workflows
- **Timeout management** - Workflows have timeouts for each step
- **Workflow history** - Complete audit trail of all actions