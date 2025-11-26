package main

import (
	"encoding/json"
	"fmt"
	"time"

	packtrack "github.com/commandant-labs/pack-track-sdk"
)

func buildEventFromConfig(cfg *Config) (packtrack.Event, error) {
	var ts time.Time
	var err error
	if cfg.Timestamp != "" {
		ts, err = time.Parse(time.RFC3339, cfg.Timestamp)
		if err != nil {
			return packtrack.Event{}, fmt.Errorf("invalid --timestamp: %w", err)
		}
	} else {
		ts = time.Now().UTC()
	}
	if cfg.SourceSystem == "" {
		return packtrack.Event{}, fmt.Errorf("missing required --source-system")
	}
	if cfg.WorkflowID == "" {
		return packtrack.Event{}, fmt.Errorf("missing required --workflow-id")
	}
	if cfg.ActorType == "" {
		return packtrack.Event{}, fmt.Errorf("missing required --actor-type")
	}
	if cfg.ActorID == "" {
		return packtrack.Event{}, fmt.Errorf("missing required --actor-id")
	}
	if cfg.Severity == "" {
		return packtrack.Event{}, fmt.Errorf("missing required --severity")
	}
	if cfg.Status == "" {
		return packtrack.Event{}, fmt.Errorf("missing required --status")
	}
	if cfg.Message == "" {
		return packtrack.Event{}, fmt.Errorf("missing required --message")
	}

	sev := packtrack.Severity(cfg.Severity)
	switch sev {
	case packtrack.SeverityDebug, packtrack.SeverityInfo, packtrack.SeverityWarn, packtrack.SeverityError:
	default:
		return packtrack.Event{}, fmt.Errorf("invalid --severity: %s", cfg.Severity)
	}
	st := packtrack.Status(cfg.Status)
	switch st {
	case packtrack.StatusRunning, packtrack.StatusSuccess, packtrack.StatusError:
	default:
		return packtrack.Event{}, fmt.Errorf("invalid --status: %s", cfg.Status)
	}

	e := packtrack.Event{
		Timestamp: ts,
		Source:    packtrack.Source{System: cfg.SourceSystem, Env: cfg.SourceEnv},
		Workflow:  packtrack.Workflow{ID: cfg.WorkflowID, Name: cfg.WorkflowName, RunID: cfg.RunID, StepID: cfg.StepID},
		Actor:     packtrack.Actor{Type: cfg.ActorType, ID: cfg.ActorID, DisplayName: cfg.ActorDisplay},
		Severity:  sev,
		Status:    st,
		Message:   cfg.Message,
	}
	if cfg.MetadataJSON != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(cfg.MetadataJSON), &m); err != nil {
			return packtrack.Event{}, fmt.Errorf("invalid --metadata JSON: %w", err)
		}
		e.Metadata = m
	}
	if cfg.ExtraJSON != "" {
		var m map[string]any
		if err := json.Unmarshal([]byte(cfg.ExtraJSON), &m); err != nil {
			return packtrack.Event{}, fmt.Errorf("invalid --extra JSON: %w", err)
		}
		e.Extra = m
	}
	return e, nil
}
