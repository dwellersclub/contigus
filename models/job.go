package models

import (
	"time"
)

type jobStatus string

type jobStatuses struct {
	Pending   jobStatus
	Running   jobStatus
	Succeeded jobStatus
	Failed    jobStatus
	Unknown   jobStatus
}

//JobEnumStatus Job status
var JobEnumStatus = jobStatuses{
	Pending:   jobStatus("Pending"),
	Running:   jobStatus("Running"),
	Succeeded: jobStatus("Succeeded"),
	Failed:    jobStatus("Failed"),
	Unknown:   jobStatus("Unknown"),
}

//Job Job to Execute
type Job struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Image        string    `json:"image"`
	CreationTime time.Time `json:"creationTime"`
	StartTime    time.Time `json:"startTime"`
	EndTime      time.Time `json:"endTime"`
	ExitCode     int32     `json:"exitCode"`
	Status       jobStatus `json:"status"`
}

//Worker Job being executed
type Worker struct {
	ID        string `json:"id"`
	BuildID   string `json:"buildId"`
	ProjectID string `json:"projectId"`
	JobID     string `json:"jobId"`
}
