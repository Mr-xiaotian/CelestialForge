package grow

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
)

type Observer interface {
	OnStart(total int)
	OnProgress(completed, total int)
	OnFinish(completed, total int)
}

type ProgressBar struct {
	description string
	bar         *progressbar.ProgressBar
	mu          sync.Mutex
}

func NewProgressBar(description string) *ProgressBar {
	return &ProgressBar{description: description}
}

func (p *ProgressBar) ensureBar(total int) {
	if p.bar != nil {
		return
	}
	p.bar = progressbar.NewOptions64(
		int64(total),
		progressbar.OptionSetDescription(p.description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(10),
		progressbar.OptionShowTotalBytes(true),
		progressbar.OptionThrottle(time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)
}

func (p *ProgressBar) OnStart(total int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ensureBar(total)
}

func (p *ProgressBar) OnProgress(completed, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ensureBar(total)
	_ = p.bar.Set(completed)
}

func (p *ProgressBar) OnFinish(completed, total int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ensureBar(total)
	_ = p.bar.Set(total)
	_ = p.bar.Finish()
}
