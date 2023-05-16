package gormzerolog

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm/logger"
)

var loggers = map[logger.LogLevel]func() *zerolog.Event{
	logger.Info:  log.Info,
	logger.Warn:  log.Warn,
	logger.Error: log.Error,
}

const (
	traceErrMsg  = "%s %s\n[%.3fms] [rows:%v] %s"
	traceWarnMsg = "%s %s\n[%.3fms] [rows:%v] %s"
	traceInfoMsg = "%s\n[%.3fms] [rows:%v] %s"
)

type GormLogger struct {
	logLevel                logger.LogLevel
	ignoreRecordNotFoundErr bool
	slowThreshold           time.Duration
}

func NewGormLogger() *GormLogger {
	return &GormLogger{
		logLevel:                logger.Info,
		ignoreRecordNotFoundErr: true,
		slowThreshold:           time.Millisecond * 200,
	}
}

func (l *GormLogger) IgnoreRecordNotFoundError(b bool) {
	l.ignoreRecordNotFoundErr = b
}

func (l *GormLogger) LogMode(logLevel logger.LogLevel) logger.Interface {
	l.logLevel = logLevel
	return l
}

func (l *GormLogger) SlowThreshold(slowThreshold time.Duration) {
	l.slowThreshold = slowThreshold
}

func (l *GormLogger) log(logLevel logger.LogLevel, msg string, data ...any) {
	if l.logLevel >= logLevel {
		if l, ok := loggers[logLevel]; ok {
			l().Str("source", "GORM").Msgf(msg, data...)
		}
	}
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...any) {
	l.log(logger.Info, msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...any) {
	l.log(logger.Warn, msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...any) {
	l.log(logger.Error, msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	var rowsAffected any = rows
	if rows == -1 {
		rowsAffected = "-"
	}

	switch {
	case err != nil && (!errors.Is(err, logger.ErrRecordNotFound) || !l.ignoreRecordNotFoundErr):
		l.log(logger.Error, traceErrMsg, fileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rowsAffected, sql)
	case elapsed > l.slowThreshold && l.slowThreshold != 0:
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.slowThreshold)
		l.log(logger.Warn, traceWarnMsg, fileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rowsAffected, sql)
	}

	l.log(logger.Info, traceInfoMsg, fileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rowsAffected, sql)
}

var gormSourceDir string

// FileWithLineNum return the file name and line number of the current file
func fileWithLineNum() string {
	// the second caller usually from gorm internal, so set i start from 2
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.HasPrefix(file, gormSourceDir) || strings.HasSuffix(file, "_test.go")) {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}

func init() {
	_, file, _, _ := runtime.Caller(0)
	// compatible solution to get gorm source directory with various operating systems
	gormSourceDir = regexp.MustCompile(`gorm.utils.utils\.go`).ReplaceAllString(file, "")
}
