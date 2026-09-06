// Package incident defines the payment-incident case lifecycle.
package incident

import (
	"fmt"
	"strings"
	"time"
)

type Status string

const (
	StatusPendingVerification  Status = "pending_verification"
	StatusInvestigating        Status = "investigating"
	StatusAwaitingConfirmation Status = "awaiting_confirmation"
	StatusResolved             Status = "resolved"
	StatusEscalated            Status = "escalated"
)

// PaymentIncident is a human-owned case. It deliberately stores neither a
// payment-operation command nor executable compensation parameters.
type PaymentIncident struct {
	ID                  string    `json:"id"`
	TenantID            int64     `json:"tenant_id"`
	CreatedBy           int64     `json:"created_by"`
	OrderID             string    `json:"order_id"`
	Channel             string    `json:"channel"`
	IncidentType        string    `json:"incident_type"`
	Priority            string    `json:"priority"`
	Status              Status    `json:"status"`
	AssigneeID          *int64    `json:"assignee_id,omitempty"`
	AssessmentRequestID string    `json:"assessment_request_id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func NewPaymentIncident(id string, tenantID, createdBy int64, incidentType string) (*PaymentIncident, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("incident id is required")
	}
	if tenantID <= 0 {
		return nil, fmt.Errorf("tenant id must be positive")
	}
	if createdBy <= 0 {
		return nil, fmt.Errorf("creator id must be positive")
	}
	if strings.TrimSpace(incidentType) == "" {
		return nil, fmt.Errorf("incident type is required")
	}
	now := time.Now()
	return &PaymentIncident{
		ID:           id,
		TenantID:     tenantID,
		CreatedBy:    createdBy,
		IncidentType: strings.TrimSpace(incidentType),
		Priority:     "P2",
		Status:       StatusPendingVerification,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (i *PaymentIncident) CanTransitionTo(target Status) bool {
	if i == nil || target == "" || i.Status == target {
		return false
	}
	switch i.Status {
	case StatusPendingVerification:
		return target == StatusInvestigating || target == StatusEscalated
	case StatusInvestigating:
		return target == StatusAwaitingConfirmation || target == StatusResolved || target == StatusEscalated
	case StatusAwaitingConfirmation:
		return target == StatusInvestigating || target == StatusResolved || target == StatusEscalated
	default:
		return false
	}
}

func (i *PaymentIncident) TransitionTo(target Status) error {
	if !i.CanTransitionTo(target) {
		return fmt.Errorf("invalid incident status transition: %s -> %s", i.Status, target)
	}
	i.Status = target
	i.UpdatedAt = time.Now()
	return nil
}

func (i *PaymentIncident) IsTerminal() bool {
	return i != nil && (i.Status == StatusResolved || i.Status == StatusEscalated)
}
