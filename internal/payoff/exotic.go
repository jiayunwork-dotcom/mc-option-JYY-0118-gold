// exotic.go 扩展 payoff 包，增加更多奇异期权收益结构。
package payoff

import "math"

// BestOf 多资产最优期权：收益为多条路径终值中最高者减行权价。
type BestOf struct {
	Strike float64
}

// Compute 取路径中最高终值（简化：单路径取最大值）。
func (b BestOf) Compute(prices []float64) float64 {
	if len(prices) == 0 {
		return 0
	}
	return math.Max(prices[len(prices)-1]-b.Strike, 0)
}

// Name 实现 Payoffer。
func (b BestOf) Name() string { return "best-of" }

// Chooser 选择者期权：到期时持有者选择 call 还是 put，取较高者。
type Chooser struct {
	Strike float64
}

// Compute 实现选择者期权。
func (c Chooser) Compute(prices []float64) float64 {
	final := prices[len(prices)-1]
	callPay := math.Max(final-c.Strike, 0)
	putPay := math.Max(c.Strike-final, 0)
	return math.Max(callPay, putPay)
}

// Name 实现 Payoffer。
func (c Chooser) Name() string { return "chooser" }

// PowerCall 幂期权：收益为 max(S^n - K, 0)。
type PowerCall struct {
	Strike float64
	Power  float64
}

// Compute 实现幂期权。
func (p PowerCall) Compute(prices []float64) float64 {
	final := prices[len(prices)-1]
	return math.Max(math.Pow(final, p.Power)-p.Strike, 0)
}

// Name 实现 Payoffer。
func (p PowerCall) Name() string { return "power-call" }

// Cliquet 棘轮期权：累计各期的正向收益率。
type Cliquet struct {
	Floor float64 // 单期最低收益率
	Cap   float64 // 单期最高收益率
}

// Compute 计算棘轮收益。
func (cl Cliquet) Compute(prices []float64) float64 {
	if len(prices) < 2 {
		return 0
	}
	total := 0.0
	for i := 1; i < len(prices); i++ {
		ret := (prices[i] - prices[i-1]) / prices[i-1]
		if ret < cl.Floor {
			ret = cl.Floor
		}
		if ret > cl.Cap {
			ret = cl.Cap
		}
		total += ret
	}
	return math.Max(total, 0) * prices[0]
}

// Name 实现 Payoffer。
func (cl Cliquet) Name() string { return "cliquet" }

// Spread call spread: max(S - K1, 0) - max(S - K2, 0) 其中 K1 < K2。
type Spread struct {
	StrikeLow  float64
	StrikeHigh float64
}

// Compute 实现 call spread。
func (s Spread) Compute(prices []float64) float64 {
	final := prices[len(prices)-1]
	return math.Max(final-s.StrikeLow, 0) - math.Max(final-s.StrikeHigh, 0)
}

// Name 实现 Payoffer。
func (s Spread) Name() string { return "call-spread" }

// Butterfly 蝶式组合。
type Butterfly struct {
	StrikeLow  float64
	StrikeMid  float64
	StrikeHigh float64
}

// Compute 实现蝶式组合收益。
func (bf Butterfly) Compute(prices []float64) float64 {
	final := prices[len(prices)-1]
	return math.Max(final-bf.StrikeLow, 0) - 2*math.Max(final-bf.StrikeMid, 0) + math.Max(final-bf.StrikeHigh, 0)
}

// Name 实现 Payoffer。
func (bf Butterfly) Name() string { return "butterfly" }
