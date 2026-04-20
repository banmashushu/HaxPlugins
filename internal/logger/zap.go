package logger

import (
	"go.uber.org/zap"
)

var log *zap.Logger

// Init 初始化日志
func Init() error {
	var err error
	config := zap.NewDevelopmentConfig()
	config.DisableCaller = true
	log, err = config.Build()
	if err != nil {
		return err
	}
	return nil
}

// Get 获取日志实例
func Get() *zap.Logger {
	if log == nil {
		_ = Init()
	}
	return log
}

// Sync 刷新日志缓冲区
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}
