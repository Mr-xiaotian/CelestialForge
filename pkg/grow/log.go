package grow

import (
	"fmt"
	"os"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/funnel"
)

var levelOrder = map[string]int{
	"TRACE":   0,
	"DEBUG":   1,
	"INFO":    2,
	"SUCCESS": 3,
	"WARNING": 4,
	"ERROR":   5,
}

// ==== Record ====

// LogRecord 日志记录条目。
type LogRecord struct {
	FormatTime string
	Level      string
	Message    string
}

// ==== Spout (RecordHandler) ====

// LogRecordHandler 日志记录的 RecordHandler 实现，将日志写入文本文件。
type LogRecordHandler struct {
	LogPath string
	logFile *os.File
}

// BeforeStart 创建日志目录并打开日志文件。
func (l *LogRecordHandler) BeforeStart() error {
	var err error

	today := time.Now().Format("2006-01-02")
	l.LogPath = fmt.Sprintf("logs/grow_log(%s).log", today)
	if err = os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	l.logFile, err = os.OpenFile(l.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	return nil
}

// HandleRecord 将日志记录格式化后追加写入文件。
func (l *LogRecordHandler) HandleRecord(record LogRecord) error {
	line := fmt.Sprintf("%s %s %s\n", record.FormatTime, record.Level, record.Message)

	_, err := l.logFile.WriteString(line)
	if err != nil {
		return fmt.Errorf("写入日志文件失败: %w", err)
	}

	return nil
}

// AfterStop 关闭日志文件。
func (l *LogRecordHandler) AfterStop() error {
	err := l.logFile.Close()
	if err != nil {
		return fmt.Errorf("关闭日志文件失败: %w", err)
	}

	return nil
}

// ==== Inlet ====

// LogInlet 日志生产端，内嵌 Inlet 并提供级别过滤和领域方法。
// 低于 minLevel 的日志不会被发送到通道。
type LogInlet struct {
	funnel.Inlet[LogRecord]
	minLevel int
}

// NewLogInlet 创建 LogInlet，level 指定最低日志级别（不存在则默认 INFO）。
func NewLogInlet(ch chan<- LogRecord, timeout time.Duration, level string) *LogInlet {
	minLevel, ok := levelOrder[level]
	if !ok {
		minLevel = levelOrder["INFO"]
	}
	return &LogInlet{
		Inlet:    *funnel.NewInlet(ch, timeout),
		minLevel: minLevel,
	}
}

// log 发送一条日志。级别低于 minLevel 的日志会被静默丢弃。
func (l *LogInlet) log(level string, message string) {
	if levelOrder[level] < l.minLevel {
		return
	}
	l.Send(LogRecord{
		FormatTime: time.Now().Format("2006-01-02 15:04:05"),
		Level:      level,
		Message:    message,
	})
}

// StartPlot 记录 Plot 启动日志。
func (l *LogInlet) StartPlot(plotName string, numTends int) {
	l.log("INFO", fmt.Sprintf("'%s' start by %d tends.", plotName, numTends))
}

// EndPlot 记录 Plot 结束日志。
func (l *LogInlet) EndPlot(plotName string, useTime float64, successNum, failedNum int) {
	l.log("INFO", fmt.Sprintf("'%s' end. Use %.2fs. %d success, %d failed.", plotName, useTime, successNum, failedNum))
}

// TendSuccess 记录种子培育成功日志。
func (l *LogInlet) TendSuccess(plotName string, seedRepr string, fruitRepr string, useTime float64) {
	l.log("SUCCESS", fmt.Sprintf("In '%s', Seed %s successed. Fruit is %s. Use %.2fs.", plotName, seedRepr, fruitRepr, useTime))
}

// TendRetry 记录种子培育重试日志。
func (l *LogInlet) TendRetry(plotName string, seedRepr string, attempt int, err error) {
	l.log("WARNING", fmt.Sprintf("In '%s', Seed %s attempt %d failed: %v. Retrying...", plotName, seedRepr, attempt, err))
}

// TendFail 记录种子培育失败日志。
func (l *LogInlet) TendFail(plotName string, seedRepr string, err error) {
	l.log("ERROR", fmt.Sprintf("In '%s', Seed %s failed: %v.", plotName, seedRepr, err))
}
