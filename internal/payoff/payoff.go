// Package payoff 提供期权收益结构的抽象接口和多种支付类型实现。
// 通过 Payoffer 接口可以灵活组合不同的收益计算逻辑。
package payoff

import "math"

// Payoffer 收益计算接口。
type Payoffer interface {
	// Compute 给定价格路径计算未折现收益。
	Compute(prices []float64) float64
	// Name 返回收益类型名。
	Name() string
}

// VanillaCall 欧式看涨期权收益。
type VanillaCall struct {
	Strike float64
}

// Compute 实现 Payoffer。
func (v VanillaCall) Compute(prices []float64) float64 {
	return math.Max(prices[len(prices)-1]-v.Strike, 0)
}

// Name 实现 Payoffer。
func (v VanillaCall) Name() string { return "vanilla-call" }

// VanillaPut 欧式看跌期权收益。
type VanillaPut struct {
	Strike float64
}

// Compute 实现 Payoffer。
func (v VanillaPut) Compute(prices []float64) float64 {
	return math.Max(v.Strike-prices[len(prices)-1], 0)
}

// Name 实现 Payoffer。
func (v VanillaPut) Name() string { return "vanilla-put" }

// AsianCall 算术平均价亚式看涨。
type AsianCall struct {
	Strike float64
}

// Compute 计算算术平均后的 call 收益。
func (a AsianCall) Compute(prices []float64) float64 {
	avg := arithmeticMean(prices)
	return math.Max(avg-a.Strike, 0)
}

// Name 实现 Payoffer。
func (a AsianCall) Name() string { return "asian-call" }

// AsianPut 算术平均价亚式看跌。
type AsianPut struct {
	Strike float64
}

// Compute 计算算术平均后的 put 收益。
func (a AsianPut) Compute(prices []float64) float64 {
	avg := arithmeticMean(prices)
	return math.Max(a.Strike-avg, 0)
}

// Name 实现 Payoffer。
func (a AsianPut) Name() string { return "asian-put" }

// LookbackCall 回望期权，收益为 max(S_final - S_min, 0)。
type LookbackCall struct{}

// Compute 实现 Payoffer。
func (l LookbackCall) Compute(prices []float64) float64 {
	min := prices[0]
	for _, s := range prices[1:] {
		if s < min {
			min = s
		}
	}
	return math.Max(prices[len(prices)-1]-min, 0)
}

// Name 实现 Payoffer。
func (l LookbackCall) Name() string { return "lookback-call" }

// LookbackPut 回望期权，收益为 max(S_max - S_final, 0)。
type LookbackPut struct{}

// Compute 实现 Payoffer。
func (l LookbackPut) Compute(prices []float64) float64 {
	max := prices[0]
	for _, s := range prices[1:] {
		if s > max {
			max = s
		}
	}
	return math.Max(max-prices[len(prices)-1], 0)
}

// Name 实现 Payoffer。
func (l LookbackPut) Name() string { return "lookback-put" }

// DigitalCall 数字期权：若到期价 > Strike 则收益为固定 Amount，否则 0。
type DigitalCall struct {
	Strike float64
	Amount float64
}

// Compute 实现 Payoffer。
func (d DigitalCall) Compute(prices []float64) float64 {
	if prices[len(prices)-1] > d.Strike {
		return d.Amount
	}
	return 0
}

// Name 实现 Payoffer。
func (d DigitalCall) Name() string { return "digital-call" }

// DigitalPut 数字看跌：若到期价 < Strike 则收益为固定 Amount。
type DigitalPut struct {
	Strike float64
	Amount float64
}

// Compute 实现 Payoffer。
func (d DigitalPut) Compute(prices []float64) float64 {
	if prices[len(prices)-1] < d.Strike {
		return d.Amount
	}
	return 0
}

// Name 实现 Payoffer。
func (d DigitalPut) Name() string { return "digital-put" }

// Straddle 跨式组合：同时买入同行权价的 call 和 put。
type Straddle struct {
	Strike float64
}

// Compute 计算跨式收益。
func (s Straddle) Compute(prices []float64) float64 {
	final := prices[len(prices)-1]
	return math.Abs(final - s.Strike)
}

// Name 实现 Payoffer。
func (s Straddle) Name() string { return "straddle" }

// arithmeticMean 计算平均值。
func arithmeticMean(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sum := 0.0
	for _, x := range xs {
		sum += x
	}
	return sum / float64(len(xs))
}
