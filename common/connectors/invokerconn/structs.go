//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=JobType

package invokerconn

import (
	"fmt"
	"time"
)

type JobType int

const (
	CompileJob JobType = iota + 1
	TestJob
)

type Job struct {
	ID       string  `json:"id"  binding:"required"`
	SubmitID uint    `json:"submit_id" binding:"required"`
	Type     JobType `json:"type" binding:"required"`
	Test     uint64  `json:"test"`

	RequiredJobIDs []string
}

func (j Job) String() string {
	return fmt.Sprintf("ID: %s Submit: %d Type %v Test: %d", j.ID, j.SubmitID, j.Type, j.Test)
}

type Status struct {
	MaxNewJobs   int      `json:"max_new_jobs"`
	ActiveJobIDs []string `json:"active_job_ids"`
	Epoch        string   `json:"epoch"`
	Address      string   `json:"address"`

	Metrics *StatusMetrics `json:"metrics"`
}

type StatusMetrics struct {
	Lifetime       time.Duration         `json:"life_time"`
	SandboxMetrics *StatusThreadsMetrics `json:"sandbox_metrics"`
	ThreadMetrics  *StatusThreadsMetrics `json:"thread_metrics"`
}

type StatusThreadsMetrics struct {
	Count             int             `json:"count"`
	TotalWaitDuration []time.Duration `json:"total_wait_duration"`
}
