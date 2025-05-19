package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/mazezen/mid/proto/pb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func main() {
	Init()
	snowflake, err := NewSnowflake(1, 1)
	if err != nil {
		mLog.Error("创建snowfake失败", zap.Error(err))
	}

	// 启动 Prometheus 端点
	go func () {
		http.Handle("/metrics", promhttp.Handler())
		mLog.Info("Prometheus metrics server starting on :9190")
		if err := http.ListenAndServe(":9190", nil); err != nil {
			mLog.Error("Failed to start Prometheus server", zap.Error(err))
		}
	}()

	// dsn
	dsn := "root:123456@tcp(localhost:3306)/mid"
	segment, err := NewSegment(dsn)
	if err != nil {
		mLog.Error("创建segment失败", zap.Error(err))
	}
	segment.StartPreload()
	defer segment.Close()

	s := &server{
		snowflake: snowflake,
		segment: segment,
		snowfalkeBuffers: &struct{
			buffer1 *IDBuffer
            buffer2 *IDBuffer
            m1     sync.Mutex
            m2     sync.Mutex
			}{
				buffer1: NewIDBuffer(10000, 5000), // Buffer 大小 1000，阈值 50%
				buffer2: NewIDBuffer(10000, 5000),
			},
		segmentBuffers: &struct{
			buffer1 *IDBuffer
            buffer2 *IDBuffer
            m1     sync.Mutex
            m2     sync.Mutex
			}{
				buffer1: NewIDBuffer(10000, 5000), // Buffer 大小 1000，阈值 50%
				buffer2: NewIDBuffer(10000, 5000),
			 },
	}

	// 顺序填充 Buffer，避免并发竞争
	if err := s.fillBuffer(s.snowfalkeBuffers.buffer1, "snowflake"); err != nil {
		mLog.Error("初始化补充 snowflake buffer1 失败", zap.Error(err))
	}
	if err := s.fillBuffer(s.snowfalkeBuffers.buffer2, "snowflake"); err != nil {
		mLog.Error("初始化补充 snowflake buffer2 失败", zap.Error(err))
	}

	if err := s.fillBuffer(s.segmentBuffers.buffer1, "segment"); err != nil {
		mLog.Error("初始化补充 segment buffer1 失败", zap.Error(err))
	}
	if err := s.fillBuffer(s.segmentBuffers.buffer2, "segment"); err != nil {
		mLog.Error("初始化补充 segment buffer2 失败", zap.Error(err))
	}

	// 配置 gRPC 服务端 KeepAlive 参数
	serverOptions := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(10000), // 最大并发限流
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Second, // 最大空闲时间
			MaxConnectionAge:      30 * time.Second, // 最大连接存活时间
			MaxConnectionAgeGrace: 5 * time.Second,  // 优雅关闭的宽限期
			Time:                  5 * time.Second,  // 发送 ping 的间隔
			Timeout:               5 * time.Second,  // 等待 ping 响应的超时
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second, // 客户端 ping 的最小间隔
			PermitWithoutStream: true,            // 允许无活跃流时发送 ping
		}),
	}
	grpcServer := grpc.NewServer(serverOptions...)
	pb.RegisterIDMakerServer(grpcServer, s)

	lis, err := net.Listen("tcp", ":50051")
	fmt.Println("grpc server listen:", ":50051")
	if err != nil {
		mLog.Error("创建 gRPC 服务 失败", zap.Error(err))
	}
	
	mLog.Info("gRPC server running on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		mLog.Error("创建 gRPC 服务 失败", zap.Error(err))
	}
}