// Package sensitivity 提供期权参数敏感性分析：对某一参数做网格扫描，
// 观察价格或希腊字母如何随之变化。
package sensitivity

import (
	"errors"
	"math"

	"mc-option/internal/engine"
)

// ErrInvalidGrid 网格参数无效。
var ErrInvalidGrid = errors.New("sensitivity: invalid grid parameters")

// GridPoint 网格上一个点的结果。
type GridPoint struct {
	ParamValue float64 `json:"param_value"`
	Price      float64 `json:"price"`
	StdErr     float64 `json:"stderr"`
}

// SweepResult 参数扫描结果。
type SweepResult struct {
	ParamName string      `json:"param_name"`
	Points    []GridPoint `json:"points"`
	MinPrice  float64     `json:"min_price"`
	MaxPrice  float64     `json:"max_price"`
}

// SpotSweep 对现价做等间距扫描。
func SpotSweep(base engine.Params, isCall bool, lo, hi float64, steps int) (*SweepResult, error) {
	return sweep(base, isCall, "spot", lo, hi, steps, func(p *engine.Params, v float64) { p.Spot = v })
}

// VolSweep 对波动率做等间距扫描。
func VolSweep(base engine.Params, isCall bool, lo, hi float64, steps int) (*SweepResult, error) {
	return sweep(base, isCall, "vol", lo, hi, steps, func(p *engine.Params, v float64) { p.Vol = v })
}

// StrikeSweep 对行权价做等间距扫描。
func StrikeSweep(base engine.Params, isCall bool, lo, hi float64, steps int) (*SweepResult, error) {
	return sweep(base, isCall, "strike", lo, hi, steps, func(p *engine.Params, v float64) { p.Strike = v })
}

// MaturitySweep 对期限做等间距扫描。
func MaturitySweep(base engine.Params, isCall bool, lo, hi float64, steps int) (*SweepResult, error) {
	return sweep(base, isCall, "maturity", lo, hi, steps, func(p *engine.Params, v float64) { p.Maturity = v })
}

// RateSweep 对利率做等间距扫描。
func RateSweep(base engine.Params, isCall bool, lo, hi float64, steps int) (*SweepResult, error) {
	return sweep(base, isCall, "rate", lo, hi, steps, func(p *engine.Params, v float64) { p.Rate = v })
}

func sweep(base engine.Params, isCall bool, name string, lo, hi float64, steps int, set func(*engine.Params, float64)) (*SweepResult, error) {
	if steps < 2 || lo >= hi {
		return nil, ErrInvalidGrid
	}
	res := &SweepResult{ParamName: name, MinPrice: math.MaxFloat64}
	step := (hi - lo) / float64(steps-1)
	for i := 0; i < steps; i++ {
		v := lo + float64(i)*step
		p := base
		set(&p, v)
		pr, err := engine.European(p, isCall)
		if err != nil {
			continue
		}
		gp := GridPoint{ParamValue: v, Price: pr.Value, StdErr: pr.StdErr}
		res.Points = append(res.Points, gp)
		if pr.Value < res.MinPrice {
			res.MinPrice = pr.Value
		}
		if pr.Value > res.MaxPrice {
			res.MaxPrice = pr.Value
		}
	}
	return res, nil
}

// VolSurface 对一组 strike x maturity 组合计算隐含波动率平面的数据点。
// 这里简化为对每组参数定价，返回 price 矩阵。
type SurfacePoint struct {
	Strike   float64 `json:"strike"`
	Maturity float64 `json:"maturity"`
	Price    float64 `json:"price"`
}

// PriceSurface 计算 strike-maturity 平面上的价格网格。
func PriceSurface(base engine.Params, isCall bool, strikes []float64, maturities []float64) ([]SurfacePoint, error) {
	var points []SurfacePoint
	for _, k := range strikes {
		for _, t := range maturities {
			p := base
			p.Strike = k
			p.Maturity = t
			pr, err := engine.European(p, isCall)
			if err != nil {
				continue
			}
			points = append(points, SurfacePoint{Strike: k, Maturity: t, Price: pr.Value})
		}
	}
	return points, nil
}
