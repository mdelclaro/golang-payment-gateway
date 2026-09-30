package domain

type PaymentStatus string

const (
	PaymentAuthorized PaymentStatus = "AUTHORIZED"
	PaymentDeclined   PaymentStatus = "DECLINED"
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
