package initialize

import "testing"

func TestPaymentSeedKey(t *testing.T) {
	t.Parallel()
	if got := paymentSeedKey(`{"payment_key":"offline"}`); got != "offline" {
		t.Fatalf("got %q", got)
	}
	if got := paymentSeedKey(""); got != "" {
		t.Fatalf("empty: got %q", got)
	}
	if got := paymentSeedKey(`{"foo":1}`); got != "" {
		t.Fatalf("no key: got %q", got)
	}
}
