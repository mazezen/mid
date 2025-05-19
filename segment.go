package main

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

// Segment 结构体
type Segment struct {
	db      *sql.DB
	bizTag  string // 业务标识
	current int64  // 当前ID
	max     int64  // 当前段的最大 ID
	step    int64  // 每次分配的 ID 段大小
	mu      sync.Mutex
}

func NewSegment(dsn string) (*Segment, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		mLog.Error("Failed to connect to MySQL", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to mysql: %v", err)
	}
	if err := db.Ping(); err != nil {
		mLog.Error("Failed to ping MySQL", zap.Error(err))
		return nil, fmt.Errorf("failed to ping MySQL: %v", err)
	}
	db.SetMaxOpenConns(10000)
	db.SetConnMaxIdleTime(500)
	return &Segment{
		db:     db,
		bizTag: "default",
		step:   10000, // 每次分配 10000 个 ID
	}, nil
}

func (s *Segment) fetchNewSegment() (int64, error) {
	var newMax int64
	startTime := time.Now()
	operation := func() error {
		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start transaction: %v", err)
		}
		defer tx.Rollback()

		// 原子更新 max_id
		result, err := tx.Exec(
			"UPDATE id_segments SET max_id = max_id + ? WHERE biz_tag = ?",
			s.step, s.bizTag,
		)
		if err != nil {
			return fmt.Errorf("failed to update max_id: %v", err)
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil || rowsAffected != 1 {
			return fmt.Errorf("failed to update id_segments")
		}

		// 获取新的 max_id
		err = tx.QueryRow(
			"SELECT max_id FROM id_segments WHERE biz_tag = ?",
			s.bizTag,
		).Scan(&newMax)
		if err != nil {
			return fmt.Errorf("failed to query max_id: %v", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %v", err)
		}
		return nil
	}

	// 配置指数退避重试
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 100 * time.Millisecond
	b.MaxInterval = 2 * time.Second
	b.MaxElapsedTime = 10 * time.Second

	err := backoff.Retry(func() error {
		err := operation()
		if err != nil {
			mLog.Warn("Retrying MySQL operation", zap.Error(err))
		}
		return err
	}, b)
	if err != nil {
		mLog.Error("Failed to fetch new segment after retries", zap.Error(err))
		return 0, err
	}

	duration := time.Since(startTime).Seconds()
	mysqlQueryDuration.Observe(duration)
	mLog.Info("Fetched new segment",
		zap.Int64("new_max", newMax),
		zap.Float64("duration_seconds", duration))
	return newMax, nil
}

func (s *Segment) StartPreload() {
	go func () {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C { 
			s.mu.Lock()
			if s.current+5000 >= s.max { // 剩余 ID 少于 50% 时预加载
				newMax, err := s.fetchNewSegment()
				if err != nil {
					mLog.Error("Failed to preload segment", zap.Error(err))
					s.mu.Unlock()
					continue
				}
				s.current = newMax - s.step + 1
				s.max = newMax
			}
			s.mu.Unlock()
		}
	}()
}

// NextID 获取下一个 ID
func (s *Segment) NextID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 当前段有可用 ID
	if s.current < s.max {
		s.current++
		return s.current, nil
	}

	// 获取新 ID 段
	newMax, err := s.fetchNewSegment()
	if err != nil {
		return 0, err
	}

	s.current = newMax - s.step + 1
	s.max = newMax
	return s.current, nil
}

// Close 关闭 MySQL 连接
func (s *Segment) Close() error {
	return s.db.Close()
}
