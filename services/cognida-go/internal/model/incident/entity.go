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
	ID                    string `json:"id"`
	TenantID              int64  `json:"tenant_id"`
	CreatedBy             int64  `json:"created_by"`
	OrderID               string `json:"order_id"`
	Channel               string `json:"channel"`
	IncidentType          string `json:"incident_type"`
	Priority              string `json:"priority"`
	Status                Status `json:"status"`
	AssigneeID            *int64 `json:"assignee_id,omitempty"`
	AssessmentRequestID   string `json:"assessment_request_id"`
	RecommendationVersion string `json:"recommendation_version"`
	// EvidenceSnapshot is the immutable JSON snapshot supplied by the evidence-bound
	// assessment which led to this case. It is descriptive evidence, never an action.
	EvidenceSnapshot string    `json:"evidence_snapshot"`
	Conclusion       string    `json:"conclusion"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TimelineEvent records a human-visible, tenant-scoped case history entry.
// Attachment references are metadata only; this MVP deliberately does not upload
// or execute files.
type TimelineEvent struct {
	ID            string    `json:"id"`
	IncidentID    string    `json:"incident_id"`
	TenantID      int64     `json:"tenant_id"`
	ActorID       int64     `json:"actor_id"`
	EventType     string    `json:"event_type"`
	FromStatus    Status    `json:"from_status,omitempty"`
	ToStatus      Status    `json:"to_status,omitempty"`
	Content       string    `json:"content,omitempty"`
	AttachmentRef string    `json:"attachment_ref,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// SimulatedOrder and its related records are deliberately read-only fixtures
// for the payment-incident prototype. They never connect to a real PSP.
type SimulatedOrder struct {
	ID             string    `json:"id"`
	TenantID       int64     `json:"tenant_id"`
	Status         string    `json:"status"`
	AmountMinor    int64     `json:"amount_minor"`
	Currency       string    `json:"currency"`
	PaymentChannel string    `json:"payment_channel"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SimulatedPayment struct {
	ID              string    `json:"id"`
	TenantID        int64     `json:"tenant_id"`
	OrderID         string    `json:"order_id"`
	ProviderTradeNo string    `json:"provider_trade_no"`
	Status          string    `json:"status"`
	AmountMinor     int64     `json:"amount_minor"`
	OccurredAt      time.Time `json:"occurred_at"`
}

type SimulatedCallbackLog struct {
	ID             string    `json:"id"`
	TenantID       int64     `json:"tenant_id"`
	OrderID        string    `json:"order_id"`
	PaymentID      string    `json:"payment_id"`
	EventType      string    `json:"event_type"`
	DeliveryStatus string    `json:"delivery_status"`
	PayloadSummary string    `json:"payload_summary"`
	ReceivedAt     time.Time `json:"received_at"`
}

type DispositionDraft struct {
	ID              string    `json:"id"`
	TenantID        int64     `json:"tenant_id"`
	IncidentID      string    `json:"incident_id"`
	CreatedBy       int64     `json:"created_by"`
	ApprovedBy      *int64    `json:"approved_by,omitempty"`
	ActionType      string    `json:"action_type"`
	Status          string    `json:"status"`
	IdempotencyKey  string    `json:"idempotency_key"`
	ExecutionResult string    `json:"execution_result"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
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
