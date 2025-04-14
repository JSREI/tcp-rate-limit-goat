package sdk

// Factory 提供便捷的方法来创建各种类型的检测器
type Factory struct{}

// NewFactory 创建一个新的检测器工厂
func NewFactory() *Factory {
	return &Factory{}
}

// RateLimit 创建一个新的速率限制检测器，使用默认配置
func (f *Factory) RateLimit() *RateLimitDetector {
	return NewRateLimitDetector()
}

// RateLimitWithOptions 创建一个新的速率限制检测器，使用自定义配置
func (f *Factory) RateLimitWithOptions(options *RateLimitOptions) *RateLimitDetector {
	return NewRateLimitDetectorWithOptions(options)
}

// PortIncrement 创建一个新的端口递增检测器，使用默认配置
func (f *Factory) PortIncrement() *PortIncrementDetector {
	return NewPortIncrementDetector()
}

// PortIncrementWithOptions 创建一个新的端口递增检测器，使用自定义配置
func (f *Factory) PortIncrementWithOptions(options *PortIncrementOptions) *PortIncrementDetector {
	return NewPortIncrementDetectorWithOptions(options)
}

// Default 全局默认工厂实例
var Default = NewFactory()
