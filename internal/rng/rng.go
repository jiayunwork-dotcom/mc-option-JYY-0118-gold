// Package rng 提供基于 Box-Muller 变换的标准正态随机数
// 与几何布朗运动（GBM）离散路径生成。同 seed 产生完全相同的序列。
package rng

import (
	"errors"
	"math"
	"math/rand"
)

// Normal 基于math/rand的确定性标准正态随机源。
type Normal struct {
	src  *rand.Rand
	spare float64
	hasSpare bool
}

// NewNormal 返回以 seed 初始化的正态随机源。
func NewNormal(seed int64) *Normal {
	return &Normal{src: rand.New(rand.NewSource(seed))}
}

// Next 返回下一个标准正态样本（Box-Muller，缓存第二个分量）。
func (n *Normal) Next() float64 {
	if n.hasSpare {
		n.hasSpare = false
		return n.spare
	}
	u1 := n.src.Float64()
	for u1 <= 0 {
		u1 = n.src.Float64()
	}
	u2 := n.src.Float64()
	r := math.Sqrt(-2 * math.Log(u1))
	theta := 2 * math.Pi * u2
	n.spare = r * math.Sin(theta)
	n.hasSpare = true
	return r * math.Cos(theta)
}

// Antithetic 返回第 i 个抽样 z 的对偶样本，即 -z。
func (n *Normal) Antithetic(i int, z float64) float64 {
	return -z
}

// Path 按 GBM 离散格式生成价格路径：
// S(t+dt) = S·exp((drift - σ²/2)·dt + σ√dt·z)，
// 返回长度 len(zs)+1 的切片（含初始 spot）。zs 为空时返回 error。
func Path(spot, drift, vol, dt float64, zs []float64) ([]float64, error) {
	if len(zs) < 1 {
		return nil, errors.New("rng: zs must contain at least one draw")
	}
	path := make([]float64, len(zs)+1)
	path[0] = spot
	growth := math.Exp((drift - vol*vol/2) * dt)
	diffusion := vol * math.Sqrt(dt)
	for i, z := range zs {
		path[i+1] = path[i] * growth * math.Exp(diffusion*z)
	}
	return path, nil
}
