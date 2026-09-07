package incident

import "context"

type Repository interface {
	Create(context.Context, *PaymentIncident) error
	FindByID(context.Context, string, int64) (*PaymentIncident, error)
	List(context.Context, int64, int, int) ([]*PaymentIncident, int64, error)
	Update(context.Context, *PaymentIncident) error
	CreateEvent(context.Context, *TimelineEvent) error
	ListEvents(context.Context, string, int64) ([]*TimelineEvent, error)
	FindSimulatedOrder(context.Context, string, int64) (*SimulatedOrder, error)
	ListSimulatedPayments(context.Context, string, int64) ([]*SimulatedPayment, error)
	ListSimulatedCallbackLogs(context.Context, string, int64) ([]*SimulatedCallbackLog, error)
	CreateDraft(context.Context, *DispositionDraft) error
	FindDraft(context.Context, string, int64) (*DispositionDraft, error)
	UpdateDraft(context.Context, *DispositionDraft) error
}
