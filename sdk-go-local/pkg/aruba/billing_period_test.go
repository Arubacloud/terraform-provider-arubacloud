package aruba

import "testing"

func TestDefaultBillingPeriod(t *testing.T) {
	tests := []struct {
		name  string
		input *BillingPeriod
		want  BillingPeriod
	}{
		{name: "nil defaults to Hour", input: nil, want: BillingPeriodHour},
		{name: "explicit value echoed back", input: func() *BillingPeriod { v := BillingPeriod("Month"); return &v }(), want: "Month"},
		{name: "Hour is echoed, not re-allocated", input: func() *BillingPeriod { v := BillingPeriodHour; return &v }(), want: BillingPeriodHour},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := defaultBillingPeriod(tc.input)
			if got == nil {
				t.Fatal("returned nil")
			}
			if *got != tc.want {
				t.Errorf("got %q, want %q", *got, tc.want)
			}
		})
	}
}

var fixtureTranslator = newBillingPeriodTranslator(map[BillingPeriod]string{
	BillingPeriodHour:  "hourly",
	BillingPeriodMonth: "monthly",
})

func TestBillingPeriodTranslator_Out_KnownValues(t *testing.T) {
	cases := []struct {
		in   BillingPeriod
		want string
	}{
		{BillingPeriodHour, "hourly"},
		{BillingPeriodMonth, "monthly"},
	}
	for _, c := range cases {
		p := c.in
		got := fixtureTranslator.Out(&p)
		if got == nil || string(*got) != c.want {
			t.Errorf("Out(%q) = %v, want %q", c.in, got, c.want)
		}
	}
}

func TestBillingPeriodTranslator_In_KnownValues(t *testing.T) {
	cases := []struct {
		in   string
		want BillingPeriod
	}{
		{"hourly", BillingPeriodHour},
		{"monthly", BillingPeriodMonth},
	}
	for _, c := range cases {
		p := BillingPeriod(c.in)
		got := fixtureTranslator.In(&p)
		if got == nil || *got != c.want {
			t.Errorf("In(%q) = %v, want %q", c.in, got, c.want)
		}
	}
}

func TestBillingPeriodTranslator_Out_Nil(t *testing.T) {
	if got := fixtureTranslator.Out(nil); got != nil {
		t.Errorf("Out(nil) = %v, want nil", got)
	}
}

func TestBillingPeriodTranslator_In_Nil(t *testing.T) {
	if got := fixtureTranslator.In(nil); got != nil {
		t.Errorf("In(nil) = %v, want nil", got)
	}
}

func TestBillingPeriodTranslator_Out_UnknownPassThrough(t *testing.T) {
	unknown := BillingPeriod("Year")
	got := fixtureTranslator.Out(&unknown)
	if got == nil || *got != unknown {
		t.Errorf("Out(unknown) = %v, want pass-through %q", got, unknown)
	}
}

func TestBillingPeriodTranslator_In_UnknownPassThrough(t *testing.T) {
	unknown := BillingPeriod("yearly")
	got := fixtureTranslator.In(&unknown)
	if got == nil || *got != unknown {
		t.Errorf("In(unknown) = %v, want pass-through %q", got, unknown)
	}
}

func TestBillingPeriodTranslator_RoundTrip(t *testing.T) {
	original := BillingPeriodHour
	wire := fixtureTranslator.Out(&original)
	back := fixtureTranslator.In(wire)
	if back == nil || *back != original {
		t.Errorf("round-trip Hour: got %v, want %q", back, original)
	}
}

func TestNewBillingPeriodTranslator_DuplicateWirePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for duplicate wire value, got none")
		}
	}()
	newBillingPeriodTranslator(map[BillingPeriod]string{
		BillingPeriodHour:  "same",
		BillingPeriodMonth: "same",
	})
}
