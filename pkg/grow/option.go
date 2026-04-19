package grow

import (
	"runtime"
	"time"
)

// Option 配置 Plot 的可选参数。
type Option func(*plotOptions)

type plotOptions struct {
	numTends   int
	maxRetries int
	retryDelay func(attempt int) time.Duration
	retryIf    func(error) bool
	logLevel   string
}

func defaultOptions() plotOptions {
	return plotOptions{
		numTends:   runtime.NumCPU(),
		maxRetries: 1,
		retryDelay: func(attempt int) time.Duration { return 0 },
		retryIf:    func(error) bool { return true },
		logLevel:   "INFO",
	}
}

// WithTends 设置并发工作协程数。默认为 runtime.NumCPU()。
func WithTends(n int) Option {
	return func(o *plotOptions) {
		o.numTends = n
	}
}

// WithMaxRetries 设置最大重试次数。默认为 1, 执行 2 次（1 次原始 + 1 次重试）。
func WithMaxRetries(n int) Option {
	return func(o *plotOptions) {
		o.maxRetries = n
	}
}

// WithRetryDelay 设置重试间隔策略。attempt 从 1 开始。
func WithRetryDelay(fn func(attempt int) time.Duration) Option {
	return func(o *plotOptions) {
		o.retryDelay = fn
	}
}

// WithRetryIf 设置哪些错误值得重试。返回 true 则重试。默认全部重试。
func WithRetryIf(fn func(error) bool) Option {
	return func(o *plotOptions) {
		o.retryIf = fn
	}
}

// WithLogLevel 设置日志级别。默认为 "INFO"。
func WithLogLevel(level string) Option {
	return func(o *plotOptions) {
		o.logLevel = level
	}
}
