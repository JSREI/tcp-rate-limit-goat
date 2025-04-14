package sdk

import (
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/time/rate"
)

// RateLimitOptions 定义速率限制检测器的专用配置选项
type RateLimitOptions struct {
	// 继承通用检测器选项
	*DetectorOptions

	// Limit 每秒允许的请求数
	Limit rate.Limit

	// Burst 允许的突发请求数
	Burst int
}

// DefaultRateLimitOptions 返回速率限制检测器的默认配置
func DefaultRateLimitOptions() *RateLimitOptions {
	return &RateLimitOptions{
		DetectorOptions: DefaultDetectorOptions(),
		Limit:           20, // 默认每秒20个请求
		Burst:           20, // 默认允许20个突发请求
	}
}

// RateLimitDetector 实现基于TCP连接频率的访问限制检测
// 使用令牌桶算法限制每个IP:端口的访问频率
type RateLimitDetector struct {
	// 保护并发访问的互斥锁
	lock *sync.Mutex

	// 缓存每个远程地址的限速器
	cache *expirable.LRU[string, *rate.Limiter]

	// 每秒允许的请求数
	limit rate.Limit

	// 允许的突发请求数
	burst int
}

// NewRateLimitDetector 创建一个新的速率限制检测器实例
// 使用默认配置选项
func NewRateLimitDetector() *RateLimitDetector {
	return NewRateLimitDetectorWithOptions(DefaultRateLimitOptions())
}

// NewRateLimitDetectorWithOptions 使用自定义配置选项创建速率限制检测器
func NewRateLimitDetectorWithOptions(options *RateLimitOptions) *RateLimitDetector {
	ttl := time.Duration(options.TTL) * time.Second
	return &RateLimitDetector{
		lock:  &sync.Mutex{},
		cache: expirable.NewLRU[string, *rate.Limiter](options.Capacity, nil, ttl),
		limit: options.Limit,
		burst: options.Burst,
	}
}

// Detect 实现Detector接口
// 检查给定的远程地址是否超过了速率限制
// 如果允许访问返回false，如果拒绝访问返回true
func (d *RateLimitDetector) Detect(remoteAddress string) bool {
	d.lock.Lock()
	defer d.lock.Unlock()

	// 从缓存中获取该地址的限速器
	limiter, exists := d.cache.Get(remoteAddress)
	if !exists {
		// 如果不存在，创建新的限速器
		limiter = rate.NewLimiter(d.limit, d.burst)
		d.cache.Add(remoteAddress, limiter)
	}

	// 使用令牌桶算法检查是否允许访问
	// 返回结果取反，因为Allow()为true表示允许访问，
	// 而我们的Detect()为true表示检测到异常
	return !limiter.Allow()
}

// SetLimit 更新速率限制值
func (d *RateLimitDetector) SetLimit(limit rate.Limit) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.limit = limit
}

// SetBurst 更新突发请求限制值
func (d *RateLimitDetector) SetBurst(burst int) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.burst = burst
}
