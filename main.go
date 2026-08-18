// mc-option：蒙特卡洛期权定价与风险度量 CLI。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"mc-option/internal/api"
	"mc-option/internal/engine"
	"mc-option/internal/risk"
)

type options struct {
	Spot     float64 `json:"spot"`
	Vol      float64 `json:"vol"`
	Rate     float64 `json:"rate"`
	Strike   float64 `json:"strike"`
	Maturity float64 `json:"maturity"`
	Steps    int     `json:"steps"`
	Paths    int     `json:"paths"`
	Seed     int64   `json:"seed"`
}

func main() {
	opts := options{Spot: 100, Vol: 0.2, Rate: 0.05, Strike: 105, Maturity: 1, Steps: 64, Paths: 20000, Seed: 42}
	paramsFile := flag.String("params", "", "JSON 参数文件路径（可选，覆盖默认值）")
	typ := flag.String("type", "euro-call", "期权类型: euro-call|euro-put|asian-call|asian-put")
	serveMode := flag.Bool("serve", false, "启动 HTTP API 服务器")
	serveAddr := flag.String("addr", ":8080", "HTTP 监听地址（-serve 模式）")
	webDir := flag.String("web", "web", "前端静态文件目录（-serve 模式）")
	flag.Float64Var(&opts.Spot, "spot", opts.Spot, "标的现价")
	flag.Float64Var(&opts.Vol, "vol", opts.Vol, "波动率")
	flag.Float64Var(&opts.Rate, "rate", opts.Rate, "无风险利率")
	flag.Float64Var(&opts.Strike, "strike", opts.Strike, "行权价")
	flag.Float64Var(&opts.Maturity, "maturity", opts.Maturity, "期限（年）")
	flag.IntVar(&opts.Steps, "steps", opts.Steps, "每条路径的时间步数")
	flag.IntVar(&opts.Paths, "paths", opts.Paths, "模拟路径数")
	flag.Int64Var(&opts.Seed, "seed", opts.Seed, "随机种子")
	flag.Parse()

	if *serveMode {
		cfg := api.Config{Addr: *serveAddr, WebDir: *webDir}
		srv := api.New(cfg)
		fmt.Fprintf(os.Stderr, "mc-option: serving on %s\n", *serveAddr)
		if err := srv.ListenAndServe(); err != nil {
			fail(err)
		}
		return
	}

	if *paramsFile != "" {
		f, err := os.Open(*paramsFile)
		if err != nil {
			fail(err)
		}
		if err := json.NewDecoder(f).Decode(&opts); err != nil {
			f.Close()
			fail(err)
		}
		f.Close()
	}

	p := engine.Params{
		Spot:     opts.Spot,
		Vol:      opts.Vol,
		Rate:     opts.Rate,
		Strike:   opts.Strike,
		Maturity: opts.Maturity,
		Steps:    opts.Steps,
		Paths:    opts.Paths,
		Seed:     opts.Seed,
	}
	if err := engine.Validate(p); err != nil {
		fail(err)
	}

	var (
		price engine.Price
		err   error
	)
	switch *typ {
	case "euro-call":
		price, err = engine.European(p, true)
	case "euro-put":
		price, err = engine.European(p, false)
	case "asian-call":
		price, err = engine.Asian(p, true)
	case "asian-put":
		price, err = engine.Asian(p, false)
	default:
		err = fmt.Errorf("main: unknown -type %q (want euro-call|euro-put|asian-call|asian-put)", *typ)
	}
	if err != nil {
		fail(err)
	}

	pnl, err := risk.PnLSeries(p)
	if err != nil {
		fail(err)
	}
	var95, err := risk.VaR(pnl, 0.95)
	if err != nil {
		fail(err)
	}
	es95, err := risk.ES(pnl, 0.95)
	if err != nil {
		fail(err)
	}
	lo, hi := risk.CI(price.Value, price.StdErr)

	fmt.Printf("type=%s\n", *typ)
	fmt.Printf("price=%.6f\n", price.Value)
	fmt.Printf("stderr=%.6f\n", price.StdErr)
	fmt.Printf("ci95=[%.6f, %.6f]\n", lo, hi)
	fmt.Printf("var95=%.6f\n", var95)
	fmt.Printf("es95=%.6f\n", es95)
	os.Exit(0)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "mc-option:", err)
	os.Exit(1)
}
