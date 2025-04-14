# TCP 检测 SDK

本SDK提供了基于TCP连接的访问检测功能，包括速率限制和端口递增检测。可用于防止网络爬虫、CC攻击和其他异常流量。

## 功能特点

- 速率限制检测：基于令牌桶算法，限制每个IP的访问频率
- 端口递增检测：检测连续递增的端口访问，识别潜在的端口扫描行为
- 高性能：使用LRU缓存优化内存使用
- 灵活配置：提供丰富的配置选项
- 简单集成：为Gin框架提供开箱即用的中间件

## 安装

```bash
go get github.com/JSREI/tcp-rate-limit-goat
```

## 快速开始

### 基本用法

```go
package main

import (
    "github.com/JSREI/tcp-rate-limit-goat/pkg/sdk"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    // 创建检测器
    rateDetector := sdk.Default.RateLimit()
    
    // 注册中间件
    r.Use(func(c *gin.Context) {
        if rateDetector.Detect(c.Request.RemoteAddr) {
            c.AbortWithStatus(403)
            return
        }
        c.Next()
    })
    
    r.GET("/", func(c *gin.Context) {
        c.String(200, "Hello World")
    })
    
    r.Run(":8080")
}
```

### 组合多个检测器

```go
// 创建检测器
rateDetector := sdk.Default.RateLimit()
portDetector := sdk.Default.PortIncrement()

// 中间件
r.Use(func(c *gin.Context) {
    remoteAddr := c.Request.RemoteAddr
    
    // 速率限制检测
    if rateDetector.Detect(remoteAddr) {
        c.AbortWithStatus(403)
        return
    }
    
    // 端口递增检测
    if portDetector.Detect(remoteAddr) {
        c.AbortWithStatus(403)
        return
    }
    
    c.Next()
})
```

### 自定义配置

```go
// 自定义速率限制配置
rateOptions := sdk.DefaultRateLimitOptions()
rateOptions.Limit = 10  // 每秒10个请求
rateOptions.Burst = 5   // 最多5个突发请求
rateDetector := sdk.Default.RateLimitWithOptions(rateOptions)

// 自定义端口递增检测配置
portOptions := sdk.DefaultPortIncrementOptions()
portOptions.Cutoff = 5  // 连续5次端口递增视为可疑
portDetector := sdk.Default.PortIncrementWithOptions(portOptions)
```

## 高级用法

### 动态调整配置

```go
// 动态调整速率限制
rateDetector := sdk.Default.RateLimit()
rateDetector.SetLimit(30)   // 调整为每秒30个请求
rateDetector.SetBurst(10)   // 调整为最多10个突发请求

// 动态调整端口递增检测
portDetector := sdk.Default.PortIncrement()
portDetector.SetCutoff(10)      // 调整为连续10次递增视为可疑
portDetector.SetBufferSize(20)  // 调整为记录最近20个端口
```

### 完整示例

请查看 [examples](./examples) 目录获取完整的使用示例。

## API 参考

### 接口

- `Detector`: 所有检测器的通用接口
  - `Detect(remoteAddress string) bool`: 检测方法，返回true表示检测到异常

### 检测器

- `RateLimitDetector`: 速率限制检测器
  - `NewRateLimitDetector()`: 创建默认配置的速率限制检测器
  - `NewRateLimitDetectorWithOptions(options *RateLimitOptions)`: 创建自定义配置的速率限制检测器
  - `SetLimit(limit rate.Limit)`: 设置速率限制
  - `SetBurst(burst int)`: 设置突发请求限制

- `PortIncrementDetector`: 端口递增检测器
  - `NewPortIncrementDetector()`: 创建默认配置的端口递增检测器
  - `NewPortIncrementDetectorWithOptions(options *PortIncrementOptions)`: 创建自定义配置的端口递增检测器
  - `SetCutoff(cutoff int)`: 设置递增阈值
  - `SetBufferSize(size int)`: 设置历史记录大小

### 工厂

- `Factory`: 提供创建检测器的便捷方法
  - `RateLimit()`: 创建默认速率限制检测器
  - `RateLimitWithOptions(options *RateLimitOptions)`: 创建自定义速率限制检测器
  - `PortIncrement()`: 创建默认端口递增检测器
  - `PortIncrementWithOptions(options *PortIncrementOptions)`: 创建自定义端口递增检测器

- `Default`: 全局默认工厂实例

## 许可证

MIT 