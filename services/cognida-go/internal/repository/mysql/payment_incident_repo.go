package mysql

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"cognida/internal/model/incident"
)

type paymentIncidentRepository struct{ *BaseRepository }

func NewPaymentIncidentRepository(db *gorm.DB) incident.Repository {
	return &paymentIncidentRepository{BaseRepository: NewBaseRepository(db, false)}
}

func (r *paymentIncidentRepository) Create(ctx context.Context, v *incident.PaymentIncident) error {
	return r.WithContext(ctx).Create(paymentIncidentFromDomain(v)).Error
}

func (r *paymentIncidentRepository) FindByID(ctx context.Context, id string, tenantID int64) (*incident.PaymentIncident, error) {
	var model PaymentIncidentModel
	if err := r.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("案件不存在")
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *paymentIncidentRepository) List(ctx context.Context, tenantID int64, page, size int) ([]*incident.PaymentIncident, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	db := r.WithContext(ctx).Model(&PaymentIncidentModel{}).Where("tenant_id = ?", tenantID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var models []PaymentIncidentModel
	if err := db.Order("created_at DESC").Offset((page - 1) * size).Limit(size).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*incident.PaymentIncident, len(models))
	for i := range models {
		items[i] = models[i].ToDomain()
	}
	return items, total, nil
}

func (r *paymentIncidentRepository) Update(ctx context.Context, v *incident.PaymentIncident) error {
	result := r.WithContext(ctx).Model(&PaymentIncidentModel{}).Where("id = ? AND tenant_id = ?", v.ID, v.TenantID).Updates(map[string]interface{}{"status": string(v.Status), "priority": v.Priority, "assignee_id": v.AssigneeID, "conclusion": v.Conclusion, "updated_at": v.UpdatedAt})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("案件不存在")
	}
	return nil
}

func (r *paymentIncidentRepository) CreateEvent(ctx context.Context, v *incident.TimelineEvent) error {
	return r.WithContext(ctx).Create(paymentIncidentEventFromDomain(v)).Error
}

func (r *paymentIncidentRepository) ListEvents(ctx context.Context, incidentID string, tenantID int64) ([]*incident.TimelineEvent, error) {
	var models []PaymentIncidentEventModel
	if err := r.WithContext(ctx).Where("incident_id = ? AND tenant_id = ?", incidentID, tenantID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*incident.TimelineEvent, len(models))
	for i := range models {
		items[i] = models[i].ToDomain()
	}
	return items, nil
}

func (r *paymentIncidentRepository) FindSimulatedOrder(ctx context.Context, id string, tenantID int64) (*incident.SimulatedOrder, error) {
	var model SimulatedPaymentOrderModel
	if err := r.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&model).Error; err != nil {
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *paymentIncidentRepository) ListSimulatedPayments(ctx context.Context, orderID string, tenantID int64) ([]*incident.SimulatedPayment, error) {
	var models []SimulatedPaymentRecordModel
	if err := r.WithContext(ctx).Where("order_id = ? AND tenant_id = ?", orderID, tenantID).Order("occurred_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*incident.SimulatedPayment, len(models))
	for i := range models {
		items[i] = models[i].ToDomain()
	}
	return items, nil
}

func (r *paymentIncidentRepository) ListSimulatedCallbackLogs(ctx context.Context, orderID string, tenantID int64) ([]*incident.SimulatedCallbackLog, error) {
	var models []SimulatedCallbackLogModel
	if err := r.WithContext(ctx).Where("order_id = ? AND tenant_id = ?", orderID, tenantID).Order("received_at ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*incident.SimulatedCallbackLog, len(models))
	for i := range models {
		items[i] = models[i].ToDomain()
	}
	return items, nil
}
func (r *paymentIncidentRepository) CreateDraft(ctx context.Context, v *incident.DispositionDraft) error {
	return r.WithContext(ctx).Create(dispositionDraftFromDomain(v)).Error
}
func (r *paymentIncidentRepository) FindDraft(ctx context.Context, id string, tenantID int64) (*incident.DispositionDraft, error) {
	var m PaymentDispositionDraftModel
	if err := r.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&m).Error; err != nil {
		return nil, err
	}
	return m.ToDomain(), nil
}
func (r *paymentIncidentRepository) UpdateDraft(ctx context.Context, v *incident.DispositionDraft) error {
	return r.WithContext(ctx).Model(&PaymentDispositionDraftModel{}).Where("id = ? AND tenant_id = ?", v.ID, v.TenantID).Updates(map[string]interface{}{"approved_by": v.ApprovedBy, "status": v.Status, "execution_result": v.ExecutionResult, "updated_at": v.UpdatedAt}).Error
}
