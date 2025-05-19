# 分布式 ID 生成系统

## 简介

分布式 ID 生成系统是一个高性能、可靠的 ID 生成服务，支持两种模式：Snowflake（基于时间戳的内存生成）和 Segment（基于 MySQL 的号段分配）。系统采用双 Buffer 策略优化性能，集成 Prometheus 监控和 Zap 结构化日志，确保高可用性和可观测性。通过 gRPC 提供服务接口，支持高并发场景下的唯一 ID 生成，适用于分布式系统中的订单号、用户 ID 等场景。

设计灵感来源于<a href="https://tech.meituan.com/2019/03/07/open-source-project-leaf.html">美团 Leaf 系统</a>，结合现代技术栈（gRPC、Prometheus、Zap），优化了性能和运维体验。





### BenchMark

- 硬件：8 核 CPU，16GB 内存
- MySQL：本地部署，单实例
- 网络：本地环回，无网络延迟

```bash
goos: darwin
goarch: amd64
pkg: github.com/mazezen/mid
cpu: VirtualApple @ 2.50GHz
BenchmarkSnowflake
BenchmarkSnowflake/Concurrency_10
    bench_test.go:140: Snowflake: Concurrency=10, TotalTime=31.4005ms, QPS=31846.63, AvgLatency=0.31ms, P99Latency=2.66ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=10, TotalTime=25.564333ms, QPS=39117.00, AvgLatency=0.25ms, P99Latency=1.89ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=10, TotalTime=19.667792ms, QPS=50844.55, AvgLatency=0.19ms, P99Latency=0.90ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=10, TotalTime=21.112542ms, QPS=47365.21, AvgLatency=0.21ms, P99Latency=0.63ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=10, TotalTime=17.301125ms, QPS=57799.71, AvgLatency=0.17ms, P99Latency=0.47ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=10, TotalTime=17.8015ms, QPS=56175.04, AvgLatency=0.17ms, P99Latency=0.53ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=10, TotalTime=18.653416ms, QPS=53609.48, AvgLatency=0.18ms, P99Latency=0.50ms, Errors=0
BenchmarkSnowflake/Concurrency_10-10         	1000000000	         0.02024 ns/op	     53609 QPS	         0.1816 avg_latency_ms	         0.4980 p99_latency_ms	       0 B/op	       0 allocs/op
BenchmarkSnowflake/Concurrency_50
    bench_test.go:140: Snowflake: Concurrency=50, TotalTime=69.456125ms, QPS=71987.89, AvgLatency=0.68ms, P99Latency=1.68ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=50, TotalTime=66.675208ms, QPS=74990.39, AvgLatency=0.66ms, P99Latency=1.76ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=50, TotalTime=69.426ms, QPS=72019.13, AvgLatency=0.69ms, P99Latency=1.62ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=50, TotalTime=69.431833ms, QPS=72013.08, AvgLatency=0.68ms, P99Latency=1.54ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=50, TotalTime=69.238209ms, QPS=72214.46, AvgLatency=0.68ms, P99Latency=1.82ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=50, TotalTime=68.337542ms, QPS=73166.23, AvgLatency=0.67ms, P99Latency=1.55ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=50, TotalTime=64.036916ms, QPS=78079.96, AvgLatency=0.63ms, P99Latency=1.78ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=50, TotalTime=67.973334ms, QPS=73558.26, AvgLatency=0.67ms, P99Latency=1.70ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=50, TotalTime=66.527958ms, QPS=75156.37, AvgLatency=0.66ms, P99Latency=1.57ms, Errors=0
BenchmarkSnowflake/Concurrency_50-10         	1000000000	         0.06800 ns/op	     75156 QPS	         0.6569 avg_latency_ms	         1.570 p99_latency_ms	       0 B/op	       0 allocs/op
BenchmarkSnowflake/Concurrency_100
    bench_test.go:140: Snowflake: Concurrency=100, TotalTime=123.748208ms, QPS=80809.25, AvgLatency=1.23ms, P99Latency=2.30ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=100, TotalTime=128.854667ms, QPS=77606.81, AvgLatency=1.28ms, P99Latency=2.20ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=100, TotalTime=128.356125ms, QPS=77908.24, AvgLatency=1.27ms, P99Latency=2.38ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=100, TotalTime=133.159666ms, QPS=75097.82, AvgLatency=1.32ms, P99Latency=2.61ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=100, TotalTime=129.191166ms, QPS=77404.67, AvgLatency=1.28ms, P99Latency=2.58ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=100, TotalTime=130.603ms, QPS=76567.92, AvgLatency=1.30ms, P99Latency=2.27ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=100, TotalTime=129.626667ms, QPS=77144.62, AvgLatency=1.29ms, P99Latency=2.50ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=100, TotalTime=132.06575ms, QPS=75719.86, AvgLatency=1.30ms, P99Latency=2.21ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=100, TotalTime=130.185625ms, QPS=76813.40, AvgLatency=1.29ms, P99Latency=2.62ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=100, TotalTime=131.359459ms, QPS=76126.99, AvgLatency=1.30ms, P99Latency=2.46ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=100, TotalTime=130.921292ms, QPS=76381.77, AvgLatency=1.30ms, P99Latency=2.61ms, Errors=0
BenchmarkSnowflake/Concurrency_100-10        	1000000000	         0.1330 ns/op	     76382 QPS	         1.297 avg_latency_ms	         2.608 p99_latency_ms	       0 B/op	       0 allocs/op
BenchmarkSnowflake/Concurrency_200
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=245.872042ms, QPS=81343.12, AvgLatency=2.44ms, P99Latency=3.86ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=238.704792ms, QPS=83785.50, AvgLatency=2.37ms, P99Latency=3.38ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=249.073333ms, QPS=80297.64, AvgLatency=2.47ms, P99Latency=3.72ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=254.07825ms, QPS=78715.91, AvgLatency=2.51ms, P99Latency=4.06ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=240.126208ms, QPS=83289.53, AvgLatency=2.38ms, P99Latency=3.68ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=241.384542ms, QPS=82855.35, AvgLatency=2.40ms, P99Latency=3.61ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=242.778291ms, QPS=82379.69, AvgLatency=2.41ms, P99Latency=4.19ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=235.83025ms, QPS=84806.76, AvgLatency=2.33ms, P99Latency=4.00ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=234.147167ms, QPS=85416.37, AvgLatency=2.32ms, P99Latency=3.69ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=235.215541ms, QPS=85028.40, AvgLatency=2.33ms, P99Latency=3.74ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=234.413875ms, QPS=85319.18, AvgLatency=2.33ms, P99Latency=3.74ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=236.196417ms, QPS=84675.29, AvgLatency=2.35ms, P99Latency=3.37ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=229.3235ms, QPS=87213.04, AvgLatency=2.28ms, P99Latency=3.35ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=234.777ms, QPS=85187.22, AvgLatency=2.33ms, P99Latency=3.76ms, Errors=0
    bench_test.go:140: Snowflake: Concurrency=200, TotalTime=239.484584ms, QPS=83512.68, AvgLatency=2.38ms, P99Latency=3.71ms, Errors=0
BenchmarkSnowflake/Concurrency_200-10        	1000000000	         0.2434 ns/op	     83513 QPS	         2.379 avg_latency_ms	         3.714 p99_latency_ms	       0 B/op	       0 allocs/op
BenchmarkSegment
BenchmarkSegment/Concurrency_10
    bench_test.go:156: Segment: Concurrency=10, TotalTime=25.108916ms, QPS=39826.49, AvgLatency=0.25ms, P99Latency=0.71ms, Errors=0
    bench_test.go:156: Segment: Concurrency=10, TotalTime=20.409959ms, QPS=48995.69, AvgLatency=0.20ms, P99Latency=0.63ms, Errors=0
    bench_test.go:156: Segment: Concurrency=10, TotalTime=18.534542ms, QPS=53953.32, AvgLatency=0.18ms, P99Latency=0.53ms, Errors=0
    bench_test.go:156: Segment: Concurrency=10, TotalTime=18.429333ms, QPS=54261.32, AvgLatency=0.18ms, P99Latency=0.52ms, Errors=0
    bench_test.go:156: Segment: Concurrency=10, TotalTime=21.68525ms, QPS=46114.29, AvgLatency=0.21ms, P99Latency=0.56ms, Errors=0
    bench_test.go:156: Segment: Concurrency=10, TotalTime=17.284208ms, QPS=57856.28, AvgLatency=0.17ms, P99Latency=0.56ms, Errors=0
    bench_test.go:156: Segment: Concurrency=10, TotalTime=16.612041ms, QPS=60197.30, AvgLatency=0.16ms, P99Latency=0.49ms, Errors=0
BenchmarkSegment/Concurrency_10-10           	1000000000	         0.01753 ns/op	     60197 QPS	         0.1603 avg_latency_ms	         0.4920 p99_latency_ms	       0 B/op	       0 allocs/op
BenchmarkSegment/Concurrency_50
    bench_test.go:156: Segment: Concurrency=50, TotalTime=74.85875ms, QPS=66792.46, AvgLatency=0.74ms, P99Latency=1.85ms, Errors=0
    bench_test.go:156: Segment: Concurrency=50, TotalTime=70.563958ms, QPS=70857.70, AvgLatency=0.69ms, P99Latency=1.81ms, Errors=0
    bench_test.go:156: Segment: Concurrency=50, TotalTime=68.416459ms, QPS=73081.83, AvgLatency=0.67ms, P99Latency=1.56ms, Errors=0
    bench_test.go:156: Segment: Concurrency=50, TotalTime=69.054333ms, QPS=72406.75, AvgLatency=0.68ms, P99Latency=1.56ms, Errors=0
    bench_test.go:156: Segment: Concurrency=50, TotalTime=71.882667ms, QPS=69557.80, AvgLatency=0.71ms, P99Latency=1.72ms, Errors=0
    bench_test.go:156: Segment: Concurrency=50, TotalTime=74.840959ms, QPS=66808.34, AvgLatency=0.74ms, P99Latency=1.61ms, Errors=0
    bench_test.go:156: Segment: Concurrency=50, TotalTime=69.617208ms, QPS=71821.32, AvgLatency=0.69ms, P99Latency=1.52ms, Errors=0
    bench_test.go:156: Segment: Concurrency=50, TotalTime=69.266666ms, QPS=72184.79, AvgLatency=0.69ms, P99Latency=1.51ms, Errors=0
    bench_test.go:156: Segment: Concurrency=50, TotalTime=67.938834ms, QPS=73595.61, AvgLatency=0.67ms, P99Latency=1.54ms, Errors=0
BenchmarkSegment/Concurrency_50-10           	1000000000	         0.06941 ns/op	     73596 QPS	         0.6716 avg_latency_ms	         1.539 p99_latency_ms	       0 B/op	       0 allocs/op
BenchmarkSegment/Concurrency_100
    bench_test.go:156: Segment: Concurrency=100, TotalTime=128.367084ms, QPS=77901.59, AvgLatency=1.27ms, P99Latency=2.22ms, Errors=0
    bench_test.go:156: Segment: Concurrency=100, TotalTime=130.579458ms, QPS=76581.72, AvgLatency=1.29ms, P99Latency=2.30ms, Errors=0
    bench_test.go:156: Segment: Concurrency=100, TotalTime=134.376917ms, QPS=74417.54, AvgLatency=1.33ms, P99Latency=2.48ms, Errors=0
    bench_test.go:156: Segment: Concurrency=100, TotalTime=127.6695ms, QPS=78327.24, AvgLatency=1.26ms, P99Latency=2.47ms, Errors=0
    bench_test.go:156: Segment: Concurrency=100, TotalTime=137.06425ms, QPS=72958.48, AvgLatency=1.36ms, P99Latency=2.82ms, Errors=0
    bench_test.go:156: Segment: Concurrency=100, TotalTime=131.346583ms, QPS=76134.45, AvgLatency=1.30ms, P99Latency=2.40ms, Errors=0
    bench_test.go:156: Segment: Concurrency=100, TotalTime=131.831917ms, QPS=75854.17, AvgLatency=1.31ms, P99Latency=2.45ms, Errors=0
    bench_test.go:156: Segment: Concurrency=100, TotalTime=130.654417ms, QPS=76537.79, AvgLatency=1.29ms, P99Latency=2.26ms, Errors=0
    bench_test.go:156: Segment: Concurrency=100, TotalTime=131.662209ms, QPS=75951.94, AvgLatency=1.30ms, P99Latency=2.31ms, Errors=0
    bench_test.go:156: Segment: Concurrency=100, TotalTime=129.734583ms, QPS=77080.45, AvgLatency=1.28ms, P99Latency=2.35ms, Errors=0
    bench_test.go:156: Segment: Concurrency=100, TotalTime=136.258667ms, QPS=73389.83, AvgLatency=1.35ms, P99Latency=2.38ms, Errors=0
BenchmarkSegment/Concurrency_100-10          	1000000000	         0.1383 ns/op	     73390 QPS	         1.349 avg_latency_ms	         2.380 p99_latency_ms	       0 B/op	       0 allocs/op
BenchmarkSegment/Concurrency_200
    bench_test.go:156: Segment: Concurrency=200, TotalTime=236.5885ms, QPS=84534.96, AvgLatency=2.35ms, P99Latency=3.55ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=242.685958ms, QPS=82411.03, AvgLatency=2.40ms, P99Latency=3.70ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=246.61525ms, QPS=81097.99, AvgLatency=2.44ms, P99Latency=3.80ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=237.710458ms, QPS=84135.97, AvgLatency=2.36ms, P99Latency=3.80ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=253.854209ms, QPS=78785.38, AvgLatency=2.52ms, P99Latency=7.37ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=261.853333ms, QPS=76378.63, AvgLatency=2.60ms, P99Latency=16.17ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=230.516417ms, QPS=86761.72, AvgLatency=2.29ms, P99Latency=3.76ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=235.531291ms, QPS=84914.41, AvgLatency=2.34ms, P99Latency=3.50ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=240.797709ms, QPS=83057.27, AvgLatency=2.38ms, P99Latency=3.91ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=242.735208ms, QPS=82394.31, AvgLatency=2.41ms, P99Latency=3.71ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=241.980542ms, QPS=82651.27, AvgLatency=2.40ms, P99Latency=3.58ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=243.96475ms, QPS=81979.06, AvgLatency=2.41ms, P99Latency=4.56ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=237.542959ms, QPS=84195.30, AvgLatency=2.36ms, P99Latency=3.61ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=247.454584ms, QPS=80822.91, AvgLatency=2.45ms, P99Latency=4.09ms, Errors=0
    bench_test.go:156: Segment: Concurrency=200, TotalTime=242.396375ms, QPS=82509.48, AvgLatency=2.41ms, P99Latency=3.56ms, Errors=0
BenchmarkSegment/Concurrency_200-10          	1000000000	         0.2462 ns/op	     82509 QPS	         2.406 avg_latency_ms	         3.556 p99_latency_ms	       0 B/op	       0 allocs/op
PASS
ok  	github.com/mazezen/mid	12.508s
```

### 性能测试结果汇总

以下表格总结了 Snowflake 和 Segment 模式的性能指标：

| 模式      | 并发级别 | QPS (请求/秒) | 平均延迟 (ms) | P99 延迟 (ms) | 错误次数 |
| --------- | -------- | ------------- | ------------- | ------------- | -------- |
| Snowflake | 10       | 53609.48      | 0.18          | 0.50          | 0        |
| Snowflake | 50       | 75156.37      | 0.66          | 1.57          | 0        |
| Snowflake | 100      | 76381.77      | 1.30          | 2.61          | 0        |
| Snowflake | 200      | 83512.68      | 2.38          | 3.71          | 0        |
| Segment   | 10       | 60197.30      | 0.16          | 0.49          | 0        |
| Segment   | 50       | 73595.61      | 0.67          | 1.54          | 0        |
| Segment   | 100      | 73389.83      | 1.35          | 2.38          | 0        |
| Segment   | 200      | 82509.48      | 2.41          | 3.56          | 0        |

### 说明



### 说明

- **数据来源**：每个并发级别的结果取自测试输出的最后一次运行（如 bench_test.go:140 和 bench_test.go:156 的最后一行），以反映稳定的性能指标。
- **QPS**：每秒请求数，反映吞吐量。Snowflake 模式在并发 200 时 QPS 最高（83512.68），Segment 模式在并发 10 时 QPS 较高（60197.30），但在高并发下略低于 Snowflake。
- **平均延迟**：平均每次请求的延迟（毫秒）。Segment 在低并发（10）下延迟略低（0.16ms vs. 0.18ms），但高并发下两者接近（~2.4ms @ 200）。
- **P99 延迟**：99% 请求的延迟，反映尾部延迟。Segment 在并发 200 时 P99 延迟略低（3.56ms vs. 3.71ms），但在某些运行中出现异常高峰（例如 16.17ms）。
- **错误次数**：所有测试均无错误（Errors=0），表明系统稳定性良好。
- **运行环境**：测试在 macOS（Darwin）、AMD64 架构、VirtualApple @ 2.50GHz CPU 上运行，Buffer 大小为 10000，请求数为 10 次/协程。

### 

## 设计目的

本系统旨在解决分布式系统中唯一 ID 生成的以下问题：

1. **唯一性**：保证全局唯一的 64 位整数 ID，无冲突。
2. **高性能**：支持高并发请求，低延迟（平均延迟 < 1ms），高吞吐量（QPS > 100,000）。
3. **高可用**：通过双 Buffer 和 MySQL 重试机制，减少服务中断。
4. **可扩展**：支持多种生成模式（Snowflake 和 Segment），便于业务扩展。
5. **可观测**：集成 Prometheus 监控 ID 生成速率、Buffer 使用率、MySQL 延迟和 NTP 偏移，使用 Zap 日志记录关键操作，便于调试和优化。

## 技术栈

- **编程语言**：Go 1.21+
- **服务框架**：gRPC（高性能 RPC 框架）
- **ID 生成**：
  - Snowflake：基于时间戳、数据中心 ID、机器 ID 和序列号，集成 `github.com/beevik/ntp` 进行时钟同步。
  - Segment：基于 MySQL 号段分配，集成 `github.com/cenkalti/backoff/v4` 实现指数退避重试。
- **数据库**：MySQL 8.0+（存储 Segment 模式的 ID 段）
- **监控**：Prometheus（`github.com/prometheus/client_golang`）暴露指标，Grafana 可选。
- **日志**：Zap（`go.uber.org/zap`）提供结构化、高性能日志。
- **依赖管理**：Go Modules

## 系统架构

- **核心组件**：
  - **Snowflake 模式**：内存生成 ID，依赖 NTP 同步时钟，适合高性能场景。
  - **Segment 模式**：通过 MySQL 分配 ID 段，适合需要持久化和严格递增的场景。
  - **双 Buffer**：每个模式维护两个 Buffer（buffer1 服务，buffer2 异步填充），减少阻塞。
- **服务接口**：gRPC 服务，提供 `GenerateID` 方法，通过 `mode` 参数选择生成模式（`snowflake` 或 `segment`）。
- **监控与日志**：
  - Prometheus 暴露 `/metrics` 端点，监控 ID 生成速率、Buffer 使用率、MySQL 延迟和 NTP 偏移。
  - Zap 记录 Buffer 填充、MySQL 查询、NTP 同步等关键事件。

## 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- protoc（Protocol Buffers 编译器）
- Prometheus（可选，用于监控）

### 安装依赖

```bash
go mod init idgenerator
go get google.golang.org/grpc
go get github.com/go-sql-driver/mysql
go get github.com/cenkalti/backoff/v4
go get github.com/prometheus/client_golang/prometheus
go get go.uber.org/zap
```

### 配置 MySQL

1. 创建数据库和表：

```sql
CREATE DATABASE mid;
USE mid;
CREATE TABLE id_segments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    biz_tag VARCHAR(50) NOT NULL UNIQUE,
    max_id BIGINT NOT NULL DEFAULT 0,
    step INT NOT NULL DEFAULT 10000,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
INSERT INTO id_segments (biz_tag, max_id, step) VALUES ('default', 0, 10000);
```

2. 更新 server.go 中的 MySQL 数据源（DSN）：

```base
dsn := "user:password@tcp(localhost:3306)/mid"
```

### 编译 gRPC 服务

```base
protoc --go_out=. --go-grpc_out=. id_maker.proto
```

### 运行服务端

```bash
go run server.go
```

* 服务监听 localhost:50051（gRPC）和 localhost:9190（Prometheus 指标）。

* 确认 Zap 日志：

```log
2025-05-17 14:29:11	INFO	mid/main.go:24	Prometheus metrics server starting on :9190
2025-05-17 14:29:11	INFO	mid/server.go:100	Buffer filled	{"mode": "snowflake", "size": 1000}
2025-05-17 14:29:11	INFO	mid/server.go:100	Buffer filled	{"mode": "snowflake", "size": 1000}
2025-05-17 14:29:11	INFO	mid/segment.go:101	Fetched new segment	{"new_max": 11000, "duration_seconds": 0.020109417}
2025-05-17 14:29:11	INFO	mid/server.go:100	Buffer filled	{"mode": "segment", "size": 1000}
2025-05-17 14:29:11	INFO	mid/segment.go:101	Fetched new segment	{"new_max": 12000, "duration_seconds": 0.0031245}
2025-05-17 14:29:11	INFO	mid/server.go:100	Buffer filled	{"mode": "segment", "size": 1000}
2025-05-17 14:29:11	INFO	mid/main.go:80	gRPC server running on :50051
```

### 运行客户端

```
go run client.go
```

- 输出示例：

  ```text
  Snowflake ID: 578779521390612487
  Segment ID: 1001
  ```



### 配置 Prometheus

1. 创建 prometheus.yml：

```yaml
global:
  scrape_interval: 15s
scrape_configs:
  - job_name: 'mid'
    static_configs:
      - targets: ['localhost:9190']
```

1. 运行 Prometheus：

```
prometheus --config.file=prometheus.yml
```

1. 访问 http://localhost:9090 查看指标：

* id_generate_total：ID 生成次数。
* buffer_usage：Buffer 剩余 ID 数量。
* mysql_query_duration_seconds：MySQL 查询延迟。
