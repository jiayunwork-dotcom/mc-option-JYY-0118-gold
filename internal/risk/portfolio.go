// portfolio.go 提供组合风险分析：多头寸的联合 VaR、组合希腊字母、
// 以及场景分析。
package risk

import (
	"math"
	"sort"
)

// Position 描述一个期权仓位。
type Position struct {
	Name     string
	Quantity float64 // 正为多头，负为空头
	PnL      []float64 // 该仓位的损益序列
}

// Portfolio 持有一组仓位。
type Portfolio struct {
	Positions []Position
}

// NewPortfolio 创建空组合。
func NewPortfolio() *Portfolio {
	return &Portfolio{}
}

// Add 添加一个仓位。
func (pf *Portfolio) Add(pos Position) {
	pf.Positions = append(pf.Positions, pos)
}

// CombinedPnL 将所有仓位的 PnL 按数量加权求和，返回组合级别损益序列。
// 要求所有仓位 PnL 序列等长。
func (pf *Portfolio) CombinedPnL() []float64 {
	if len(pf.Positions) == 0 {
		return nil
	}
	n := len(pf.Positions[0].PnL)
	combined := make([]float64, n)
	for _, pos := range pf.Positions {
		for i := 0; i < n && i < len(pos.PnL); i++ {
			combined[i] += pos.Quantity * pos.PnL[i]
		}
	}
	return combined
}

// PortfolioVaR 计算组合的 VaR。
func (pf *Portfolio) PortfolioVaR(conf float64) (float64, error) {
	return VaR(pf.CombinedPnL(), conf)
}

// PortfolioES 计算组合的 ES。
func (pf *Portfolio) PortfolioES(conf float64) (float64, error) {
	return ES(pf.CombinedPnL(), conf)
}

// DiversificationBenefit 计算分散化收益：各仓位独立 VaR 之和 - 组合 VaR。
func (pf *Portfolio) DiversificationBenefit(conf float64) (float64, error) {
	sumIndividual := 0.0
	for _, pos := range pf.Positions {
		scaled := make([]float64, len(pos.PnL))
		for i, v := range pos.PnL {
			scaled[i] = pos.Quantity * v
		}
		v, err := VaR(scaled, conf)
		if err != nil {
			return 0, err
		}
		sumIndividual += v
	}
	combined, err := pf.PortfolioVaR(conf)
	if err != nil {
		return 0, err
	}
	return sumIndividual - combined, nil
}

// Scenario 描述一个压力测试场景。
type Scenario struct {
	Name      string
	SpotShift float64 // 标的价格变化百分比
	VolShift  float64 // 波动率绝对变化
}

// ScenarioResult 场景分析结果。
type ScenarioResult struct {
	ScenarioName string  `json:"scenario"`
	PnLChange    float64 `json:"pnl_change"`
}

// Percentiles 返回损益分布的多个分位数。
func Percentiles(xs []float64, pcts []float64) []float64 {
	sorted := append([]float64(nil), xs...)
	sort.Float64s(sorted)
	out := make([]float64, len(pcts))
	n := float64(len(sorted))
	for i, p := range pcts {
		idx := p / 100 * (n - 1)
		lo := int(math.Floor(idx))
		hi := int(math.Ceil(idx))
		if lo < 0 {
			lo = 0
		}
		if hi >= len(sorted) {
			hi = len(sorted) - 1
		}
		frac := idx - float64(lo)
		out[i] = sorted[lo]*(1-frac) + sorted[hi]*frac
	}
	return out
}

// MaxDrawdown 计算最大回撤。
func MaxDrawdown(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	peak := xs[0]
	maxDD := 0.0
	for _, x := range xs {
		if x > peak {
			peak = x
		}
		dd := peak - x
		if dd > maxDD {
			maxDD = dd
		}
	}
	return maxDD
}

// SharpeRatio 计算（简化）Sharpe 比率：mean(excess)/std(excess)。
func SharpeRatio(pnl []float64, riskFreeReturn float64) float64 {
	if len(pnl) < 2 {
		return 0
	}
	sum := 0.0
	for _, x := range pnl {
		sum += x - riskFreeReturn
	}
	mean := sum / float64(len(pnl))
	sumSq := 0.0
	for _, x := range pnl {
		d := (x - riskFreeReturn) - mean
		sumSq += d * d
	}
	std := math.Sqrt(sumSq / float64(len(pnl)-1))
	if std == 0 {
		return 0
	}
	return mean / std
}
