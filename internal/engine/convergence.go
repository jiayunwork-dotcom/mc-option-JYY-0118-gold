// convergence.go 提供收敛性分析：按路径数递增观察价格估计的稳定性。
package engine

import "math"

// ConvergencePoint 收敛曲线上一个数据点。
type ConvergencePoint struct {
	Paths  int     `json:"paths"`
	Price  float64 `json:"price"`
	StdErr float64 `json:"stderr"`
}

// ConvergenceStudy 按 paths 从 startPaths 到 endPaths 等比递增进行定价，
// 返回收敛曲线。
func ConvergenceStudy(p Params, isCall bool, startPaths, endPaths, numPoints int) ([]ConvergencePoint, error) {
	if numPoints < 2 {
		numPoints = 2
	}
	if startPaths < 100 {
		startPaths = 100
	}
	if endPaths <= startPaths {
		endPaths = startPaths * 10
	}
	ratio := math.Pow(float64(endPaths)/float64(startPaths), 1.0/float64(numPoints-1))
	var points []ConvergencePoint
	for i := 0; i < numPoints; i++ {
		n := int(float64(startPaths) * math.Pow(ratio, float64(i)))
		if n < 100 {
			n = 100
		}
		pp := p
		pp.Paths = n
		pr, err := European(pp, isCall)
		if err != nil {
			return nil, err
		}
		points = append(points, ConvergencePoint{Paths: n, Price: pr.Value, StdErr: pr.StdErr})
	}
	return points, nil
}

// HasConverged 检查最后两个点的价格差是否在 tolerance 内。
func HasConverged(points []ConvergencePoint, tolerance float64) bool {
	if len(points) < 2 {
		return false
	}
	last := points[len(points)-1]
	prev := points[len(points)-2]
	return math.Abs(last.Price-prev.Price) < tolerance
}

// RelativeError 计算蒙特卡洛估计相对于参考值的相对误差。
func RelativeError(estimate, reference float64) float64 {
	if reference == 0 {
		return math.Abs(estimate)
	}
	return math.Abs(estimate-reference) / math.Abs(reference)
}

// RequiredPaths 根据目标标准误估算所需路径数。
// stderr ∝ 1/√N → N ≈ (currentStdErr * √currentN / targetStdErr)²
func RequiredPaths(currentStdErr float64, currentPaths int, targetStdErr float64) int {
	if targetStdErr <= 0 || currentStdErr <= 0 {
		return currentPaths
	}
	ratio := currentStdErr / targetStdErr
	n := int(math.Ceil(float64(currentPaths) * ratio * ratio))
	if n < 100 {
		n = 100
	}
	return n
}
