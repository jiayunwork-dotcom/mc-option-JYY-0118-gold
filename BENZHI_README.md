# mc-option

蒙特卡洛期权定价与风险度量工具。

## 功能

- 欧式期权（call/put）蒙特卡洛定价
- 亚式期权（算术平均价）定价
- 障碍期权（knock-in/knock-out）定价
- Black-Scholes 解析解对照
- 希腊字母（Delta/Gamma/Vega/Theta/Rho）数值估算
- VaR/ES 风险度量
- HTTP API + Web UI

## 构建与运行

```bash
go build -o mc-option .
./mc-option -type euro-call -spot 100 -vol 0.2 -rate 0.05 -strike 105 -maturity 1
./mc-option serve -addr :8080
```

## Docker

```bash
docker build -t mc-option .
docker run --rm -p 8080:8080 mc-option serve -addr :8080
```
