package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ares/dp-vc-webApp/configs/env"
	"github.com/ares/dp-vc-webApp/configs/logger"
	appredis "github.com/ares/dp-vc-webApp/configs/redis"
	paymentservice "github.com/ares/dp-vc-webApp/services/payments"
	"github.com/ares/dp-vc-webApp/worker/jobs"
)

var log = logger.Worker()

const (
	// DefaultPollingInterval is the default interval for job polling
	DefaultPollingInterval = 5 * time.Second
	// DefaultLockDuration is the default duration for Redis lock
	DefaultLockDuration = 2 * time.Minute
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
	log.Info("Initializing worker", map[string]interface{}{
		"notification_interval_minutes": env.GetConfig().NotificationIntervalMinutes,
		"payment_reminder_days":         env.GetConfig().PaymentReminderDays,
	})

	worker := New(ctx)
	worker.Run()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Warn("Worker shutdown signal received", map[string]interface{}{
		"signal": sig.String(),
	})
	worker.Stop()
}

// Run executes the worker loop
func (w *Worker) Run() {
	// Start job processors
	go w.processJobs()
	go w.processScheduledTasks()

	log.Info("Worker started successfully", nil)
}

// RunNotificationsOnce checks MongoDB and sends due payment notifications once.
// It is intended for system cron, Replit scheduled jobs, or similar schedulers.
func RunNotificationsOnce(ctx context.Context) error {
	return RunNotificationsOnceWithLock(ctx, DefaultLockDuration)
}

// RunNotificationsOnceWithLock checks MongoDB and sends due payment notifications with custom lock duration
func RunNotificationsOnceWithLock(ctx context.Context, lockDuration time.Duration) error {
	startTime := time.Now()
	log.Info("Starting notification scan", nil)

	// Try to acquire lock
	locked, err := appredis.AcquireLock(ctx, "payment-notifications", lockDuration)
	if err != nil {
		log.Error("Failed to acquire lock", err, nil)
		return fmt.Errorf("acquire lock: %w", err)
	}
	if !locked {
		log.Info("Skipping notification scan - another instance holds the lock", nil)
		return nil
	}

	defer func() {
		// Release lock
		appredis.ReleaseLock(ctx, "payment-notifications")
	}()

	// Send notifications
	err = paymentservice.New(env.GetConfig()).SendUpcomingNotifications(ctx, time.Now().UTC())
	duration := time.Since(startTime)

	if err != nil {
		log.Error("Notification scan failed", err, map[string]interface{}{
			"duration_ms": duration.Milliseconds(),
		})
		return err
	}

	log.Info("Notification scan completed", map[string]interface{}{
		"duration_ms": duration.Milliseconds(),
	})
	return nil
}

// Stop gracefully stops the worker
func (w *Worker) Stop() {
	log.Info("Stopping worker...", nil)
	w.cancel()
	time.Sleep(2 * time.Second) // Allow time for current jobs to complete
	log.Info("Worker stopped gracefully", nil)
}

// processJobs handles incoming jobs from Redis queues
func (w *Worker) processJobs() {
	log.Info("Job processor started", map[string]interface{}{
		"polling_interval": DefaultPollingInterval.String(),
	})

	ticker := time.NewTicker(DefaultPollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			log.Info("Job processor stopped", nil)
			return
		case <-ticker.C:
			// Poll for jobs from Redis
			w.pollJobs()
		}
	}
}

// processScheduledTasks handles scheduled/recurring tasks
func (w *Worker) processScheduledTasks() {
	interval := time.Duration(env.GetConfig().NotificationIntervalMinutes) * time.Minute
	if interval < time.Minute {
		interval = time.Minute
	}

	log.Info("Scheduled task processor started", map[string]interface{}{
		"interval": interval.String(),
	})

	// Run immediately on start
	w.runScheduledTasks()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			log.Info("Scheduled task processor stopped", nil)
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
		if err := RunNotificationsOnce(w.ctx); err != nil {
			log.Error("Payment notification scan failed", err, nil)
		}
	}
}
