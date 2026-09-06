package incident

import "context"

type Repository interface {
	Create(context.Context, *PaymentIncident) error
	FindByID(context.Context, string, int64) (*PaymentIncident, error)
	List(context.Context, int64, int, int) ([]*PaymentIncident, int64, error)
	Update(context.Context, *PaymentIncident) error
}
