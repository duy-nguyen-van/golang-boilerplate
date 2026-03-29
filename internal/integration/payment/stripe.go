package payment

import (
	"context"
	"golang-boilerplate/internal/config"
	"golang-boilerplate/internal/models"

	"github.com/stripe/stripe-go/v82"
	portalsession "github.com/stripe/stripe-go/v82/billingportal/session"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	"github.com/stripe/stripe-go/v82/webhook"
)

type StripeAdapter struct {
	config *config.Config
}

// NewStripeAdapter creates a new Stripe adapter instance
func NewStripeAdapter(config *config.Config) (*StripeAdapter, error) {
	stripe.Key = config.StripeSecretKey

	return &StripeAdapter{
		config: config,
	}, nil
}

// CreateCheckoutSession creates a new Stripe Checkout session
func (a *StripeAdapter) CreateCheckoutSession(ctx context.Context, priceID string, user models.User, mode stripe.CheckoutSessionMode) (*stripe.CheckoutSession, error) { // TODO: add plan to models
	params := &stripe.CheckoutSessionParams{
		SuccessURL: stripe.String(a.config.StripeSuccessURL + "?session_id={CHECKOUT_SESSION_ID}&product_name"),
		CancelURL:  stripe.String(a.config.StripeCancelURL + "?product_name"),
		Mode:       stripe.String(string(mode)),
		Currency:   stripe.String(string(stripe.CurrencyUSD)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Customer: stripe.String(user.StripeCustomerID), // TODO: add StripeCustomerID to user model
		Metadata: map[string]string{
			"user_id":    user.ID,
			"user_email": user.Email,
			"price_id":   priceID,
		},
	}

	return session.New(params)
}

// GetCheckoutSession retrieves a Stripe Checkout session
func (a *StripeAdapter) GetCheckoutSession(ctx context.Context, sessionID string) (*stripe.CheckoutSession, error) {
	return session.Get(sessionID, nil)
}

// CreateCustomerPortalSession creates a new Customer Portal session
func (a *StripeAdapter) CreateCustomerPortalSession(ctx context.Context, customerID string) (*stripe.BillingPortalSession, error) { // TODO: add plan to models
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(a.config.StripeCustomerPortalURL + "?product_name"),
	}

	return portalsession.New(params)
}

// HandleWebhook processes incoming Stripe webhooks
func (a *StripeAdapter) HandleWebhook(ctx context.Context, payload []byte, signature string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, signature, a.config.StripeWebhookSecret)
}

// CreateCustomer creates a new Stripe customer for a user
func (a *StripeAdapter) CreateCustomer(ctx context.Context, email string, userID string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Metadata: map[string]string{
			"user_id": userID,
		},
	}

	return customer.New(params)
}
