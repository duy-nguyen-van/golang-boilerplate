package payment

import (
	"context"
	"fmt"

	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/constants"
	"golang-boilerplate/internal/errors"
	"golang-boilerplate/internal/models"

	"github.com/stripe/stripe-go/v82"
)

// PaymentAdapter defines the interface for payment operations
type PaymentAdapter interface {
	CreateCheckoutSession(ctx context.Context, priceID string, user models.User, mode stripe.CheckoutSessionMode) (*stripe.CheckoutSession, error)
	GetCheckoutSession(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error)
	CreateCustomerPortalSession(ctx context.Context, customerID string) (*stripe.BillingPortalSession, error)
	HandleWebhook(ctx context.Context, payload []byte, signature string) (stripe.Event, error)
	CreateCustomer(ctx context.Context, email string, userID string) (*stripe.Customer, error)
}

func ProvidePaymentAdapter(config *config.Config) (PaymentAdapter, error) {
	switch config.PaymentProvider {
	case constants.PaymentProviderStripe:
		stripeAdapter, err := NewStripeAdapter(config)
		if err != nil {
			return nil, errors.ExternalServiceError("Failed to initialize Stripe payment adapter", err).
				WithOperation("initialize_payment_adapter").
				WithResource("payment")
		}
		return stripeAdapter, nil
	default:
		return nil, errors.InternalError("Invalid payment provider", fmt.Errorf("invalid payment provider: %s", config.PaymentProvider)).
			WithOperation("initialize_payment_adapter").
			WithResource("payment").
			WithContext("payment_provider", config.PaymentProvider)
	}
}
