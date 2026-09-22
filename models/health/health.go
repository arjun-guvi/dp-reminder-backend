package health

import (
	"time"
)

// Health represents the health status document
type Health struct {
	Service   string    `json:"service" bson:"service"`
	Status    string    `json:"status" bson:"status"`
	Version   string    `json:"version" bson:"version"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
}

// ServiceDependency represents a dependency's health status
type ServiceDependency struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Latency int64  `json:"latency_ms"`
	Error   string `json:"error,omitempty"`
}

// HealthCheckResult aggregates all dependency health checks
type HealthCheckResult struct {
	Service      string              `json:"service"`
	Status       string              `json:"status"`
	Dependencies []ServiceDependency `json:"dependencies"`
	Timestamp    time.Time           `json:"timestamp"`
}
