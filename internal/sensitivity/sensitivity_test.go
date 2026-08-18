package sensitivity

import (
	"testing"

	"mc-option/internal/engine"
)

func baseParams() engine.Params {
	return engine.Params{Spot: 100, Vol: 0.2, Rate: 0.05, Strike: 100, Maturity: 1, Steps: 16, Paths: 500, Seed: 42}
}

func TestSpotSweep(t *testing.T) {
	res, err := SpotSweep(baseParams(), true, 80, 120, 5)
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if len(res.Points) != 5 {
		t.Fatalf("points = %d, want 5", len(res.Points))
	}
	// Higher spot → higher call price.
	if res.Points[0].Price >= res.Points[4].Price {
		t.Fatal("call price should increase with spot")
	}
}

func TestVolSweep(t *testing.T) {
	res, err := VolSweep(baseParams(), true, 0.1, 0.5, 4)
	if err != nil {
		t.Fatalf("sweep: %v", err)
	}
	if len(res.Points) < 4 {
		t.Fatalf("points = %d", len(res.Points))
	}
}

func TestInvalidGrid(t *testing.T) {
	_, err := SpotSweep(baseParams(), true, 100, 80, 5)
	if err != ErrInvalidGrid {
		t.Fatalf("expected ErrInvalidGrid, got %v", err)
	}
	_, err = SpotSweep(baseParams(), true, 80, 120, 1)
	if err != ErrInvalidGrid {
		t.Fatalf("expected ErrInvalidGrid for steps=1, got %v", err)
	}
}

func TestPriceSurface(t *testing.T) {
	strikes := []float64{95, 100, 105}
	mats := []float64{0.5, 1.0}
	pts, err := PriceSurface(baseParams(), true, strikes, mats)
	if err != nil {
		t.Fatalf("surface: %v", err)
	}
	if len(pts) != 6 {
		t.Fatalf("points = %d, want 6", len(pts))
	}
}
