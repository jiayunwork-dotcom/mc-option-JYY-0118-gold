// Package barrier 实现障碍期权定价。障碍期权在标的价格触碰或突破预设障碍
// 水平时生效（knock-in）或失效（knock-out）。
package barrier

import (
	"errors"
	"math"

	"mc-option/internal/engine"
	"mc-option/internal/rng"
)

// BarrierType 障碍类型。
type BarrierType int

const (
	UpAndOut   BarrierType = iota // 上涨触碰障碍后失效
	UpAndIn                       // 上涨触碰障碍后生效
	DownAndOut                    // 下跌触碰障碍后失效
	DownAndIn                     // 下跌触碰障碍后生效
)

// String 返回障碍类型的字符串表示。
func (bt BarrierType) String() string {
	switch bt {
	case UpAndOut:
		return "up-and-out"
	case UpAndIn:
		return "up-and-in"
	case DownAndOut:
		return "down-and-out"
	case DownAndIn:
		return "down-and-in"
	default:
		return "unknown"
	}
}

// Params 障碍期权定价参数。
type Params struct {
	engine.Params
	Barrier     float64     // 障碍水平
	BarrierType BarrierType // 障碍类型
	IsCall      bool        // call/put
}

// ErrBarrier 障碍参数无效。
var ErrBarrier = errors.New("barrier: invalid barrier level")

// Validate 校验障碍参数。
func Validate(p Params) error {
	if err := engine.Validate(p.Params); err != nil {
		return err
	}
	if p.Barrier <= 0 {
		return ErrBarrier
	}
	switch p.BarrierType {
	case UpAndOut, UpAndIn:
		if p.Barrier <= p.Spot {
			return errors.New("barrier: up barrier must be > spot")
		}
	case DownAndOut, DownAndIn:
		if p.Barrier >= p.Spot {
			return errors.New("barrier: down barrier must be < spot")
		}
	}
	return nil
}

// Price 用蒙特卡洛为障碍期权定价。
func Price(p Params) (engine.Price, error) {
	if err := Validate(p); err != nil {
		return engine.Price{}, err
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
		payoff := computePayoff(path, p.Strike, p.IsCall, touched, p.BarrierType)
		payoffs = append(payoffs, disc*payoff)
	}

	mean, sd := sampleStats(payoffs)
	return engine.Price{Value: mean, StdErr: sd / math.Sqrt(float64(len(payoffs)))}, nil
}

// barrierTouched 检查路径是否触碰了障碍。
func barrierTouched(path []float64, barrier float64, bt BarrierType) bool {
	for _, s := range path {
		switch bt {
		case UpAndOut, UpAndIn:
			if s >= barrier {
				return true
			}
		case DownAndOut, DownAndIn:
			if s <= barrier {
				return true
			}
		}
	}
	return false
}

// computePayoff 根据障碍状态计算收益。
func computePayoff(path []float64, strike float64, isCall, touched bool, bt BarrierType) float64 {
	final := path[len(path)-1]
	var intrinsic float64
	if isCall {
		intrinsic = math.Max(final-strike, 0)
	} else {
		intrinsic = math.Max(strike-final, 0)
	}
	switch bt {
	case UpAndOut, DownAndOut:
		if touched {
			return 0 // knocked out
		}
		return intrinsic
	case UpAndIn, DownAndIn:
		if touched {
			return intrinsic // knocked in
		}
		return 0
	}
	return 0
}

func sampleStats(xs []float64) (mean, sd float64) {
	for _, x := range xs {
		mean += x
	}
	mean /= float64(len(xs))
	for _, x := range xs {
		sd += (x - mean) * (x - mean)
	}
	if len(xs) > 1 {
		sd = math.Sqrt(sd / float64(len(xs)-1))
	}
	return mean, sd
}
