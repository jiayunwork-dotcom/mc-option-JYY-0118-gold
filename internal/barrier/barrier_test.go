package barrier

import (
	"testing"

	"mc-option/internal/engine"
)

func baseParams() Params {
	return Params{
		Params:      engine.Params{Spot: 100, Vol: 0.2, Rate: 0.05, Strike: 100, Maturity: 1, Steps: 32, Paths: 5000, Seed: 42},
		Barrier:     120,
		BarrierType: UpAndOut,
		IsCall:      true,
	}
}

func TestUpAndOutCallCheaper(t *testing.T) {
	p := baseParams()
	bpr, err := Price(p)
	if err != nil {
		t.Fatalf("price: %v", err)
	}
	// Vanilla European should be more expensive than up-and-out.
	vpr, _ := engine.European(p.Params, true)
	if bpr.Value >= vpr.Value {
		t.Fatalf("barrier(%f) should be < vanilla(%f)", bpr.Value, vpr.Value)
	}
}

func TestDownAndOutPut(t *testing.T) {
	p := baseParams()
	p.Barrier = 80
	p.BarrierType = DownAndOut
	p.IsCall = false
	bpr, err := Price(p)
	if err != nil {
		t.Fatalf("price: %v", err)
	}
	if bpr.Value < 0 {
		t.Fatalf("price should be >= 0, got %f", bpr.Value)
	}
}

func TestUpAndInCall(t *testing.T) {
	p := baseParams()
	p.BarrierType = UpAndIn
	bpr, err := Price(p)
	if err != nil {
		t.Fatalf("price: %v", err)
	}
	if bpr.Value < 0 {
		t.Fatalf("price should be >= 0, got %f", bpr.Value)
	}
}

func TestValidateBarrier(t *testing.T) {
	p := baseParams()
	p.Barrier = -1
	if err := Validate(p); err != ErrBarrier {
		t.Fatalf("expected ErrBarrier, got %v", err)
	}
	p.Barrier = 90 // up barrier must be > spot=100
	p.BarrierType = UpAndOut
	if err := Validate(p); err == nil {
		t.Fatal("expected error for up barrier <= spot")
	}
}

func TestBarrierTypeString(t *testing.T) {
	if UpAndOut.String() != "up-and-out" {
		t.Fatalf("got %q", UpAndOut.String())
	}
}
