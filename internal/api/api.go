// Package api 提供 mc-option 的 HTTP API 和前端文件服务。
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"mc-option/internal/bs"
	"mc-option/internal/engine"
	"mc-option/internal/greeks"
	"mc-option/internal/risk"
)

// Server 是 HTTP API 服务器。
type Server struct {
	mux    *http.ServeMux
	addr   string
	webDir string
}

// Config 服务器配置。
type Config struct {
	Addr   string
	WebDir string
}

// DefaultConfig 默认配置。
func DefaultConfig() Config {
	return Config{Addr: ":8080", WebDir: "web"}
}

// New 创建新的 API 服务器。
func New(cfg Config) *Server {
	s := &Server{mux: http.NewServeMux(), addr: cfg.Addr, webDir: cfg.WebDir}
	s.routes()
	return s
}

// Handler 返回 HTTP Handler（用于测试）。
func (s *Server) Handler() http.Handler { return s.mux }

// Addr 返回监听地址。
func (s *Server) Addr() string { return s.addr }

// ListenAndServe 启动服务器。
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(s.addr, s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/price", s.handlePrice)
	s.mux.HandleFunc("/api/greeks", s.handleGreeks)
	s.mux.HandleFunc("/api/bs", s.handleBS)
	s.mux.HandleFunc("/api/health", s.handleHealth)
	if s.webDir != "" {
		s.mux.Handle("/", http.FileServer(http.Dir(s.webDir)))
	}
}

// PriceRequest 定价请求。
type PriceRequest struct {
	Spot     float64 `json:"spot"`
	Vol      float64 `json:"vol"`
	Rate     float64 `json:"rate"`
	Strike   float64 `json:"strike"`
	Maturity float64 `json:"maturity"`
	Steps    int     `json:"steps"`
	Paths    int     `json:"paths"`
	Seed     int64   `json:"seed"`
	Type     string  `json:"type"` // euro-call|euro-put|asian-call|asian-put
}

// PriceResponse 定价响应。
type PriceResponse struct {
	Price   float64 `json:"price"`
	StdErr  float64 `json:"stderr"`
	CI95Lo  float64 `json:"ci95_lo"`
	CI95Hi  float64 `json:"ci95_hi"`
	VaR95   float64 `json:"var95"`
	ES95    float64 `json:"es95"`
}

// GreeksResponse 希腊字母响应。
type GreeksResponse struct {
	Delta float64 `json:"delta"`
	Gamma float64 `json:"gamma"`
	Vega  float64 `json:"vega"`
	Theta float64 `json:"theta"`
	Rho   float64 `json:"rho"`
}

// BSResponse Black-Scholes 响应。
type BSResponse struct {
	CallPrice float64 `json:"call_price"`
	PutPrice  float64 `json:"put_price"`
	Delta     float64 `json:"delta"`
	Gamma     float64 `json:"gamma"`
	Vega      float64 `json:"vega"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handlePrice(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpErr(w, 405, "POST only")
		return
	}
	var req PriceRequest
	if err := readBody(r, &req); err != nil {
		httpErr(w, 400, err.Error())
		return
	}
	p := engine.Params{
		Spot: req.Spot, Vol: req.Vol, Rate: req.Rate,
		Strike: req.Strike, Maturity: req.Maturity,
		Steps: req.Steps, Paths: req.Paths, Seed: req.Seed,
	}
	if err := engine.Validate(p); err != nil {
		httpErr(w, 422, err.Error())
		return
	}
	var pr engine.Price
	var err error
	isCall := true
	switch req.Type {
	case "euro-call":
		pr, err = engine.European(p, true)
	case "euro-put":
		pr, err = engine.European(p, false)
		isCall = false
	case "asian-call":
		pr, err = engine.Asian(p, true)
	case "asian-put":
		pr, err = engine.Asian(p, false)
		isCall = false
	default:
		httpErr(w, 400, "type must be euro-call|euro-put|asian-call|asian-put")
		return
	}
	_ = isCall
	if err != nil {
		httpErr(w, 500, err.Error())
		return
	}
	pnl, _ := risk.PnLSeries(p)
	v95, _ := risk.VaR(pnl, 0.95)
	es95, _ := risk.ES(pnl, 0.95)
	lo, hi := risk.CI(pr.Value, pr.StdErr)
	writeBody(w, PriceResponse{Price: pr.Value, StdErr: pr.StdErr, CI95Lo: lo, CI95Hi: hi, VaR95: v95, ES95: es95})
}

func (s *Server) handleGreeks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpErr(w, 405, "POST only")
		return
	}
	var req PriceRequest
	if err := readBody(r, &req); err != nil {
		httpErr(w, 400, err.Error())
		return
	}
	p := engine.Params{
		Spot: req.Spot, Vol: req.Vol, Rate: req.Rate,
		Strike: req.Strike, Maturity: req.Maturity,
		Steps: req.Steps, Paths: req.Paths, Seed: req.Seed,
	}
	isCall := req.Type == "euro-call" || req.Type == "asian-call"
	isAsian := req.Type == "asian-call" || req.Type == "asian-put"
	g, err := greeks.Compute(p, isCall, isAsian, greeks.DefaultConfig())
	if err != nil {
		httpErr(w, 422, err.Error())
		return
	}
	writeBody(w, GreeksResponse{Delta: g.Delta, Gamma: g.Gamma, Vega: g.Vega, Theta: g.Theta, Rho: g.Rho})
}

func (s *Server) handleBS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpErr(w, 405, "POST only")
		return
	}
	var req PriceRequest
	if err := readBody(r, &req); err != nil {
		httpErr(w, 400, err.Error())
		return
	}
	c, err := bs.Call(req.Spot, req.Strike, req.Rate, req.Vol, req.Maturity)
	if err != nil {
		httpErr(w, 422, err.Error())
		return
	}
	p, _ := bs.Put(req.Spot, req.Strike, req.Rate, req.Vol, req.Maturity)
	d, _ := bs.Delta(req.Spot, req.Strike, req.Rate, req.Vol, req.Maturity, true)
	g, _ := bs.Gamma(req.Spot, req.Strike, req.Rate, req.Vol, req.Maturity)
	v, _ := bs.Vega(req.Spot, req.Strike, req.Rate, req.Vol, req.Maturity)
	writeBody(w, BSResponse{CallPrice: c.Price, PutPrice: p.Price, Delta: d, Gamma: g, Vega: v})
}

func readBody(r *http.Request, v interface{}) error {
	b, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}
	return json.Unmarshal(b, v)
}

func writeBody(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func httpErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
