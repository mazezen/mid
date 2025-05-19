package main

import (
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/robfig/cron"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var mLog *zap.Logger

func Init() {
	encoder := getEncoder()
	writerSyncer := getLogWriter("./logs/mid.log")
	core := zapcore.NewCore(encoder, writerSyncer, zapcore.DebugLevel)
	var allCode []zapcore.Core
	allCode = append(allCode, core)
	c := zapcore.NewTee(allCode...)
	mLog = zap.New(c, zap.AddCaller())
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	// 在日志文件中使用大写字母记录日志级别
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	// NewConsoleEncoder 打印更符合观察的方式
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter(filepath string) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filepath,
		MaxSize:    10240,
		MaxAge:     7,
		Compress:   true,
		LocalTime:  true,
		MaxBackups: 4,
	}

	c := cron.New()
	c.AddFunc("0 0 0 1/1 * ?", func() {
		lumberJackLogger.Rotate()
	})
	c.Start()
	return zapcore.AddSync(lumberJackLogger)
}