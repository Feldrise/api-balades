package stripe

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stripe/stripe-go/v83"
)

func TestOnlyARefundStripeAnsweredWithAClientErrorIsRefused(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		refused   bool
		exceeding bool
	}{
		{"amount greater than left to refund", &stripe.Error{HTTPStatusCode: 400, Type: stripe.ErrorTypeInvalidRequest, Param: "amount"}, true, true},
		{"charge refunded already", &stripe.Error{HTTPStatusCode: 400, Code: stripe.ErrorCodeChargeAlreadyRefunded}, true, true},
		{"card error", &stripe.Error{HTTPStatusCode: 402, Type: stripe.ErrorTypeCard}, true, false},
		{"no such payment intent", &stripe.Error{HTTPStatusCode: 404, Type: stripe.ErrorTypeInvalidRequest}, true, false},
		{"idempotent request in progress", &stripe.Error{HTTPStatusCode: 409, Type: stripe.ErrorTypeIdempotency}, false, false},
		{"too many requests", &stripe.Error{HTTPStatusCode: 429}, false, false},
		{"server error", &stripe.Error{HTTPStatusCode: 500, Type: stripe.ErrorTypeAPI}, false, false},
		{"no answer in time", fmt.Errorf("request: %w", context.DeadlineExceeded), false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var refusal *RefundRefusedError
			refused := errors.As(refundError(c.err), &refusal)
			if refused != c.refused {
				t.Fatalf("expected refused %v, got %v", c.refused, refused)
			}
			if refused && refusal.ExceedsRefundable != c.exceeding {
				t.Errorf("expected exceeding the refundable %v, got %v", c.exceeding, refusal.ExceedsRefundable)
			}
		})
	}
}
