// stats.go 提供蒙特卡洛结果的统计分析工具函数。
package engine

import (
	"math"
	"sort"
)

// Percentile 计算有序样本的 p 分位数（0-100）。
func Percentile(xs []float64, p float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	sorted := append([]float64(nil), xs...)
	sort.Float64s(sorted)
	idx := p / 100.0 * float64(len(sorted)-1)
	lo := int(math.Floor(idx))
	hi := int(math.Ceil(idx))
	if lo < 0 {
		lo = 0
	}
	if hi >= len(sorted) {
		hi = len(sorted) - 1
	}
	frac := idx - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

// Skewness 计算样本偏度。
func Skewness(xs []float64) float64 {
	n := float64(len(xs))
	if n < 3 {
		return 0
	}
	mean, sd := sampleStats(xs)
	if sd == 0 {
		return 0
	}
	sum3 := 0.0
	for _, x := range xs {
		d := (x - mean) / sd
		sum3 += d * d * d
	}
	return sum3 * n / ((n - 1) * (n - 2))
}

// Kurtosis 计算样本超额峰度。
func Kurtosis(xs []float64) float64 {
	n := float64(len(xs))
	if n < 4 {
		return 0
	}
	mean, sd := sampleStats(xs)
	if sd == 0 {
		return 0
	}
	sum4 := 0.0
	for _, x := range xs {
		d := (x - mean) / sd
		sum4 += d * d * d * d
	}
	k := (n*(n+1)*sum4)/((n-1)*(n-2)*(n-3)) - 3*(n-1)*(n-1)/((n-2)*(n-3))
	return k
}

// Histogram 将样本分成 bins 个等宽箱并返回频率。
type HistogramBin struct {
	Lower float64 `json:"lower"`
	Upper float64 `json:"upper"`
	Count int     `json:"count"`
}

// BuildHistogram 构建直方图。
func BuildHistogram(xs []float64, bins int) []HistogramBin {
	if len(xs) == 0 || bins < 1 {
		return nil
	}
	sorted := append([]float64(nil), xs...)
	sort.Float64s(sorted)
	min, max := sorted[0], sorted[len(sorted)-1]
	if min == max {
		return []HistogramBin{{Lower: min, Upper: max, Count: len(xs)}}
	}
	width := (max - min) / float64(bins)
	hist := make([]HistogramBin, bins)
	for i := 0; i < bins; i++ {
		hist[i] = HistogramBin{
			Lower: min + float64(i)*width,
			Upper: min + float64(i+1)*width,
		}
	}
	for _, x := range xs {
		idx := int((x - min) / width)
		if idx >= bins {
			idx = bins - 1
		}
		if idx < 0 {
			idx = 0
		}
		hist[idx].Count++
	}
	return hist
}

// MedianAbsoluteDeviation 计算 MAD。
func MedianAbsoluteDeviation(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	med := Percentile(xs, 50)
	devs := make([]float64, len(xs))
	for i, x := range xs {
		devs[i] = math.Abs(x - med)
	}
	sort.Float64s(devs)
	return Percentile(devs, 50)
}
