package grow

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/funnel"
)

// ==== Record ====

// FailRecord 单条失败记录，序列化为 JSONL 写入文件。
type FailRecord struct {
	FormatTime   string
	PlotName     string
	SeedString   string
	ErrorMessage string
}

// ==== Record Handler (Spout) ====

// FailRecordHandler 失败记录的消费端处理器。
// 实现 funnel.RecordHandler[FailRecord] 接口，将失败信息以 JSONL 格式写入文件。
type FailRecordHandler struct {
	FailPath string
	FailFile *os.File
}

// BeforeStart 创建 fallback/{date}/ 目录并打开 JSONL 文件（追加模式）。
func (f *FailRecordHandler) BeforeStart() error {
	var err error

	today := time.Now().Format("2006-01-02")
	now := time.Now().Format("15-04-05.000")
	f.FailPath = fmt.Sprintf("fallback/%s/grow_fail(%s).jsonl", today, now)
	if err = os.MkdirAll(fmt.Sprintf("fallback/%s", today), 0755); err != nil {
		return fmt.Errorf("创建失败记录目录失败: %w", err)
	}

	f.FailFile, err = os.OpenFile(f.FailPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开失败记录文件失败: %w", err)
	}

	return nil
}

// HandleRecord 将失败记录序列化为 JSON 并追加写入文件（每条一行）。
func (f *FailRecordHandler) HandleRecord(record FailRecord) error {
	var err error
	var b []byte

	b, err = json.Marshal(record)
	if err != nil {
		return fmt.Errorf("序列化失败记录失败: %w", err)
	}

	_, err = f.FailFile.Write(append(b, '\n'))
	if err != nil {
		return fmt.Errorf("写入失败记录文件失败: %w", err)
	}

	return nil
}

// AfterStop 关闭 JSONL 文件句柄。
func (f *FailRecordHandler) AfterStop() error {
	err := f.FailFile.Close()
	if err != nil {
		return fmt.Errorf("关闭失败记录文件失败: %w", err)
	}

	return nil
}

// ==== Inlet ====

// FailInlet 失败记录生产端。内嵌 funnel.Inlet，提供面向 Plot 的失败记录方法。
type FailInlet struct {
	funnel.Inlet[FailRecord]
}

// NewFailInlet 创建 FailInlet，绑定到指定的写通道。
func NewFailInlet(ch chan<- FailRecord, timeout time.Duration) *FailInlet {
	return &FailInlet{
		Inlet: *funnel.NewInlet(ch, timeout),
	}
}

// ==== Fail Methods ====

// SeedWither 发送一条种子培育失败记录，包含 plot 名称、种子值和错误信息。
func (f *FailInlet) SeedWither(plotName string, seedString string, err error) {
	f.Send(FailRecord{
		FormatTime:   time.Now().Format("2006-01-02 15:04:05"),
		PlotName:     plotName,
		SeedString:   seedString,
		ErrorMessage: fmt.Sprintf("%v", err),
	})
}
