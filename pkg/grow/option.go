package grow

import "runtime"

// Option 配置 Plot 的可选参数。
type Option func(*plotOptions)

type plotOptions struct {
	numWorkers int
}

func defaultOptions() plotOptions {
	return plotOptions{
		numWorkers: runtime.NumCPU(),
	}
}

// WithWorkers 设置并发工作协程数。默认为 runtime.NumCPU()。
func WithWorkers(n int) Option {
	return func(o *plotOptions) {
		o.numWorkers = n
	}
}
