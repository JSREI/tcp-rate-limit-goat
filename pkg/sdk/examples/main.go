// 本示例展示了如何使用TCP检测SDK
// 包括速率限制检测器和端口递增检测器的基本用法
package main

import (
	"fmt"
	"net/http"

	"github.com/JSREI/tcp-rate-limit-goat/pkg/sdk"
	"github.com/gin-gonic/gin"
)

// 组合多个检测器，提供更全面的检测
func combinedMiddleware(c *gin.Context) {
	remoteAddr := c.Request.RemoteAddr

	// 创建速率限制检测器
	rateDetector := sdk.Default.RateLimit()

	// 创建端口递增检测器
	portDetector := sdk.Default.PortIncrement()

	// 两个检测器都通过才允许访问
	if rateDetector.Detect(remoteAddr) {
		fmt.Println("速率限制：拒绝访问", remoteAddr)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "访问频率超限",
		})
		return
	}

	if portDetector.Detect(remoteAddr) {
		fmt.Println("端口递增：拒绝访问", remoteAddr)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "疑似扫描行为",
		})
		return
	}

	fmt.Println("允许访问:", remoteAddr)
	c.Next()
}

// 只使用速率限制检测器
func rateMiddleware(c *gin.Context) {
	remoteAddr := c.Request.RemoteAddr

	// 使用默认配置创建速率限制检测器
	detector := sdk.Default.RateLimit()

	if detector.Detect(remoteAddr) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Next()
}

// 只使用端口递增检测器
func portMiddleware(c *gin.Context) {
	remoteAddr := c.Request.RemoteAddr

	// 使用自定义配置创建端口递增检测器
	options := sdk.DefaultPortIncrementOptions()
	options.Cutoff = 5 // 设置更高的阈值
	detector := sdk.Default.PortIncrementWithOptions(options)

	if detector.Detect(remoteAddr) {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	c.Next()
}

func main() {
	r := gin.Default()

	// 注册全局中间件 - 使用组合检测器
	r.Use(combinedMiddleware)

	// 简单的欢迎页面
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "欢迎使用TCP检测SDK示例程序！")
	})

	// 仅使用速率限制检测的路由组
	rateGroup := r.Group("/rate")
	rateGroup.Use(rateMiddleware)
	{
		rateGroup.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "速率限制检测通过！")
		})
	}

	// 仅使用端口递增检测的路由组
	portGroup := r.Group("/port")
	portGroup.Use(portMiddleware)
	{
		portGroup.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "端口递增检测通过！")
		})
	}

	// 启动服务器
	r.Run(":8080")
}
