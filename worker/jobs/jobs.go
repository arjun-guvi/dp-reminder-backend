package jobs

import (
	"context"
	"fmt"
	"time"
)

// Job represents a background job
type Job struct {
	ID      string
	Name    string
	Payload interface{}
	Attempt int
}

// JobFunc is the function signature for job handlers
type JobFunc func(context.Context, *Job) error

// ExampleJob is a harmless example job for testing
func ExampleJob(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		fmt.Printf("[Job] Example job executed at %s\n", time.Now().Format(time.RFC3339))
		return nil
	}
}

// ExampleJobWithData demonstrates job execution with data
func ExampleJobWithData(ctx context.Context, data map[string]interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		fmt.Printf("[Job] Processing job with data: %+v\n", data)
		// Simulate work
		time.Sleep(100 * time.Millisecond)
		return nil
	}
}

// JobRegistry holds registered job handlers
var JobRegistry = make(map[string]JobFunc)

// RegisterJob registers a job handler by name
func RegisterJob(name string, handler JobFunc) {
	JobRegistry[name] = handler
}

// GetJobHandler returns the handler for a job name
func GetJobHandler(name string) (JobFunc, bool) {
	handler, exists := JobRegistry[name]
	return handler, exists
}

func init() {
	// Register built-in jobs
	RegisterJob("example", func(ctx context.Context, job *Job) error {
		return ExampleJobWithData(ctx, map[string]interface{}{
			"job_id": job.ID,
		})
	})
}

// ExecuteJob executes a job by name
func ExecuteJob(ctx context.Context, name string, job *Job) error {
	handler, exists := GetJobHandler(name)
	if !exists {
		return fmt.Errorf("unknown job: %s", name)
	}

	return handler(ctx, job)
}
