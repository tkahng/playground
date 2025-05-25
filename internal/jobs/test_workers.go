package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// EmailJobArgs is defined in server/handlers.go, so you'd need to export it or
// have your workers accept a generic interface if types are in different packages.
// For simplicity, let's define it here too, or ensure it's imported correctly.
// A common pattern is to put JobArgs structs in a shared 'models' or 'jobtypes' package.
type EmailJobArgs struct {
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

func (e EmailJobArgs) Kind() string {
	return "email_send"
}

// EmailWorker implements the Worker interface for email jobs.
type EmailWorker struct {
	WorkFunc func(ctx context.Context, job *Job[EmailJobArgs]) error
}

var _ Worker[EmailJobArgs] = (*EmailWorker)(nil)

func (w *EmailWorker) Work(ctx context.Context, job *Job[EmailJobArgs]) error {
	if w.WorkFunc != nil {
		return w.WorkFunc(ctx, job)
	}
	slog.InfoContext(ctx, "Processing email job",
		"job_id", job.ID.String(),
		"recipient", job.Args.Recipient,
		"subject", job.Args.Subject,
	)

	// Simulate work, potentially long-running or prone to failure
	time.Sleep(2 * time.Second) // Simulate sending email

	if job.Args.Recipient == "fail@example.com" {
		return fmt.Errorf("simulated email send failure to %s", job.Args.Recipient)
	}

	// Check if context was cancelled during work
	select {
	case <-ctx.Done():
		slog.WarnContext(ctx, "Email job cancelled", "job_id", job.ID.String(), "error", ctx.Err())
		return ctx.Err() // Return context error if cancelled
	default:
		// Continue
	}

	slog.InfoContext(ctx, "Email job completed successfully", "job_id", job.ID.String())
	return nil
}

// ReportJobArgs for a different kind of job.
type ReportJobArgs struct {
	ReportID string `json:"report_id"`
	UserID   string `json:"user_id"`
	Format   string `json:"format"`
}

func (r ReportJobArgs) Kind() string {
	return "generate_report"
}

// ReportWorker implements the Worker interface for report generation jobs.
type ReportWorker struct {
	WorkFunc func(ctx context.Context, job *Job[ReportJobArgs]) error
}

var _ Worker[ReportJobArgs] = (*ReportWorker)(nil)

func (w *ReportWorker) Work(ctx context.Context, job *Job[ReportJobArgs]) error {
	if w.WorkFunc != nil {
		return w.WorkFunc(ctx, job)
	}
	slog.InfoContext(ctx, "Generating report",
		"job_id", job.ID.String(),
		"report_id", job.Args.ReportID,
		"user_id", job.Args.UserID,
		"format", job.Args.Format,
	)

	// Simulate heavy computation
	time.Sleep(5 * time.Second)

	// Check for cancellation
	select {
	case <-ctx.Done():
		slog.WarnContext(ctx, "Report generation job cancelled", "job_id", job.ID.String(), "error", ctx.Err())
		return ctx.Err()
	default:
		// Continue
	}

	if job.Args.ReportID == "invalid" {
		return fmt.Errorf("simulated invalid report ID for user %s", job.Args.UserID)
	}

	slog.InfoContext(ctx, "Report generated successfully", "job_id", job.ID.String())
	return nil
}
