package gormlogger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

type SlogAdapter struct {
	log      *slog.Logger
	logLevel gormlogger.LogLevel
}

func NewSlogAdapter(log *slog.Logger) *SlogAdapter {
	return &SlogAdapter{log: log, logLevel: gormlogger.Info}
}

func (s *SlogAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return &SlogAdapter{log: s.log, logLevel: level}
}

func (s *SlogAdapter) Info(_ context.Context, msg string, args ...interface{}) {
	if s.logLevel >= gormlogger.Info {
		s.log.Info(fmt.Sprintf(msg, args...))
	}
}

func (s *SlogAdapter) Warn(_ context.Context, msg string, args ...interface{}) {
	if s.logLevel >= gormlogger.Warn {
		s.log.Warn(fmt.Sprintf(msg, args...))
	}
}

func (s *SlogAdapter) Error(_ context.Context, msg string, args ...interface{}) {
	if s.logLevel >= gormlogger.Error {
		s.log.Error(fmt.Sprintf(msg, args...))
	}
}

func (s *SlogAdapter) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if s.logLevel < gormlogger.Info {
		return
	}
	sql, rows := fc()
	elapsed := time.Since(begin)
	if err != nil {
		s.log.Error("gorm trace", "sql", sql, "rows", rows, "elapsed", elapsed, "err", err)
	} else {
		s.log.Debug("gorm trace", "sql", sql, "rows", rows, "elapsed", elapsed)
	}
}
