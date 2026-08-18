package risk

import (
	"math"
	"testing"

	"mc-option/internal/engine"
)

func TestVaRKnown(t *testing.T) {
	xs := make([]float64, 100)
	for i := range xs {
		xs[i] = float64(i + 1)
	}
	v, err := VaR(xs, 0.95)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(v-95) > 1e-9 {
		t.Fatalf("VaR95 = %v, want 95", v)
	}
	es, err := ES(xs, 0.95)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(es-3) > 1e-9 {
		t.Fatalf("ES95 = %v, want 3", es)
	}
}

func TestVaRErrors(t *testing.T) {
	bad := []struct {
		xs   []float64
		conf float64
	}{
		{nil, 0.95},
		{[]float64{1}, 0.95},
		{[]float64{1, 2, 3}, 0},
		{[]float64{1, 2, 3}, 1},
		{[]float64{1, 2, 3}, -0.1},
		{[]float64{1, 2, 3}, 1.5},
	}
	for i, tc := range bad {
		if _, err := VaR(tc.xs, tc.conf); err == nil {
			t.Errorf("VaR case %d: expected error, got nil", i)
		}
		if _, err := ES(tc.xs, tc.conf); err == nil {
			t.Errorf("ES case %d: expected error, got nil", i)
		}
	}
}

func TestCI(t *testing.T) {
	lo, hi := CI(10, 1)
	if math.Abs(lo-(10-1.96)) > 1e-12 {
		t.Fatalf("lo = %v, want %v", lo, 10-1.96)
	}
	if math.Abs(hi-(10+1.96)) > 1e-12 {
		t.Fatalf("hi = %v, want %v", hi, 10+1.96)
	}
}

func TestPnLSeries(t *testing.T) {
	p := engine.Params{Spot: 100, Vol: 0.2, Rate: 0.05, Strike: 105, Maturity: 1, Steps: 8, Paths: 500, Seed: 42}
	xs, err := PnLSeries(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(xs) != p.Paths {
		t.Fatalf("len(xs) = %d, want %d", len(xs), p.Paths)
	}
	for i, x := range xs {
		if x < 0 {
			t.Fatalf("xs[%d] = %v, want >= 0", i, x)
		}
	}
	ys, _ := PnLSeries(p)
	for i := range xs {
		if xs[i] != ys[i] {
			t.Fatalf("series not deterministic at %d", i)
		}
	}

	bad := p
	bad.Spot = -1
	if _, err := PnLSeries(bad); err == nil {
		t.Fatal("expected error for invalid params")
	}
}
