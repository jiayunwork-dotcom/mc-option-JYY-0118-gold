// Package bs 提供 Black-Scholes 解析解，用作蒙特卡洛定价的对照基准。
// 仅处理欧式 vanilla call/put。
package bs

import (
	"errors"
	"math"
)

// ErrInvalidInput 参数不合法时返回。
var ErrInvalidInput = errors.New("bs: invalid input")

// Result 持有解析解结果。
type Result struct {
	Price float64
	D1    float64
	D2    float64
}

// Call 返回欧式 call 的 Black-Scholes 价格。
func Call(spot, strike, rate, vol, maturity float64) (Result, error) {
	if spot <= 0 || strike <= 0 || vol <= 0 || maturity <= 0 || rate <= 0 {
		return Result{}, ErrInvalidInput
	}
	d1, d2 := d1d2(spot, strike, rate, vol, maturity)
	price := spot*normCDF(d1) - strike*math.Exp(-rate*maturity)*normCDF(d2)
	return Result{Price: price, D1: d1, D2: d2}, nil
}

// Put 返回欧式 put 的 Black-Scholes 价格。
func Put(spot, strike, rate, vol, maturity float64) (Result, error) {
	if spot <= 0 || strike <= 0 || vol <= 0 || maturity <= 0 || rate <= 0 {
		return Result{}, ErrInvalidInput
	}
	d1, d2 := d1d2(spot, strike, rate, vol, maturity)
	price := strike*math.Exp(-rate*maturity)*normCDF(-d2) - spot*normCDF(-d1)
	return Result{Price: price, D1: d1, D2: d2}, nil
}

// PutCallParity 验证 put-call parity: C - P = S - K*exp(-rT)。
func PutCallParity(callPrice, putPrice, spot, strike, rate, maturity float64) float64 {
	lhs := callPrice - putPrice
	rhs := spot - strike*math.Exp(-rate*maturity)
	return lhs - rhs
}

func d1d2(spot, strike, rate, vol, maturity float64) (float64, float64) {
	sqrtT := math.Sqrt(maturity)
	d1 := (math.Log(spot/strike) + (rate+vol*vol/2)*maturity) / (vol * sqrtT)
	d2 := d1 - vol*sqrtT
	return d1, d2
}

// normCDF 标准正态累积分布函数（Abramowitz & Stegun 近似）。
func normCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt2))
}

// normPDF 标准正态概率密度函数。
func normPDF(x float64) float64 {
	return math.Exp(-x*x/2) / math.Sqrt(2*math.Pi)
}

// Delta 返回 call/put 的 delta。
func Delta(spot, strike, rate, vol, maturity float64, isCall bool) (float64, error) {
	if spot <= 0 || strike <= 0 || vol <= 0 || maturity <= 0 {
		return 0, ErrInvalidInput
	}
	d1, _ := d1d2(spot, strike, rate, vol, maturity)
	if isCall {
		return normCDF(d1), nil
	}
	return normCDF(d1) - 1, nil
}

// Gamma 返回 option 的 gamma（call 和 put 相同）。
func Gamma(spot, strike, rate, vol, maturity float64) (float64, error) {
	if spot <= 0 || strike <= 0 || vol <= 0 || maturity <= 0 {
		return 0, ErrInvalidInput
	}
	d1, _ := d1d2(spot, strike, rate, vol, maturity)
	return normPDF(d1) / (spot * vol * math.Sqrt(maturity)), nil
}

// Vega 返回 option 的 vega（对 vol 的偏导，以百分比表示）。
func Vega(spot, strike, rate, vol, maturity float64) (float64, error) {
	if spot <= 0 || strike <= 0 || vol <= 0 || maturity <= 0 {
		return 0, ErrInvalidInput
	}
	d1, _ := d1d2(spot, strike, rate, vol, maturity)
	return spot * normPDF(d1) * math.Sqrt(maturity) / 100, nil
}

// Theta 返回 call/put 的 theta（每天损耗）。
func Theta(spot, strike, rate, vol, maturity float64, isCall bool) (float64, error) {
	if spot <= 0 || strike <= 0 || vol <= 0 || maturity <= 0 {
		return 0, ErrInvalidInput
	}
	d1, d2 := d1d2(spot, strike, rate, vol, maturity)
	sqrtT := math.Sqrt(maturity)
	term1 := -spot * normPDF(d1) * vol / (2 * sqrtT)
	if isCall {
		return (term1 - rate*strike*math.Exp(-rate*maturity)*normCDF(d2)) / 365, nil
	}
	return (term1 + rate*strike*math.Exp(-rate*maturity)*normCDF(-d2)) / 365, nil
}

// Rho 返回 call/put 的 rho（对利率的偏导，以百分比表示）。
func Rho(spot, strike, rate, vol, maturity float64, isCall bool) (float64, error) {
	if spot <= 0 || strike <= 0 || vol <= 0 || maturity <= 0 {
		return 0, ErrInvalidInput
	}
	_, d2 := d1d2(spot, strike, rate, vol, maturity)
	if isCall {
		return strike * maturity * math.Exp(-rate*maturity) * normCDF(d2) / 100, nil
	}
	return -strike * maturity * math.Exp(-rate*maturity) * normCDF(-d2) / 100, nil
}
