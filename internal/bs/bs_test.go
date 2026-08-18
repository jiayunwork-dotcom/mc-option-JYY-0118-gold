package bs

import (
	"math"
	"testing"
)

func TestCallPut(t *testing.T) {
	// Known BS values: S=100, K=100, r=0.05, vol=0.2, T=1
	// Call ≈ 10.45, Put ≈ 5.57
	c, err := Call(100, 100, 0.05, 0.2, 1)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if math.Abs(c.Price-10.45) > 0.1 {
		t.Fatalf("call price = %f, want ~10.45", c.Price)
	}
	p, err := Put(100, 100, 0.05, 0.2, 1)
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	if math.Abs(p.Price-5.57) > 0.1 {
		t.Fatalf("put price = %f, want ~5.57", p.Price)
	}
}

func TestPutCallParity(t *testing.T) {
	c, _ := Call(100, 100, 0.05, 0.2, 1)
	p, _ := Put(100, 100, 0.05, 0.2, 1)
	diff := PutCallParity(c.Price, p.Price, 100, 100, 0.05, 1)
	if math.Abs(diff) > 1e-10 {
		t.Fatalf("put-call parity violation: %e", diff)
	}
}

func TestDelta(t *testing.T) {
	d, _ := Delta(100, 100, 0.05, 0.2, 1, true)
	if d < 0.5 || d > 0.7 {
		t.Fatalf("call delta = %f, expected ~0.6", d)
	}
	dp, _ := Delta(100, 100, 0.05, 0.2, 1, false)
	if dp > -0.3 || dp < -0.5 {
		t.Fatalf("put delta = %f, expected ~-0.4", dp)
	}
}

func TestGamma(t *testing.T) {
	g, _ := Gamma(100, 100, 0.05, 0.2, 1)
	if g <= 0 {
		t.Fatalf("gamma should be positive, got %f", g)
	}
}

func TestVega(t *testing.T) {
	v, _ := Vega(100, 100, 0.05, 0.2, 1)
	if v <= 0 {
		t.Fatalf("vega should be positive, got %f", v)
	}
}

func TestTheta(t *testing.T) {
	tc, _ := Theta(100, 100, 0.05, 0.2, 1, true)
	if tc >= 0 {
		t.Fatalf("call theta should be negative, got %f", tc)
	}
}

func TestRho(t *testing.T) {
	rc, _ := Rho(100, 100, 0.05, 0.2, 1, true)
	if rc <= 0 {
		t.Fatalf("call rho should be positive, got %f", rc)
	}
	rp, _ := Rho(100, 100, 0.05, 0.2, 1, false)
	if rp >= 0 {
		t.Fatalf("put rho should be negative, got %f", rp)
	}
}

func TestInvalidInput(t *testing.T) {
	_, err := Call(-1, 100, 0.05, 0.2, 1)
	if err != ErrInvalidInput {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}
