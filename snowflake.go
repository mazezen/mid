package main

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	epoch          int64 = 1609459200000 // 2021-01-01 00:00:00 UTC
	timestampBits  = 41
	datacenterBits = 5
	machineBits    = 5
	sequenceBits   = 12
	maxDatacenter  = -1 ^ (-1 << datacenterBits) // 31
	maxMachine     = -1 ^ (-1 << machineBits)    // 31
	maxSequence    = -1 ^ (-1 << sequenceBits)   // 4095
	sequenceMask   = maxSequence
	timestampShift = datacenterBits + machineBits + sequenceBits
	datacenterShift = machineBits + sequenceBits
	machineShift    = sequenceBits
)

type Snowflake struct {
	mu sync.Mutex
	lastTimestamp int64
	datacenterID int64
	machineID int64
	sequence int64
	clockDrift int64 // 允许的时钟漂移（毫秒）
}

func NewSnowflake(datacenterID, machineID int64) (*Snowflake, error) {
	if datacenterID > maxDatacenter ||  machineID > maxMachine {
		return nil, fmt.Errorf("invalid datacenter or machine id");
	}
	s := &Snowflake{
		datacenterID: datacenterID,
		machineID: machineID,
		clockDrift: 1000, // 允许 1 秒漂移
	}
	return s, nil
}


// getTimestamp 获取当前时间戳
func (s *Snowflake) getTimestamp() int64 {
	return time.Now().UnixMilli()
}

func (s *Snowflake) NextID() (int64, error) {
	for {
		timestamp := s.getTimestamp()

		// 处理时钟回退（无锁）
		if timestamp < s.lastTimestamp {
			if s.lastTimestamp-timestamp <= s.clockDrift {
				for timestamp <= s.lastTimestamp {
					time.Sleep(time.Microsecond * 100)
					timestamp = s.getTimestamp()
				}
			} else {
				mLog.Error("Clock moved backwards beyond drift tolerance",
					zap.Int64("last_timestamp", s.lastTimestamp),
					zap.Int64("current_timestamp", timestamp))
				return 0, fmt.Errorf("clock moved backwards beyond drift tolerance")
			}
		}

		s.mu.Lock()

		// 再次检查时间戳
		if timestamp < s.lastTimestamp {
			s.mu.Unlock()
			continue
		}

		// 序列号处理
		if timestamp == s.lastTimestamp {
			s.sequence = (s.sequence + 1) & sequenceMask
			if s.sequence == 0 {
				// 序列号用尽，释放锁后等待
				s.mu.Unlock()
				for timestamp <= s.lastTimestamp {
					time.Sleep(time.Microsecond * 100)
					timestamp = s.getTimestamp()
				}
				continue
			}
		} else {
			s.sequence = 0
		}

		s.lastTimestamp = timestamp

		// 生成 ID
		id := ((timestamp - epoch) << timestampShift) |
			(s.datacenterID << datacenterShift) |
			(s.machineID << machineShift) |
			s.sequence

		s.mu.Unlock()
		return id, nil
	}
}