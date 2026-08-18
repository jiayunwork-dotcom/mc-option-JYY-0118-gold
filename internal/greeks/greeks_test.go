package greeks

import (
	"math"
	"testing"

	"mc-option/internal/engine"
)

func baseParams() engine.Params {
	return engine.Params{Spot: 100, Vol: 0.2, Rate: 0.05, Strike: 100, Maturity: 1, Steps: 32, Paths: 5000, Seed: 42}
}

func TestComputeCallGreeks(t *testing.T) {
	g, err := Compute(baseParams(), true, false, DefaultConfig())
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	// Call delta should be between 0 and 1.
	if g.Delta < 0.3 || g.Delta > 0.9 {
		t.Fatalf("delta = %f, expected in [0.3, 0.9]", g.Delta)
	}
	// Gamma should be positive.
	if g.Gamma <= 0 {
		t.Fatalf("gamma = %f, should be positive", g.Gamma)
	}
	// Vega should be positive.
	if g.Vega <= 0 {
		t.Fatalf("vega = %f, should be positive", g.Vega)
	}
}

func TestComputePutGreeks(t *testing.T) {
	g, err := Compute(baseParams(), false, false, DefaultConfig())
	if err != nil {
		t.Fatalf("compute: %v", err)
	}
	// Put delta should be between -1 and 0.
	if g.Delta > 0 || g.Delta < -1 {
		t.Fatalf("put delta = %f, expected in [-1, 0]", g.Delta)
	}
}

func TestComputeInvalidBump(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SpotBump = 0
	_, err := Compute(baseParams(), true, false, cfg)
	if err != ErrBump {
		t.Fatalf("expected ErrBump, got %v", err)
	}
}

func TestImpliedVol(t *testing.T) {
	p := baseParams()
	p.Paths = 10000
	// Price with vol=0.2.
	pr, _ := engine.European(p, true)
	// Recover implied vol.
	p.Vol = 0.5 // start from wrong value.
	iv, err := ImpliedVol(p, pr.Value, true, false)
	if err != nil {
		t.Fatalf("implied vol: %v", err)
	}
	if math.Abs(iv-0.2) > 0.05 {
		t.Fatalf("implied vol = %f, want ~0.2", iv)
	}
}
