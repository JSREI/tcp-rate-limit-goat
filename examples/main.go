// 本示例展示了如何使用TCP检测SDK
// 包括速率限制检测器和端口递增检测器的基本用法和高级配置
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/JSREI/tcp-rate-limit-goat/pkg/sdk"
	"github.com/gin-gonic/gin"
)

// combinedMiddleware 组合使用多个检测器，提供更全面的保护
// 这是推荐的使用方式，可以检测多种可疑行为
func combinedMiddleware(c *gin.Context) {
	// 获取客户端的远程地址（IP:端口）
	remoteAddr := c.Request.RemoteAddr

	// 创建速率限制检测器（使用默认配置）
	rateDetector := sdk.Default.RateLimit()

	// 创建端口递增检测器（使用默认配置）
	portDetector := sdk.Default.PortIncrement()

	// 首先检查访问频率
	if rateDetector.Detect(remoteAddr) {
		// 输出样例: [速率限制] 拒绝访问: 127.0.0.1:52134
		fmt.Printf("[速率限制] 拒绝访问: %s\n", remoteAddr)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":     "访问频率超限",
			"code":      1001,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// 然后检查端口递增行为
	if portDetector.Detect(remoteAddr) {
		// 输出样例: [端口递增] 拒绝访问: 127.0.0.1:52134
		fmt.Printf("[端口递增] 拒绝访问: %s\n", remoteAddr)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":     "疑似扫描行为",
			"code":      1002,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	// 通过所有检测，允许访问
	// 输出样例: [通过] 允许访问: 127.0.0.1:52134
	fmt.Printf("[通过] 允许访问: %s\n", remoteAddr)
	c.Next()
}

// customRateLimitMiddleware 展示如何使用自定义配置的速率限制检测器
// 适用于需要对不同API路径设置不同访问限制的场景
func customRateLimitMiddleware(c *gin.Context) {
	remoteAddr := c.Request.RemoteAddr

	// 创建自定义配置
	options := sdk.DefaultRateLimitOptions()
	options.Limit = 5  // 每秒只允许5个请求（更严格的限制）
	options.Burst = 3  // 最多允许3个突发请求
	options.TTL = 1800 // 缓存生存时间减少到30分钟

	// 使用自定义配置创建检测器
	detector := sdk.Default.RateLimitWithOptions(options)

	// 执行检测
	if detector.Detect(remoteAddr) {
		// 输出样例: [自定义速率限制] 拒绝访问: 127.0.0.1:52134
		fmt.Printf("[自定义速率限制] 拒绝访问: %s\n", remoteAddr)
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":     "API访问频率超限",
			"code":      1003,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	c.Next()
}

// advancedPortMiddleware 展示端口递增检测器的高级用法
// 适用于对特定敏感API提供更严格的扫描防护
func advancedPortMiddleware(c *gin.Context) {
	remoteAddr := c.Request.RemoteAddr

	// 1. 创建具有默认配置的检测器
	detector := sdk.Default.PortIncrement()

	// 2. 动态调整配置（运行时修改）
	detector.SetCutoff(2)      // 降低阈值，只需连续2次端口递增即视为可疑
	detector.SetBufferSize(20) // 增加历史记录大小，记录更多的端口

	// 3. 执行检测
	if detector.Detect(remoteAddr) {
		// 输出样例: [高级端口检测] 拒绝访问: 127.0.0.1:52134
		fmt.Printf("[高级端口检测] 拒绝访问: %s\n", remoteAddr)

		// 记录详细日志（在实际应用中可以写入到日志文件）
		now := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("[%s] 可能的扫描行为: %s 访问敏感API\n", now, remoteAddr)

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error":     "检测到可疑行为",
			"code":      1004,
			"timestamp": time.Now().Unix(),
		})
		return
	}

	c.Next()
}

// 主函数，设置HTTP服务器并注册路由和中间件
func main() {
	// 创建Gin路由器，默认已包含Logger和Recovery中间件
	r := gin.Default()

	// 为所有路由注册组合检测中间件（全局防护）
	r.Use(combinedMiddleware)

	// 首页路由
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "欢迎使用TCP检测SDK示例程序！\n\n可用路由:\n- /api/* (标准保护)\n- /sensitive/* (增强保护)\n- /rate-test (自定义速率限制)")
	})

	// 标准API路由组
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/users", func(c *gin.Context) {
			// 模拟返回用户列表
			users := []gin.H{
				{"id": 1, "name": "张三"},
				{"id": 2, "name": "李四"},
				{"id": 3, "name": "王五"},
			}
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data":    users,
			})
		})

		apiGroup.GET("/products", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data":    "产品列表",
			})
		})
	}

	// 敏感API路由组（使用增强的端口检测）
	sensitiveGroup := r.Group("/sensitive")
	sensitiveGroup.Use(advancedPortMiddleware) // 添加高级端口检测
	{
		sensitiveGroup.GET("/admin", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data":    "管理员数据",
			})
		})

		sensitiveGroup.GET("/config", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data":    "系统配置",
			})
		})
	}

	// 自定义速率限制测试路由
	rateTestGroup := r.Group("/rate-test")
	rateTestGroup.Use(customRateLimitMiddleware) // 添加自定义速率限制
	{
		rateTestGroup.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"code":    0,
				"message": "success",
				"data":    "速率限制测试通过",
				"limit":   "5次/秒",
			})
		})
	}

	// 启动HTTP服务器
	fmt.Println("=== TCP检测SDK示例服务已启动 ===")
	fmt.Println("访问 http://localhost:8080 查看示例")
	fmt.Println("可以使用 ab -n 100 -c 10 http://localhost:8080/rate-test 测试速率限制")
	fmt.Println("=================================")
	r.Run(":8080")
}
