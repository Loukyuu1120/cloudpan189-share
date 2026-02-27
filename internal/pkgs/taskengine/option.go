package taskengine

import "time"

type Options struct {
	WorkerCount    int           // 工作协程数量，1为单线程
	BufferSize     int           // 消息缓冲区大小
	ProcessTimeout time.Duration // 单个消息处理超时
	MaxRetry       int           // 默认最大重试次数
	RetryDelay     time.Duration // 重试延迟
	EnableStats    bool          // 是否启用统计
}

func defaultOptions() *Options {
	return &Options{
		WorkerCount:    4,
		BufferSize:     89120,
		ProcessTimeout: time.Minute * 30,
		MaxRetry:       3,
		RetryDelay:     time.Second * 1,
		EnableStats:    true,
	}
}

// OptionFunc 选项函数类型
type OptionFunc func(*Options)

// WithWorkerCount 设置工作协程数量
func WithWorkerCount(count int) OptionFunc {
	return func(o *Options) {
		if count > 0 {
			o.WorkerCount = count
		}
	}
}

// WithBufferSize 设置缓冲区大小
func WithBufferSize(size int) OptionFunc {
	return func(o *Options) {
		if size > 0 {
			o.BufferSize = size
		}
	}
}

// WithProcessTimeout 设置处理超时
func WithProcessTimeout(timeout time.Duration) OptionFunc {
	return func(o *Options) {
		if timeout > 0 {
			o.ProcessTimeout = timeout
		}
	}
}

// WithMaxRetry 设置最大重试次数
func WithMaxRetry(maxRetry int) OptionFunc {
	return func(o *Options) {
		if maxRetry >= 0 {
			o.MaxRetry = maxRetry
		}
	}
}

// WithRetryDelay 设置重试延迟
func WithRetryDelay(delay time.Duration) OptionFunc {
	return func(o *Options) {
		if delay >= 0 {
			o.RetryDelay = delay
		}
	}
}

// WithStats 设置是否启用统计
func WithStats(enable bool) OptionFunc {
	return func(o *Options) {
		o.EnableStats = enable
	}
}
