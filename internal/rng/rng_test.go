package rng

import (
	"math"
	"testing"
)

func TestNormalDeterministic(t *testing.T) {
	a := NewNormal(42)
	b := NewNormal(42)
	for i := 0; i < 1000; i++ {
		if a.Next() != b.Next() {
			t.Fatalf("draw %d diverged between same-seed generators", i)
		}
	}
}

func TestNormalMeanStd(t *testing.T) {
	n := NewNormal(7)
	const N = 200000
	sum, sumSq := 0.0, 0.0
	for i := 0; i < N; i++ {
		z := n.Next()
		sum += z
		sumSq += z * z
	}
	mean := sum / N
	std := math.Sqrt(sumSq/N - mean*mean)
	if math.Abs(mean) > 0.01 {
		t.Fatalf("mean = %v, want ~0", mean)
	}
	if math.Abs(std-1) > 0.01 {
		t.Fatalf("std = %v, want ~1", std)
	}
}

func TestAntithetic(t *testing.T) {
	n := NewNormal(7)
	z := n.Next()
	if got := n.Antithetic(0, z); got != -z {
		t.Fatalf("Antithetic = %v, want %v", got, -z)
	}
}

func TestPathEmptyZs(t *testing.T) {
	if _, err := Path(100, 0.05, 0.2, 0.01, nil); err == nil {
		t.Fatal("expected error for empty zs, got nil")
	}
	if _, err := Path(100, 0.05, 0.2, 0.01, []float64{}); err == nil {
		t.Fatal("expected error for zero-length zs, got nil")
	}
}

func TestPathLength(t *testing.T) {
	path, err := Path(100, 0.05, 0.2, 0.01, make([]float64, 16))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(path) != 17 {
		t.Fatalf("len(path) = %d, want 17", len(path))
	}
	if path[0] != 100 {
		t.Fatalf("path[0] = %v, want 100", path[0])
	}
}

func TestGBMExactStep(t *testing.T) {
	spot, drift, vol, dt, z := 100.0, 0.05, 0.2, 0.25, 1.5
	path, err := Path(spot, drift, vol, dt, []float64{z})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := spot * math.Exp((drift-vol*vol/2)*dt+vol*math.Sqrt(dt)*z)
	if math.Abs(path[1]-want) > 1e-12 {
		t.Fatalf("path[1] = %v, want %v", path[1], want)
	}
}
