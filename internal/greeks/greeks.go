// Package greeks 使用有限差分法从蒙特卡洛定价引擎数值估算期权希腊字母。
// 这里不依赖解析公式，而是通过扰动参数后重新定价来计算偏导数。
package greeks

import (
	"errors"
	"math"

	"mc-option/internal/engine"
)

// ErrBump 扰动参数不合法。
var ErrBump = errors.New("greeks: bump size must be positive")

// Greeks 持有一组希腊字母值。
type Greeks struct {
	Delta float64
	Gamma float64
	Vega  float64
	Theta float64
	Rho   float64
}

// Config 控制有限差分参数。
type Config struct {
	SpotBump     float64 // delta/gamma 扰动比例 (默认 0.01)
	VolBump      float64 // vega 扰动绝对量 (默认 0.01)
	TimeBump     float64 // theta 扰动量（年）(默认 1/365)
	RateBump     float64 // rho 扰动绝对量 (默认 0.0001)
}

// DefaultConfig 返回默认有限差分参数。
func DefaultConfig() Config {
	return Config{
		SpotBump: 0.01,
		VolBump:  0.01,
		TimeBump: 1.0 / 365.0,
		RateBump: 0.0001,
	}
}

// Compute 使用蒙特卡洛定价引擎数值估算所有希腊字母。
// isCall=true 为 call，false 为 put; isAsian=true 为亚式。
func Compute(p engine.Params, isCall, isAsian bool, cfg Config) (*Greeks, error) {
	if cfg.SpotBump <= 0 || cfg.VolBump <= 0 || cfg.TimeBump <= 0 || cfg.RateBump <= 0 {
		return nil, ErrBump
	}
	base, err := price(p, isCall, isAsian)
	if err != nil {
		return nil, err
	}
	g := &Greeks{}

	// Delta: dP/dS ≈ (P(S+h) - P(S-h)) / (2h)
	dS := p.Spot * cfg.SpotBump
	pUp := bump(p, func(pp *engine.Params) { pp.Spot += dS })
	pDown := bump(p, func(pp *engine.Params) { pp.Spot -= dS })
	up, err := price(pUp, isCall, isAsian)
	if err != nil {
		return nil, err
	}
	down, err := price(pDown, isCall, isAsian)
	if err != nil {
		return nil, err
	}
	g.Delta = (up - down) / (2 * dS)

	// Gamma: d²P/dS² ≈ (P(S+h) - 2P(S) + P(S-h)) / h²
	g.Gamma = (up - 2*base + down) / (dS * dS)

	// Vega: dP/dσ
	vUp := bump(p, func(pp *engine.Params) { pp.Vol += cfg.VolBump })
	vDown := bump(p, func(pp *engine.Params) { pp.Vol -= cfg.VolBump })
	vU, err := price(vUp, isCall, isAsian)
	if err != nil {
		return nil, err
	}
	vD, err := price(vDown, isCall, isAsian)
	if err != nil {
		return nil, err
	}
	g.Vega = (vU - vD) / (2 * cfg.VolBump) / 100

	// Theta: -dP/dT
	if p.Maturity > cfg.TimeBump {
		tDown := bump(p, func(pp *engine.Params) { pp.Maturity -= cfg.TimeBump })
		tD, err := price(tDown, isCall, isAsian)
		if err == nil {
			g.Theta = -(base - tD) / cfg.TimeBump / 365
		}
	}

	// Rho: dP/dr
	rUp := bump(p, func(pp *engine.Params) { pp.Rate += cfg.RateBump })
	rDown := bump(p, func(pp *engine.Params) { pp.Rate -= cfg.RateBump })
	rU, err := price(rUp, isCall, isAsian)
	if err != nil {
		return nil, err
	}
	rD, err := price(rDown, isCall, isAsian)
	if err != nil {
		return nil, err
	}
	g.Rho = (rU - rD) / (2 * cfg.RateBump) / 100

	return g, nil
}

// bump 复制参数并应用修改。
func bump(p engine.Params, fn func(*engine.Params)) engine.Params {
	cp := p
	fn(&cp)
	return cp
}

// price 调用定价引擎返回价格值。
func price(p engine.Params, isCall, isAsian bool) (float64, error) {
	var pr engine.Price
	var err error
	if isAsian {
		pr, err = engine.Asian(p, isCall)
	} else {
		pr, err = engine.European(p, isCall)
	}
	if err != nil {
		return 0, err
	}
	return pr.Value, nil
}

// ImpliedVol 用二分法从蒙特卡洛引擎反推隐含波动率。
func ImpliedVol(p engine.Params, targetPrice float64, isCall, isAsian bool) (float64, error) {
	lo, hi := 0.01, 3.0
	for iter := 0; iter < 100; iter++ {
		mid := (lo + hi) / 2
		pp := p
		pp.Vol = mid
		pr, err := price(pp, isCall, isAsian)
		if err != nil {
			return 0, err
		}
		if math.Abs(pr-targetPrice) < 1e-6 {
			return mid, nil
		}
		if pr < targetPrice {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2, nil
}
