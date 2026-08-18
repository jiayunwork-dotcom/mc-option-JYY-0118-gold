package payoff

import (
	"math"
	"testing"
)

func TestVanillaCall(t *testing.T) {
	p := VanillaCall{Strike: 100}
	prices := []float64{95, 100, 110}
	if got := p.Compute(prices); math.Abs(got-10) > 1e-9 {
		t.Fatalf("call = %f, want 10", got)
	}
	if (VanillaCall{Strike: 120}).Compute(prices) != 0 {
		t.Fatal("OTM call should be 0")
	}
}

func TestVanillaPut(t *testing.T) {
	p := VanillaPut{Strike: 100}
	prices := []float64{105, 100, 90}
	if got := p.Compute(prices); math.Abs(got-10) > 1e-9 {
		t.Fatalf("put = %f, want 10", got)
	}
}

func TestAsianCall(t *testing.T) {
	prices := []float64{100, 110, 120} // avg = 110
	p := AsianCall{Strike: 105}
	if got := p.Compute(prices); math.Abs(got-5) > 1e-9 {
		t.Fatalf("asian call = %f, want 5", got)
	}
}

func TestAsianPut(t *testing.T) {
	prices := []float64{100, 110, 120} // avg = 110
	p := AsianPut{Strike: 115}
	if got := p.Compute(prices); math.Abs(got-5) > 1e-9 {
		t.Fatalf("asian put = %f, want 5", got)
	}
}

func TestLookbackCall(t *testing.T) {
	prices := []float64{100, 80, 90, 120}
	p := LookbackCall{}
	// final=120, min=80 → payoff=40
	if got := p.Compute(prices); math.Abs(got-40) > 1e-9 {
		t.Fatalf("lookback call = %f, want 40", got)
	}
}

func TestLookbackPut(t *testing.T) {
	prices := []float64{100, 130, 110, 90}
	p := LookbackPut{}
	// max=130, final=90 → payoff=40
	if got := p.Compute(prices); math.Abs(got-40) > 1e-9 {
		t.Fatalf("lookback put = %f, want 40", got)
	}
}

func TestDigitalCall(t *testing.T) {
	p := DigitalCall{Strike: 100, Amount: 1}
	if p.Compute([]float64{90, 101}) != 1 {
		t.Fatal("ITM digital call should pay amount")
	}
	if p.Compute([]float64{90, 99}) != 0 {
		t.Fatal("OTM digital call should pay 0")
	}
}

func TestStraddle(t *testing.T) {
	p := Straddle{Strike: 100}
	if got := p.Compute([]float64{90, 120}); math.Abs(got-20) > 1e-9 {
		t.Fatalf("straddle = %f, want 20", got)
	}
	if got := p.Compute([]float64{90, 80}); math.Abs(got-20) > 1e-9 {
		t.Fatalf("straddle = %f, want 20", got)
	}
}

func TestPayofferInterface(t *testing.T) {
	payers := []Payoffer{
		VanillaCall{Strike: 100},
		VanillaPut{Strike: 100},
		AsianCall{Strike: 100},
		LookbackCall{},
		DigitalCall{Strike: 100, Amount: 1},
		Straddle{Strike: 100},
	}
	prices := []float64{100, 110, 120}
	for _, p := range payers {
		name := p.Name()
		if name == "" {
			t.Fatal("empty name")
		}
		val := p.Compute(prices)
		if val < 0 {
			t.Fatalf("%s payoff = %f, should be >= 0", name, val)
		}
	}
}
