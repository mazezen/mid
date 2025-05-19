package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mazezen/mid/proto/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// 测试参数
const (
	serverAddr   = "localhost:50051" // gRPC 服务地址
	requestCount = 100            // 每个协程的请求次数
)

// 并发级别
var concurrencyLevels = []int{10, 50, 100, 200}

// 测试结果
type benchResult struct {
	Mode        string        // 测试模式（snowflake/segment）
	Concurrency int           // 并发数
	TotalTime   time.Duration // 总耗时
	QPS         float64       // 每秒请求数
	AvgLatency  float64       // 平均延迟（毫秒）
	P99Latency  float64       // P99 延迟（毫秒）
	Errors      int64         // 错误次数
}

// runBenchmark 运行单次压测
func runBenchmark(b *testing.B, mode string, concurrency int) benchResult {
	// 初始化 gRPC 客户端，启用 KeepAlive 和压缩
	conn, err := grpc.Dial(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
    	grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second,
			Timeout:             2 * time.Second,
			PermitWithoutStream: true,
		}),
		// grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")),
	)
	if err != nil {
		b.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()
	client := pb.NewIDMakerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 500 * time.Millisecond)
	defer cancel()
	

	// 统计指标
	var totalLatency int64
	var errorCount int64
	latencies := make(chan int64, requestCount*concurrency)
	var wg sync.WaitGroup
	startTime := time.Now()

	// 启动并发协程
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < requestCount; j++ {
				reqStart := time.Now()
				_, err := client.MakeIDService(ctx, &pb.MakeIDServiceRequest{Mode: mode})
				latency := time.Since(reqStart).Microseconds()
				latencies <- latency
				if err != nil {
					atomic.AddInt64(&errorCount, 1)
					b.Logf("Request failed: %v", err)
				}
			}
		}()
	}

	// 收集延迟
	p99Latencies := make([]int64, 0, requestCount*concurrency)
	go func() {
		for latency := range latencies {
			totalLatency += latency
			p99Latencies = append(p99Latencies, latency)
		}
	}()

	// 等待所有请求完成
	wg.Wait()
	close(latencies)
	totalTime := time.Since(startTime)

	// 计算 P99 延迟
	var p99Latency float64
	if len(p99Latencies) > 0 {
		sort.Slice(p99Latencies, func(i, j int) bool {
			return p99Latencies[i] < p99Latencies[j]
		})
		p99Index := int(float64(len(p99Latencies)) * 0.99)
		if p99Index >= len(p99Latencies) {
			p99Index = len(p99Latencies) - 1
		}
		p99Latency = float64(p99Latencies[p99Index]) / 1000.0 // 微秒转毫秒
	}

	// 计算结果
	totalRequests := int64(requestCount * concurrency)
	avgLatency := float64(totalLatency) / float64(totalRequests) / 1000.0 // 微秒转毫秒
	qps := float64(totalRequests) / totalTime.Seconds()

	return benchResult{
		Mode:        mode,
		Concurrency: concurrency,
		TotalTime:   totalTime,
		QPS:         qps,
		AvgLatency:  avgLatency,
		P99Latency:  p99Latency,
		Errors:      errorCount,
	}
}

// BenchmarkSnowflake 测试 Snowflake 模式
func BenchmarkSnowflake(b *testing.B) {
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			b.ResetTimer()
			result := runBenchmark(b, "snowflake", concurrency)
			b.StopTimer()
			b.ReportMetric(result.QPS, "QPS")
			b.ReportMetric(result.AvgLatency, "avg_latency_ms")
			b.ReportMetric(result.P99Latency, "p99_latency_ms")
			b.Logf("Snowflake: Concurrency=%d, TotalTime=%v, QPS=%.2f, AvgLatency=%.2fms, P99Latency=%.2fms, Errors=%d",
				result.Concurrency, result.TotalTime, result.QPS, result.AvgLatency, result.P99Latency, result.Errors)
		})
	}
}

// BenchmarkSegment 测试 Segment 模式
func BenchmarkSegment(b *testing.B) {
	for _, concurrency := range concurrencyLevels {
		b.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(b *testing.B) {
			b.ResetTimer()
			result := runBenchmark(b, "segment", concurrency)
			b.StopTimer()
			b.ReportMetric(result.QPS, "QPS")
			b.ReportMetric(result.AvgLatency, "avg_latency_ms")
			b.ReportMetric(result.P99Latency, "p99_latency_ms")
			b.Logf("Segment: Concurrency=%d, TotalTime=%v, QPS=%.2f, AvgLatency=%.2fms, P99Latency=%.2fms, Errors=%d",
				result.Concurrency, result.TotalTime, result.QPS, result.AvgLatency, result.P99Latency, result.Errors)
		})
	}
}