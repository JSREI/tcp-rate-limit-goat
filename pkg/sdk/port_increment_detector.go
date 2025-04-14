package sdk

import (
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/JSREI/tcp-rate-limit-goat/pkg/utils"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

// PortIncrementOptions 定义端口递增检测器的专用配置选项
type PortIncrementOptions struct {
	// 继承通用检测器选项
	*DetectorOptions

	// Cutoff 设置端口递增阈值，超过此值视为可疑行为
	Cutoff int

	// BufferSize 端口历史记录的缓冲区大小
	BufferSize int
}

// DefaultPortIncrementOptions 返回端口递增检测器的默认配置
func DefaultPortIncrementOptions() *PortIncrementOptions {
	return &PortIncrementOptions{
		DetectorOptions: DefaultDetectorOptions(),
		Cutoff:          3,  // 默认连续递增3次视为可疑
		BufferSize:      10, // 默认记录最近10个端口
	}
}

// PortIncrementDetector 实现基于IP端口递增的访问检测
// 用于检测网络扫描、端口探测等可疑行为
type PortIncrementDetector struct {
	// 保护并发访问的互斥锁
	lock *sync.Mutex

	// 缓存每个IP的端口历史记录
	cache *expirable.LRU[string, *utils.CircularBuffer]

	// 端口递增阈值
	cutoff int

	// 历史端口记录的大小
	bufferSize int
}

// NewPortIncrementDetector 创建一个新的端口递增检测器实例
// 使用默认配置选项
func NewPortIncrementDetector() *PortIncrementDetector {
	return NewPortIncrementDetectorWithOptions(DefaultPortIncrementOptions())
}

// NewPortIncrementDetectorWithOptions 使用自定义配置选项创建端口递增检测器
func NewPortIncrementDetectorWithOptions(options *PortIncrementOptions) *PortIncrementDetector {
	ttl := time.Duration(options.TTL) * time.Second
	return &PortIncrementDetector{
		lock:       &sync.Mutex{},
		cache:      expirable.NewLRU[string, *utils.CircularBuffer](options.Capacity, nil, ttl),
		cutoff:     options.Cutoff,
		bufferSize: options.BufferSize,
	}
}

// Detect 实现Detector接口
// 检查给定的远程地址是否存在端口递增行为
// 返回true表示检测到端口递增行为
func (d *PortIncrementDetector) Detect(remoteAddress string) bool {
	// 解析IP和端口
	ip, port, err := d.parseIPPort(remoteAddress)
	if err != nil {
		return false // 解析失败视为正常
	}

	d.lock.Lock()
	defer d.lock.Unlock()

	// 获取或创建该IP的端口历史记录
	var buffer *utils.CircularBuffer
	existingBuffer, exists := d.cache.Get(ip)
	if !exists {
		buffer = utils.NewCircularBuffer(d.bufferSize)
		d.cache.Add(ip, buffer)
	} else {
		buffer = existingBuffer
	}

	// 添加当前端口到历史记录
	buffer.Enqueue(port)

	// 检查是否有连续递增的端口
	return d.checkPortIncrement(buffer)
}

// parseIPPort 从远程地址中解析IP和端口
func (d *PortIncrementDetector) parseIPPort(remoteAddress string) (string, int, error) {
	host, portStr, err := net.SplitHostPort(remoteAddress)
	if err != nil {
		return "", 0, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}

	return host, port, nil
}

// checkPortIncrement 检查端口历史记录中是否存在连续递增模式
func (d *PortIncrementDetector) checkPortIncrement(buffer *utils.CircularBuffer) bool {
	elements := buffer.Elements()
	if len(elements) < 2 {
		return false // 至少需要2个端口才能判断
	}

	// 计算连续递增的次数
	incrementCount := 0
	for i := 1; i < len(elements); i++ {
		prev, ok1 := elements[i-1].(int)
		curr, ok2 := elements[i].(int)

		if !ok1 || !ok2 {
			continue
		}

		if curr == prev+1 {
			incrementCount++
			// 如果连续递增次数达到阈值，返回true
			if incrementCount >= d.cutoff-1 {
				return true
			}
		} else {
			// 重置计数器
			incrementCount = 0
		}
	}

	return false
}

// SetCutoff 更新端口递增阈值
func (d *PortIncrementDetector) SetCutoff(cutoff int) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.cutoff = cutoff
}

// SetBufferSize 更新端口历史记录的缓冲区大小
func (d *PortIncrementDetector) SetBufferSize(size int) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.bufferSize = size
}
