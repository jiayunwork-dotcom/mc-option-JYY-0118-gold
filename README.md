# mc-option

蒙特卡洛期权定价与风险度量 CLI（Go 标准库实现，无第三方依赖）。

## 功能

- 几何布朗运动（GBM）路径模拟：`S(t+dt) = S·exp((r - σ²/2)·dt + σ√dt·z)`，Box-Muller 产生标准正态，固定 seed 完全可复现
- 欧式 call/put、亚式（算术平均价）call/put 定价：折现期望收益 `e^{-rT}·mean(payoff)`，采用对偶变量（antithetic，一半路径用 -z）降低方差
- 风险度量：收益分布 VaR(95%)、ES(95%)、估计值标准误与 95% 置信区间

## 用法

flag 方式（全部有默认值，见下）：

```sh
mc-option -spot 100 -vol 0.2 -rate 0.05 -strike 105 -maturity 1 \
  -steps 64 -paths 20000 -seed 42 -type euro-call
```

`-type` 取值：`euro-call | euro-put | asian-call | asian-put`。

也可以用 `-params example/params.json` 从 JSON 文件读参数（覆盖同名默认值）：

```sh
mc-option -params example/params.json -type asian-call
```

JSON 字段：`spot, vol, rate, strike, maturity, steps, paths, seed`。

## 输出

正常时打印 price / stderr / 95% CI / VaR95 / ES95（VaR 与 ES 基于欧式 call 折现收益分布），exit 0；
参数非法（负值、steps<1、paths<100、未知 -type 等）时向 stderr 报错，exit 1，不会 panic。

## 包结构

| 包 | 说明 |
| --- | --- |
| `internal/rng` | Box-Muller 正态随机数、GBM 路径生成 |
| `internal/engine` | 参数校验、欧式/亚式定价、Payoff |
| `internal/risk` | PnL 序列、VaR、ES、置信区间 |

## 测试

```sh
go test ./...
```
