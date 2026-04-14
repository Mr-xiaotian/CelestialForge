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

type LogRecord struct {
	FormatTime string
	Level      string
	Message    string
}

// LogRecordHandler 处理日志记录的接口
type LogRecordHandler struct {
	LogPath string
	logFile *os.File
}

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

func (l *LogRecordHandler) HandleRecord(record LogRecord) error {
	line := fmt.Sprintf("%s %s %s\n", record.FormatTime, record.Level, record.Message)

	_, err := l.logFile.WriteString(line)
	if err != nil {
		return fmt.Errorf("写入日志文件失败: %w", err)
	}

	return nil
}

func (l *LogRecordHandler) AfterStop() error {
	err := l.logFile.Close()
	if err != nil {
		return fmt.Errorf("关闭日志文件失败: %w", err)
	}

	return nil
}

// LogInlet 日志生产端，内嵌 Inlet 并提供级别过滤和领域方法
type LogInlet struct {
	funnel.Inlet[LogRecord]
	minLevel int
}

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

func (l *LogInlet) StartExecutor(executorName string, numWorkers int) {
	l.log("INFO", fmt.Sprintf("'%s' start by %d workers.", executorName, numWorkers))
}

func (l *LogInlet) EndExecutor(executorName string, useTime float64, successNum, failedNum int) {
	l.log("INFO", fmt.Sprintf("'%s' end. Use %.2fs. %d success, %d failed.", executorName, useTime, successNum, failedNum))
}

func (l *LogInlet) TaskSuccess(executorName string, taskRepr string, resultRepr string, useTime float64) {
	l.log("SUCCESS", fmt.Sprintf("In '%s', %s successed. Result is %s. Use %.2fs.", executorName, taskRepr, resultRepr, useTime))
}

func (l *LogInlet) TaskError(executorName string, taskRepr string, err error) {
	l.log("ERROR", fmt.Sprintf("In '%s', %s failed: %v.", executorName, taskRepr, err))
}
