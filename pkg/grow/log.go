package grow

import (
	"context"
	"fmt"
	"os"
	"time"
)

var levelOrder = map[string]int{
	"TRACE":   0,
	"DEBUG":   1,
	"INFO":    2,
	"SUCCESS": 3,
	"WARNING": 4,
	"ERROR":   5,
}

type LogRecord struct {
	Timestamp string
	Level     string
	Message   string
}

// LogRecordHandler 处理日志记录的接口
type LogRecordHandler struct {
	LogPath string
	logFile *os.File
}

func (h *LogRecordHandler) BeforeStart() error {
	var err error

	today := time.Now().Format("2006-01-02")
	h.LogPath = fmt.Sprintf("logs/grow_log(%s).log", today)
	if err = os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	h.logFile, err = os.OpenFile(h.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	return nil
}

func (h *LogRecordHandler) HandleRecord(record LogRecord) error {
	line := fmt.Sprintf("%s %s %s\n", record.Timestamp, record.Level, record.Message)

	_, err := h.logFile.WriteString(line)
	if err != nil {
		return fmt.Errorf("写入日志文件失败: %w", err)
	}

	return nil
}

func (h *LogRecordHandler) AfterStop() error {
	err := h.logFile.Close()
	if err != nil {
		return fmt.Errorf("关闭日志文件失败: %w", err)
	}

	return nil
}

// LogSource 日志生产端，内嵌 Source 并提供级别过滤和领域方法
type LogSource struct {
	ch       chan<- LogRecord
	ctx      context.Context
	cancel   context.CancelFunc
	timeout  time.Duration
	minLevel int
}

func NewLogSource(ch chan<- LogRecord, timeout time.Duration, level string) *LogSource {
	ctx, cancel := context.WithCancel(context.Background())
	minLevel, ok := levelOrder[level]
	if !ok {
		minLevel = levelOrder["INFO"]
	}
	return &LogSource{
		ch: ch, ctx: ctx, cancel: cancel, timeout: timeout, minLevel: minLevel,
	}
}

// Send 发送记录，支持上下文取消和超时控制
func (s *LogSource) Send(record LogRecord) error {
	select {
	case s.ch <- record:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	case <-time.After(s.timeout):
		return fmt.Errorf("source send timeout after %v", s.timeout)
	}
}

// Close 关闭 Source，后续 Send 调用将返回错误
func (s *LogSource) Close() {
	s.cancel()
}

func (l *LogSource) log(level string, message string) {
	if levelOrder[level] < l.minLevel {
		return
	}
	l.Send(LogRecord{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Level:     level,
		Message:   message,
	})
}

func (l *LogSource) StartExecutor(executorName string, numWorkers int) {
	l.log("INFO", fmt.Sprintf("'%s' start by %d workers.", executorName, numWorkers))
}

func (l *LogSource) EndExecutor(executorName string, useTime float64, successNum, failedNum int) {
	l.log("INFO", fmt.Sprintf("'%s' end. Use %.2fs. %d success, %d failed.", executorName, useTime, successNum, failedNum))
}

func (l *LogSource) TaskSuccess(executorName string, taskRepr string, resultRepr string, useTime float64) {
	l.log("SUCCESS", fmt.Sprintf("In '%s', %s successed. Result is %s. Use %.2fs.", executorName, taskRepr, resultRepr, useTime))
}

func (l *LogSource) TaskError(executorName string, taskRepr string, err error) {
	l.log("ERROR", fmt.Sprintf("In '%s', %s failed: %v.", executorName, taskRepr, err))
}
