// Package sdk 提供TCP访问频率和端口递增检测的核心SDK功能
// 该SDK可以用于检测网络爬虫、CC攻击或其他异常的网络访问行为
package sdk

// Detector 定义了通用的检测器接口
// 所有具体的检测器实现都应实现该接口
type Detector interface {
	// Detect 检查给定的远程地址是否符合检测条件
	// 参数：
	//   - remoteAddress: 远程地址，通常是"IP:端口"格式
	// 返回值：
	//   - bool: true表示检测到异常，false表示正常
	Detect(remoteAddress string) bool
}

// DetectorOptions 定义了检测器的通用配置选项
type DetectorOptions struct {
	// Capacity 缓存容量
	Capacity int

	// TTL 缓存条目的生存时间（秒）
	TTL int
}

// DefaultDetectorOptions 返回默认的检测器配置
func DefaultDetectorOptions() *DetectorOptions {
	return &DetectorOptions{
		Capacity: 1000000, // 默认缓存100万条记录
		TTL:      3600,    // 默认缓存生存时间1小时
	}
}
