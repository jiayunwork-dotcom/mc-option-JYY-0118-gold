// implied.go 提供隐含波动率求解和波动率曲面拟合工具。
package bs

import (
	"errors"
	"math"
)

// ErrNoConverge 二分法未收敛。
var ErrNoConverge = errors.New("bs: implied vol did not converge")

// ImpliedVol 用二分法求解欧式期权的隐含波动率。
// marketPrice 为市场观察到的期权价格，isCall 指定 call/put。
func ImpliedVol(spot, strike, rate, maturity, marketPrice float64, isCall bool) (float64, error) {
	if marketPrice <= 0 || spot <= 0 || strike <= 0 || maturity <= 0 {
		return 0, ErrInvalidInput
	}
	lo, hi := 0.001, 5.0
	for iter := 0; iter < 200; iter++ {
		mid := (lo + hi) / 2
		var pr float64
		if isCall {
			r, err := Call(spot, strike, rate, mid, maturity)
			if err != nil {
				return 0, err
			}
			pr = r.Price
		} else {
			r, err := Put(spot, strike, rate, mid, maturity)
			if err != nil {
				return 0, err
			}
			pr = r.Price
		}
		if math.Abs(pr-marketPrice) < 1e-8 {
			return mid, nil
		}
		if pr < marketPrice {
			lo = mid
		} else {
			hi = mid
		}
		if hi-lo < 1e-10 {
			break
		}
	}
	return (lo + hi) / 2, nil
}

// ImpliedVolNewton 用 Newton-Raphson 法求解隐含波动率（更快收敛）。
func ImpliedVolNewton(spot, strike, rate, maturity, marketPrice float64, isCall bool) (float64, error) {
	if marketPrice <= 0 || spot <= 0 || strike <= 0 || maturity <= 0 {
		return 0, ErrInvalidInput
	}
	vol := 0.2 // initial guess
	for iter := 0; iter < 100; iter++ {
		var pr float64
		if isCall {
			r, _ := Call(spot, strike, rate, vol, maturity)
			pr = r.Price
		} else {
			r, _ := Put(spot, strike, rate, vol, maturity)
			pr = r.Price
		}
		v, _ := Vega(spot, strike, rate, vol, maturity)
		vegaFull := v * 100 // Vega was divided by 100 in bs.go
		if math.Abs(vegaFull) < 1e-12 {
			return vol, ErrNoConverge
		}
		diff := pr - marketPrice
		if math.Abs(diff) < 1e-8 {
			return vol, nil
		}
		vol -= diff / vegaFull
		if vol <= 0 {
			vol = 0.001
		}
	}
	return vol, ErrNoConverge
}

// VolSmile 对一系列行权价计算隐含波动率，生成微笑曲线数据。
type SmilePoint struct {
	Strike float64 `json:"strike"`
	IV     float64 `json:"iv"`
}

// ComputeSmile 根据一组市场价格计算隐含波动率微笑。
func ComputeSmile(spot, rate, maturity float64, strikes, marketPrices []float64, isCall bool) ([]SmilePoint, error) {
	if len(strikes) != len(marketPrices) {
		return nil, errors.New("bs: strikes and prices must have same length")
	}
	var points []SmilePoint
	for i, k := range strikes {
		iv, err := ImpliedVol(spot, k, rate, maturity, marketPrices[i], isCall)
		if err != nil {
			continue
		}
		points = append(points, SmilePoint{Strike: k, IV: iv})
	}
	return points, nil
}

// Moneyness 计算期权的 moneyness: S/K。
func Moneyness(spot, strike float64) float64 {
	if strike == 0 {
		return 0
	}
	return spot / strike
}

// LogMoneyness 计算对数 moneyness: ln(S/K)。
func LogMoneyness(spot, strike float64) float64 {
	if strike <= 0 || spot <= 0 {
		return 0
	}
	return math.Log(spot / strike)
}
