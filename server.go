package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/mazezen/mid/proto/pb"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Prometheus 指标
var (
	idGenerateCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "id_generate_total",
			Help: "Total number of IDs generated",
		},
		[]string{"mode"},
	)
	bufferUsageGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "buffer_usage",
			Help: "Number of remaining IDs in buffer",
		},
		[]string{"mode", "buffer"},
	)
	mysqlQueryDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "mysql_query_duration_seconds",
			Help:    "MySQL query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)
	ntpOffsetGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ntp_offset_milliseconds",
			Help: "NTP clock offset in milliseconds",
		},
	)
)

func init() {
	prometheus.MustRegister(idGenerateCounter, bufferUsageGauge, mysqlQueryDuration, ntpOffsetGauge)
}
// IDBuffer 管理预生成的 ID 段
type IDBuffer struct {
	ids       []int64
	index     int
	size      int
	threshold int // 触发异步填充的阈值
}

type server struct {
	pb.UnimplementedIDMakerServer
	snowflake *Snowflake
	segment *Segment
	snowfalkeBuffers *struct {
		buffer1   *IDBuffer
		buffer2   *IDBuffer
		m1     sync.Mutex // buffer1 专用锁
        m2     sync.Mutex // buffer2 专用锁
	}
	segmentBuffers *struct {
		buffer1   *IDBuffer
		buffer2   *IDBuffer
		m1     sync.Mutex // buffer1 专用锁
        m2     sync.Mutex // buffer2 专用锁
	}
}

func NewIDBuffer(size, threshold int) *IDBuffer {
	return &IDBuffer{
		ids:       make([]int64, size),
		index:     0,
		size:      size,
		threshold: threshold,
	}
}

// 填充 Buffer
func (s *server) fillBuffer(buffer *IDBuffer, mode string) error {
	ids := make([]int64, buffer.size)
	for i := 0; i < buffer.size; i++ {
		var id int64
		var err error
		if mode == "snowflake" {
			id, err = s.snowflake.NextID()
		} 
		if mode == "segment" {
			id, err = s.segment.NextID()
		}
		if err != nil {
			return err
		}

		ids[i] = id
	}
	copy(buffer.ids, ids)
	buffer.index = 0
	mLog.Info("Buffer filled",
		zap.String("mode", mode),
		zap.Int("size", buffer.size))
	return nil
}

// MakeIDService gRPC 服务实现
func (s *server) MakeIDService(ctx context.Context, req *pb.MakeIDServiceRequest) (*pb.MakeIDServiceResponse, error) {
	mode := req.Mode
	if mode != "snowflake" && mode != "segment" {
		mLog.Error("Invalid mode",
			zap.String("mode", mode))
		return nil, fmt.Errorf("invalid mode: %s, must be 'snowflake' or 'segment'", mode)
	}

	var buffers *struct {
        buffer1 *IDBuffer
        buffer2 *IDBuffer
        m1     sync.Mutex
        m2     sync.Mutex
    }
	if mode == "snowflake" {
		buffers = s.snowfalkeBuffers
	} else {
		buffers = s.segmentBuffers
	}

	// 从 buffer1 获取 ID
	buffers.m1.Lock()
	if buffers.buffer1.index < buffers.buffer1.size {
		id := buffers.buffer1.ids[buffers.buffer1.index]
		buffers.buffer1.index++
		idGenerateCounter.WithLabelValues(mode).Inc()
		bufferUsageGauge.WithLabelValues(mode, "buffer1").Set(float64(buffers.buffer1.size - buffers.buffer1.index))
		bufferUsageGauge.WithLabelValues(mode, "buffer2").Set(float64(buffers.buffer2.size - buffers.buffer2.index))

		// 达到阈值且 buffer2 为空，异步填充 buffer2
		if buffers.buffer1.index >= buffers.buffer1.threshold && buffers.buffer2.index >= buffers.buffer2.size {
			go func() {
				buffers.m2.Lock()
				defer buffers.m2.Unlock()
				if err := s.fillBuffer(buffers.buffer2, mode); err != nil {
					mLog.Error("failed to fill buffer2", zap.String("mode", mode), zap.Error(err))
				}
			}()
		}
		buffers.m1.Unlock()
		return &pb.MakeIDServiceResponse{Id: id}, nil
	}
	buffers.m1.Unlock()

	// buffer1 用尽，切换到 buffer2
	buffers.m2.Lock()
	if buffers.buffer2.index < buffers.buffer2.size {
		buffers.m1.Lock()
		buffers.buffer1, buffers.buffer2 = buffers.buffer2, buffers.buffer1
		buffers.buffer2 = NewIDBuffer(1000, 500)
		buffers.m1.Unlock()
		go func () {
			buffers.m2.Lock()
			defer buffers.m2.Unlock()
			if err := s.fillBuffer(buffers.buffer2, mode); err != nil {
				mLog.Error("failed to fill buffer2", zap.String("mode", mode), zap.Error(err))
			}
		}()
		buffers.m2.Unlock()
		return s.MakeIDService(ctx, req)
	}
	buffers.m2.Unlock()

	// 两个 Buffer 都用尽，同步填充 buffer1
	buffers.m1.Lock()
	if err := s.fillBuffer(buffers.buffer1, mode); err != nil {
		mLog.Error("Failed to fill buffer1",
			zap.String("mode", mode),
			zap.Error(err))
			buffers.m1.Unlock()
		return nil, err
	}
	buffers.m1.Unlock()
	return s.MakeIDService(ctx, req)
}
