// digital.go 扩展 barrier 包，增加数字障碍期权和双障碍期权。
package barrier

import (
	"math"

	"mc-option/internal/engine"
	"mc-option/internal/rng"
)

// DigitalBarrierParams 数字障碍期权参数。
type DigitalBarrierParams struct {
	Params
	Payout float64 // 固定收益金额
}

// PriceDigital 定价数字障碍期权：触碰障碍时支付固定金额。
func PriceDigital(p DigitalBarrierParams) (engine.Price, error) {
	if err := Validate(p.Params); err != nil {
		return engine.Price{}, err
	}
	if p.Payout <= 0 {
		return engine.Price{}, ErrBarrier
	}
	n := rng.NewNormal(p.Seed)
	dt := p.Maturity / float64(p.Steps)
	disc := math.Exp(-p.Rate * p.Maturity)
	zs := make([]float64, p.Steps)
	payoffs := make([]float64, 0, p.Paths)

	for i := 0; i < p.Paths; i++ {
		for j := range zs {
			zs[j] = n.Next()
		}
		path, err := rng.Path(p.Spot, p.Rate, p.Vol, dt, zs)
		if err != nil {
			return engine.Price{}, err
		}
		touched := barrierTouched(path, p.Barrier, p.BarrierType)
		var payout float64
		switch p.BarrierType {
		case UpAndIn, DownAndIn:
			if touched {
				payout = p.Payout
			}
		case UpAndOut, DownAndOut:
			if !touched {
				payout = p.Payout
			}
		}
		payoffs = append(payoffs, disc*payout)
	}
	mean, sd := sampleStats(payoffs)
	return engine.Price{Value: mean, StdErr: sd / math.Sqrt(float64(len(payoffs)))}, nil
}

// DoubleBarrierParams 双障碍期权参数。
type DoubleBarrierParams struct {
	engine.Params
	UpperBarrier float64
	LowerBarrier float64
	IsCall       bool
}

// PriceDoubleBarrier 定价双障碍敲出期权：标的触碰任一障碍则失效。
func PriceDoubleBarrier(p DoubleBarrierParams) (engine.Price, error) {
	if err := engine.Validate(p.Params); err != nil {
		return engine.Price{}, err
	}
	if p.UpperBarrier <= p.Spot || p.LowerBarrier >= p.Spot || p.LowerBarrier <= 0 {
		return engine.Price{}, ErrBarrier
	}
	n := rng.NewNormal(p.Seed)
	dt := p.Maturity / float64(p.Steps)
	disc := math.Exp(-p.Rate * p.Maturity)
	zs := make([]float64, p.Steps)
	payoffs := make([]float64, 0, p.Paths)

	for i := 0; i < p.Paths; i++ {
		for j := range zs {
			zs[j] = n.Next()
		}
		path, err := rng.Path(p.Spot, p.Rate, p.Vol, dt, zs)
		if err != nil {
			return engine.Price{}, err
		}
		knocked := false
		for _, s := range path {
			if s >= p.UpperBarrier || s <= p.LowerBarrier {
				knocked = true
				break
			}
		}
		var pf float64
		if !knocked {
			final := path[len(path)-1]
			if p.IsCall {
				pf = math.Max(final-p.Strike, 0)
			} else {
				pf = math.Max(p.Strike-final, 0)
			}
		}
		payoffs = append(payoffs, disc*pf)
	}
	mean, sd := sampleStats(payoffs)
	return engine.Price{Value: mean, StdErr: sd / math.Sqrt(float64(len(payoffs)))}, nil
}
