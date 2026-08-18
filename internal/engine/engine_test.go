package engine

import (
	"math"
	"strings"
	"testing"
)

func baseParams() Params {
	return Params{Spot: 100, Vol: 0.2, Rate: 0.05, Strike: 105, Maturity: 1, Steps: 64, Paths: 1000, Seed: 42}
}

func TestValidateErrors(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Params)
		want   string
	}{
		{"zero spot", func(p *Params) { p.Spot = 0 }, "spot"},
		{"negative spot", func(p *Params) { p.Spot = -5 }, "spot"},
		{"zero vol", func(p *Params) { p.Vol = 0 }, "vol"},
		{"negative vol", func(p *Params) { p.Vol = -0.2 }, "vol"},
		{"zero rate", func(p *Params) { p.Rate = 0 }, "rate"},
		{"negative rate", func(p *Params) { p.Rate = -0.05 }, "rate"},
		{"zero strike", func(p *Params) { p.Strike = 0 }, "strike"},
		{"negative strike", func(p *Params) { p.Strike = -1 }, "strike"},
		{"zero maturity", func(p *Params) { p.Maturity = 0 }, "maturity"},
		{"negative maturity", func(p *Params) { p.Maturity = -1 }, "maturity"},
		{"steps 0", func(p *Params) { p.Steps = 0 }, "steps"},
		{"paths 99", func(p *Params) { p.Paths = 99 }, "paths"},
	}
	for _, tc := range cases {
		p := baseParams()
		tc.mutate(&p)
		err := Validate(p)
		if err == nil {
			t.Errorf("%s: expected error, got nil", tc.name)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q does not mention %q", tc.name, err.Error(), tc.want)
		}
	}
	if err := Validate(baseParams()); err != nil {
		t.Fatalf("valid params rejected: %v", err)
	}
}

func TestEuropeanCallZeroVol(t *testing.T) {
	p := baseParams()
	p.Vol = 0.0001
	p.Paths = 20000
	pr, err := European(p, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := math.Max(p.Spot-p.Strike*math.Exp(-p.Rate*p.Maturity), 0)
	if math.Abs(pr.Value-want) > 1e-2 {
		t.Fatalf("price = %v, want %v (diff %v)", pr.Value, want, pr.Value-want)
	}
}

func TestEuropeanInvalidParams(t *testing.T) {
	bad := baseParams()
	bad.Spot = -1
	if _, err := European(bad, true); err == nil {
		t.Fatal("European: expected error for invalid params")
	}
	if _, err := Asian(bad, true); err == nil {
		t.Fatal("Asian: expected error for invalid params")
	}
}

func TestEuropeanDeterministic(t *testing.T) {
	a, err1 := European(baseParams(), true)
	b, err2 := European(baseParams(), true)
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v %v", err1, err2)
	}
	if a.Value != b.Value || a.StdErr != b.StdErr {
		t.Fatalf("same-seed pricing differs: %v vs %v", a, b)
	}
}

func TestPayoffCallPut(t *testing.T) {
	euro := []float64{90, 110}
	if got := Payoff(euro, 105, true, false); math.Abs(got-5) > 1e-12 {
		t.Fatalf("euro call = %v, want 5", got)
	}
	if got := Payoff(euro, 105, false, false); got != 0 {
		t.Fatalf("euro put = %v, want 0", got)
	}
	asian := []float64{100, 110, 120} // 均价 110
	if got := Payoff(asian, 105, true, true); math.Abs(got-5) > 1e-12 {
		t.Fatalf("asian call = %v, want 5", got)
	}
	if got := Payoff(asian, 115, false, true); math.Abs(got-5) > 1e-12 {
		t.Fatalf("asian put = %v, want 5", got)
	}
	if got := Payoff(asian, 115, false, false); got != 0 {
		t.Fatalf("euro put on last price = %v, want 0", got)
	}
}
