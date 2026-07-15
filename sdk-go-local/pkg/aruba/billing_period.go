package aruba

import "fmt"

// defaultBillingPeriod returns p when set, otherwise a pointer to the
// platform's default billing period (Hour). Centralising this here
// keeps every wrapper's toRequest() in sync with the API default.
func defaultBillingPeriod(p *BillingPeriod) *BillingPeriod {
	if p != nil {
		return p
	}
	v := BillingPeriodHour
	return &v
}

// billingPeriodTranslator round-trips BillingPeriod values between the
// standard SDK constants (e.g. BillingPeriodMonth = "Month") and a
// resource-specific wire form (e.g. ElasticIP "monthly", StorageBackup
// "Monthly"). Unknown values pass through unchanged in both directions
// so the SDK stays forward-compatible with periods the API may add
// before the translator catches up.
//
// One instance per resource, defined as a private function in each resource
// file that needs it. Resources whose API already accepts the standard form
// do not need a translator at all — keep using *BillingPeriod /
// defaultBillingPeriod directly.
type billingPeriodTranslator struct {
	toWire   map[BillingPeriod]string
	fromWire map[string]BillingPeriod
}

// newBillingPeriodTranslator builds a translator from a SDK→wire map and
// derives the inverse. Panics if two SDK constants map to the same wire
// value — that would silently corrupt fromResponse hydration.
func newBillingPeriodTranslator(toWire map[BillingPeriod]string) *billingPeriodTranslator {
	fromWire := make(map[string]BillingPeriod, len(toWire))
	for std, w := range toWire {
		if existing, dup := fromWire[w]; dup {
			panic(fmt.Sprintf(
				"billingPeriodTranslator: duplicate wire value %q (mapped from both %q and %q)",
				w, existing, std,
			))
		}
		fromWire[w] = std
	}
	return &billingPeriodTranslator{toWire: toWire, fromWire: fromWire}
}

// Out applies the SDK→wire translation. Nil-safe. Unknown values pass through.
func (t *billingPeriodTranslator) Out(p *BillingPeriod) *BillingPeriod {
	if p == nil {
		return nil
	}
	if w, ok := t.toWire[*p]; ok {
		v := BillingPeriod(w)
		return &v
	}
	return p
}

// In applies the wire→SDK translation. Nil-safe. Unknown values pass through.
func (t *billingPeriodTranslator) In(p *BillingPeriod) *BillingPeriod {
	if p == nil {
		return nil
	}
	if std, ok := t.fromWire[string(*p)]; ok {
		v := std
		return &v
	}
	return p
}
