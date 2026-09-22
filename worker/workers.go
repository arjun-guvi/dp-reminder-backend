package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ares/dp-vc-webApp/configs/env"
	paymentservice "github.com/ares/dp-vc-webApp/services/payments"
	"github.com/ares/dp-vc-webApp/worker/jobs"
)

const (
	// DefaultPollingInterval is the default interval for job polling
	DefaultPollingInterval = 5 * time.Second
)

// Worker represents a background worker
type Worker struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new worker instance
func New(ctx context.Context) *Worker {
	ctx, cancel := context.WithCancel(ctx)
	return &Worker{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins the worker process
func Start(ctx context.Context) {
	fmt.Println("Initializing worker...")

	worker := New(ctx)
	worker.Run()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Worker shutdown signal received")
	worker.Stop()
}

// Run executes the worker loop
func (w *Worker) Run() {
	// Start job processors
	go w.processJobs()
	go w.processScheduledTasks()

	fmt.Println("Worker started successfully")
}

// Stop gracefully stops the worker
func (w *Worker) Stop() {
	fmt.Println("Stopping worker...")
	w.cancel()
	time.Sleep(2 * time.Second) // Allow time for current jobs to complete
	fmt.Println("Worker stopped gracefully")
}

// processJobs handles incoming jobs from Redis queues
func (w *Worker) processJobs() {
	ticker := time.NewTicker(DefaultPollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			// Poll for jobs from Redis
			w.pollJobs()
		}
	}
}

// processScheduledTasks handles scheduled/recurring tasks
func (w *Worker) processScheduledTasks() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			// Run scheduled tasks
			w.runScheduledTasks()
		}
	}
}

// pollJobs polls for jobs from Redis queues
func (w *Worker) pollJobs() {
	select {
	case <-w.ctx.Done():
		return
	default:
		// Execute example job (placeholder)
		jobs.ExampleJob(w.ctx)
	}
}

// runScheduledTasks executes recurring tasks
func (w *Worker) runScheduledTasks() {
	select {
	case <-w.ctx.Done():
		return
	default:
		if err := paymentservice.New(env.GetConfig()).SendUpcomingNotifications(w.ctx, time.Now()); err != nil {
			fmt.Printf("[Worker] payment notification scan failed: %v\n", err)
		}
	}
}
