package mysql

import (
	"time"

	"cognida/internal/model/incident"
)

type PaymentIncidentModel struct {
	ID                    string `gorm:"primaryKey;size:64"`
	TenantID              int64  `gorm:"index"`
	CreatedBy             int64
	OrderID               string `gorm:"size:128"`
	Channel               string `gorm:"size:64"`
	IncidentType          string `gorm:"size:64"`
	Priority              string `gorm:"size:8"`
	Status                string `gorm:"size:32"`
	AssigneeID            *int64
	AssessmentRequestID   string `gorm:"size:64"`
	RecommendationVersion string `gorm:"size:64"`
	EvidenceSnapshot      string `gorm:"type:longtext"`
	Conclusion            string `gorm:"type:text"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (PaymentIncidentModel) TableName() string { return "payment_incidents" }

func (m *PaymentIncidentModel) ToDomain() *incident.PaymentIncident {
	return &incident.PaymentIncident{ID: m.ID, TenantID: m.TenantID, CreatedBy: m.CreatedBy, OrderID: m.OrderID, Channel: m.Channel, IncidentType: m.IncidentType, Priority: m.Priority, Status: incident.Status(m.Status), AssigneeID: m.AssigneeID, AssessmentRequestID: m.AssessmentRequestID, RecommendationVersion: m.RecommendationVersion, EvidenceSnapshot: m.EvidenceSnapshot, Conclusion: m.Conclusion, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}

func paymentIncidentFromDomain(v *incident.PaymentIncident) *PaymentIncidentModel {
	return &PaymentIncidentModel{ID: v.ID, TenantID: v.TenantID, CreatedBy: v.CreatedBy, OrderID: v.OrderID, Channel: v.Channel, IncidentType: v.IncidentType, Priority: v.Priority, Status: string(v.Status), AssigneeID: v.AssigneeID, AssessmentRequestID: v.AssessmentRequestID, RecommendationVersion: v.RecommendationVersion, EvidenceSnapshot: v.EvidenceSnapshot, Conclusion: v.Conclusion, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}

type PaymentIncidentEventModel struct {
	ID            string `gorm:"primaryKey;size:64"`
	IncidentID    string `gorm:"index;size:64"`
	TenantID      int64  `gorm:"index"`
	ActorID       int64
	EventType     string `gorm:"size:64"`
	FromStatus    string `gorm:"size:32"`
	ToStatus      string `gorm:"size:32"`
	Content       string `gorm:"type:text"`
	AttachmentRef string `gorm:"size:512"`
	CreatedAt     time.Time
}

func (PaymentIncidentEventModel) TableName() string { return "payment_incident_events" }

func (m *PaymentIncidentEventModel) ToDomain() *incident.TimelineEvent {
	return &incident.TimelineEvent{ID: m.ID, IncidentID: m.IncidentID, TenantID: m.TenantID, ActorID: m.ActorID, EventType: m.EventType, FromStatus: incident.Status(m.FromStatus), ToStatus: incident.Status(m.ToStatus), Content: m.Content, AttachmentRef: m.AttachmentRef, CreatedAt: m.CreatedAt}
}

func paymentIncidentEventFromDomain(v *incident.TimelineEvent) *PaymentIncidentEventModel {
	return &PaymentIncidentEventModel{ID: v.ID, IncidentID: v.IncidentID, TenantID: v.TenantID, ActorID: v.ActorID, EventType: v.EventType, FromStatus: string(v.FromStatus), ToStatus: string(v.ToStatus), Content: v.Content, AttachmentRef: v.AttachmentRef, CreatedAt: v.CreatedAt}
}

type SimulatedPaymentOrderModel struct {
	ID             string `gorm:"primaryKey;size:128"`
	TenantID       int64  `gorm:"index"`
	Status         string `gorm:"size:32"`
	AmountMinor    int64
	Currency       string `gorm:"size:8"`
	PaymentChannel string `gorm:"size:64"`
	UpdatedAt      time.Time
}

func (SimulatedPaymentOrderModel) TableName() string { return "simulated_payment_orders" }
func (m *SimulatedPaymentOrderModel) ToDomain() *incident.SimulatedOrder {
	return &incident.SimulatedOrder{ID: m.ID, TenantID: m.TenantID, Status: m.Status, AmountMinor: m.AmountMinor, Currency: m.Currency, PaymentChannel: m.PaymentChannel, UpdatedAt: m.UpdatedAt}
}

type SimulatedPaymentRecordModel struct {
	ID              string `gorm:"primaryKey;size:64"`
	TenantID        int64  `gorm:"index"`
	OrderID         string `gorm:"index;size:128"`
	ProviderTradeNo string `gorm:"size:128"`
	Status          string `gorm:"size:32"`
	AmountMinor     int64
	OccurredAt      time.Time
}

func (SimulatedPaymentRecordModel) TableName() string { return "simulated_payment_records" }
func (m *SimulatedPaymentRecordModel) ToDomain() *incident.SimulatedPayment {
	return &incident.SimulatedPayment{ID: m.ID, TenantID: m.TenantID, OrderID: m.OrderID, ProviderTradeNo: m.ProviderTradeNo, Status: m.Status, AmountMinor: m.AmountMinor, OccurredAt: m.OccurredAt}
}

type SimulatedCallbackLogModel struct {
	ID             string `gorm:"primaryKey;size:64"`
	TenantID       int64  `gorm:"index"`
	OrderID        string `gorm:"index;size:128"`
	PaymentID      string `gorm:"size:64"`
	EventType      string `gorm:"size:64"`
	DeliveryStatus string `gorm:"size:32"`
	PayloadSummary string `gorm:"type:text"`
	ReceivedAt     time.Time
}

func (SimulatedCallbackLogModel) TableName() string { return "simulated_payment_callback_logs" }
func (m *SimulatedCallbackLogModel) ToDomain() *incident.SimulatedCallbackLog {
	return &incident.SimulatedCallbackLog{ID: m.ID, TenantID: m.TenantID, OrderID: m.OrderID, PaymentID: m.PaymentID, EventType: m.EventType, DeliveryStatus: m.DeliveryStatus, PayloadSummary: m.PayloadSummary, ReceivedAt: m.ReceivedAt}
}

type PaymentDispositionDraftModel struct {
	ID              string `gorm:"primaryKey;size:64"`
	TenantID        int64  `gorm:"index"`
	IncidentID      string `gorm:"index;size:64"`
	CreatedBy       int64
	ApprovedBy      *int64
	ActionType      string `gorm:"size:64"`
	Status          string `gorm:"size:32"`
	IdempotencyKey  string `gorm:"size:128"`
	ExecutionResult string `gorm:"type:text"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (PaymentDispositionDraftModel) TableName() string { return "payment_disposition_drafts" }
func (m *PaymentDispositionDraftModel) ToDomain() *incident.DispositionDraft {
	return &incident.DispositionDraft{ID: m.ID, TenantID: m.TenantID, IncidentID: m.IncidentID, CreatedBy: m.CreatedBy, ApprovedBy: m.ApprovedBy, ActionType: m.ActionType, Status: m.Status, IdempotencyKey: m.IdempotencyKey, ExecutionResult: m.ExecutionResult, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}
func dispositionDraftFromDomain(v *incident.DispositionDraft) *PaymentDispositionDraftModel {
	return &PaymentDispositionDraftModel{ID: v.ID, TenantID: v.TenantID, IncidentID: v.IncidentID, CreatedBy: v.CreatedBy, ApprovedBy: v.ApprovedBy, ActionType: v.ActionType, Status: v.Status, IdempotencyKey: v.IdempotencyKey, ExecutionResult: v.ExecutionResult, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
