package grow

import "runtime"

// Option 配置 Plot 的可选参数。
type Option func(*plotOptions)

type plotOptions struct {
	numTends int
}

func defaultOptions() plotOptions {
	return plotOptions{
		numTends: runtime.NumCPU(),
	}
}

// WithTends 设置并发工作协程数。默认为 runtime.NumCPU()。
func WithTends(n int) Option {
	return func(o *plotOptions) {
		o.numTends = n
	}
}
