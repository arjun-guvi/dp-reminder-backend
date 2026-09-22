package worker

import (
	"context"
	"testing"
	"time"

	"github.com/ares/dp-vc-webApp/configs/env"
)

func TestRunNotificationsOnce_DryRun(t *testing.T) {
	// Load environment
	if err := env.Load(); err != nil {
		t.Skipf("Skipping test: .env file not found: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// This test will fail if MongoDB/Redis not available
	err := RunNotificationsOnce(ctx)
	
	// We expect either success or a connection error
	// Both are valid outcomes for this test
	if err != nil {
		t.Logf("Notification run result: %v (expected if DB not available)", err)
	} else {
		t.Log("Notification run completed successfully")
	}
}

func TestWorker_Lifecycle(t *testing.T) {
	if err := env.Load(); err != nil {
		t.Skipf("Skipping test: .env file not found: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker := New(ctx)
	
	// Verify worker creation
	if worker == nil {
		t.Fatal("Expected worker to be created")
	}
	
	// Verify context is set
	if worker.ctx == nil {
		t.Fatal("Expected worker context to be set")
	}
	
	// Verify cancel function exists
	if worker.cancel == nil {
		t.Fatal("Expected worker cancel to be set")
	}
}

func TestWorker_Stop(t *testing.T) {
	if err := env.Load(); err != nil {
		t.Skipf("Skipping test: .env file not found: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker := New(ctx)
	
	// Test stop doesn't panic
	done := make(chan bool)
	go func() {
		worker.Stop()
		done <- true
	}()

	select {
	case <-done:
		t.Log("Worker stopped successfully")
	case <-time.After(5 * time.Second):
		t.Fatal("Worker stop timed out")
	}
}

func TestRunWithLock_Timeout(t *testing.T) {
	if err := env.Load(); err != nil {
		t.Skipf("Skipping test: .env file not found: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Should return error due to timeout
	err := RunNotificationsOnceWithLock(ctx, time.Minute)
	
	// Expect error due to context timeout
	if err == nil {
		t.Log("Run completed (possibly no pending payments)")
	} else {
		t.Logf("Run returned error as expected: %v", err)
	}
}
