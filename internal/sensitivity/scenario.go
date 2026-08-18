// scenario.go 提供预定义的压力测试场景和场景分析工具。
package sensitivity

import "mc-option/internal/engine"

// Scenario 描述一个压力测试场景。
type Scenario struct {
	Name      string  `json:"name"`
	SpotPct   float64 `json:"spot_pct"`   // 标的价格变化百分比
	VolAbs    float64 `json:"vol_abs"`    // 波动率绝对变化
	RateAbs   float64 `json:"rate_abs"`   // 利率绝对变化
}

// ScenarioResult 场景分析结果。
type ScenarioResult struct {
	Scenario Scenario `json:"scenario"`
	BasePrice  float64 `json:"base_price"`
	StressPrice float64 `json:"stress_price"`
	PnL         float64 `json:"pnl"`
}

// StandardScenarios 返回一组标准压力测试场景。
func StandardScenarios() []Scenario {
	return []Scenario{
		{Name: "market-crash", SpotPct: -0.20, VolAbs: 0.15, RateAbs: -0.01},
		{Name: "vol-spike", SpotPct: -0.05, VolAbs: 0.20, RateAbs: 0},
		{Name: "rally", SpotPct: 0.15, VolAbs: -0.05, RateAbs: 0.005},
		{Name: "rate-hike", SpotPct: 0, VolAbs: 0, RateAbs: 0.02},
		{Name: "quiet-market", SpotPct: 0, VolAbs: -0.10, RateAbs: 0},
	}
}

// RunScenarios 对一组场景运行压力测试。
func RunScenarios(base engine.Params, isCall bool, scenarios []Scenario) ([]ScenarioResult, error) {
	basePr, err := engine.European(base, isCall)
	if err != nil {
		return nil, err
	}
	var results []ScenarioResult
	for _, sc := range scenarios {
		p := base
		p.Spot *= (1 + sc.SpotPct)
		p.Vol += sc.VolAbs
		if p.Vol <= 0 {
			p.Vol = 0.01
		}
		p.Rate += sc.RateAbs
		if p.Rate <= 0 {
			p.Rate = 0.001
		}
		pr, err := engine.European(p, isCall)
		if err != nil {
			continue
		}
		results = append(results, ScenarioResult{
			Scenario:    sc,
			BasePrice:   basePr.Value,
			StressPrice: pr.Value,
			PnL:         pr.Value - basePr.Value,
		})
	}
	return results, nil
}

// WorstCase 找到最差（PnL 最低）的场景。
func WorstCase(results []ScenarioResult) *ScenarioResult {
	if len(results) == 0 {
		return nil
	}
	worst := &results[0]
	for i := 1; i < len(results); i++ {
		if results[i].PnL < worst.PnL {
			worst = &results[i]
		}
	}
	return worst
}

// BestCase 找到最优（PnL 最高）的场景。
func BestCase(results []ScenarioResult) *ScenarioResult {
	if len(results) == 0 {
		return nil
	}
	best := &results[0]
	for i := 1; i < len(results); i++ {
		if results[i].PnL > best.PnL {
			best = &results[i]
		}
	}
	return best
}
