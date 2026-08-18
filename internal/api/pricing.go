// pricing.go 扩展 api 包，提供批量定价和参数验证工具。
package api

import "mc-option/internal/engine"

// BatchPriceRequest 批量定价请求。
type BatchPriceRequest struct {
	Requests []PriceRequest `json:"requests"`
}

// BatchPriceResponse 批量定价响应。
type BatchPriceResponse struct {
	Results []PriceResponse `json:"results"`
	Errors  []string        `json:"errors,omitempty"`
}

// ValidateParams 校验参数并返回默认值。
func ValidateParams(req PriceRequest) (*engine.Params, error) {
	p := engine.Params{
		Spot:     req.Spot,
		Vol:      req.Vol,
		Rate:     req.Rate,
		Strike:   req.Strike,
		Maturity: req.Maturity,
		Steps:    req.Steps,
		Paths:    req.Paths,
		Seed:     req.Seed,
	}
	if p.Steps == 0 {
		p.Steps = 64
	}
	if p.Paths == 0 {
		p.Paths = 10000
	}
	if p.Seed == 0 {
		p.Seed = 42
	}
	if err := engine.Validate(p); err != nil {
		return nil, err
	}
	return &p, nil
}

// OptionTypeIsCall 判断期权类型字符串是否为 call。
func OptionTypeIsCall(typ string) bool {
	return typ == "euro-call" || typ == "asian-call"
}

// OptionTypeIsAsian 判断期权类型字符串是否为亚式。
func OptionTypeIsAsian(typ string) bool {
	return typ == "asian-call" || typ == "asian-put"
}

// SupportedTypes 返回支持的期权类型列表。
func SupportedTypes() []string {
	return []string{"euro-call", "euro-put", "asian-call", "asian-put"}
}

// IsValidType 检查类型字符串是否合法。
func IsValidType(typ string) bool {
	for _, t := range SupportedTypes() {
		if t == typ {
			return true
		}
	}
	return false
}
