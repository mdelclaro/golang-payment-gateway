package domain

type PaymentStatus string

const (
	PaymentAuthorized PaymentStatus = "Authorized"
	PaymentDeclined   PaymentStatus = "Declined"
)

type Payment struct {
	ID                int64
	Status            PaymentStatus
	CardLastFour      string
	ExpiryMonth       int
	ExpiryYear        int
	Currency          string
	Amount            int64
	AuthorizationCode string
}
