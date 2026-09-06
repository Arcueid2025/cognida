package mysql

import (
	"time"

	"cognida/internal/model/incident"
)

type PaymentIncidentModel struct {
	ID                  string `gorm:"primaryKey;size:64"`
	TenantID            int64  `gorm:"index"`
	CreatedBy           int64
	OrderID             string `gorm:"size:128"`
	Channel             string `gorm:"size:64"`
	IncidentType        string `gorm:"size:64"`
	Priority            string `gorm:"size:8"`
	Status              string `gorm:"size:32"`
	AssigneeID          *int64
	AssessmentRequestID string `gorm:"size:64"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (PaymentIncidentModel) TableName() string { return "payment_incidents" }

func (m *PaymentIncidentModel) ToDomain() *incident.PaymentIncident {
	return &incident.PaymentIncident{ID: m.ID, TenantID: m.TenantID, CreatedBy: m.CreatedBy, OrderID: m.OrderID, Channel: m.Channel, IncidentType: m.IncidentType, Priority: m.Priority, Status: incident.Status(m.Status), AssigneeID: m.AssigneeID, AssessmentRequestID: m.AssessmentRequestID, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt}
}

func paymentIncidentFromDomain(v *incident.PaymentIncident) *PaymentIncidentModel {
	return &PaymentIncidentModel{ID: v.ID, TenantID: v.TenantID, CreatedBy: v.CreatedBy, OrderID: v.OrderID, Channel: v.Channel, IncidentType: v.IncidentType, Priority: v.Priority, Status: string(v.Status), AssigneeID: v.AssigneeID, AssessmentRequestID: v.AssessmentRequestID, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
}
