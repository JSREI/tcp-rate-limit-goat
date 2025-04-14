// 这个示例展示了TCP检测SDK的最简单用法
// 适合初学者了解核心功能
package main

import (
	"fmt"
	"net/http"

	"github.com/JSREI/tcp-rate-limit-goat/pkg/sdk"
	"github.com/gin-gonic/gin"
)

func main() {
	// 设置控制台输出
	fmt.Println("=== TCP检测SDK简单示例 ===")
	fmt.Println("启动一个简单的Web服务器，展示速率限制功能")
	fmt.Println("===============================")

	// 1. 创建Gin路由器
	r := gin.Default()

	// 2. 创建速率限制检测器 - 使用全局默认工厂
	rateDetector := sdk.Default.RateLimit()

	// 3. 添加速率限制中间件
	r.Use(func(c *gin.Context) {
		remoteAddr := c.Request.RemoteAddr

		// 检查是否超过速率限制
		if rateDetector.Detect(remoteAddr) {
			// 输出样例: [拒绝] 客户端127.0.0.1:54321访问频率过高
			fmt.Printf("[拒绝] 客户端%s访问频率过高\n", remoteAddr)

			// 返回403状态码和错误信息
			c.JSON(http.StatusForbidden, gin.H{
				"status":  "error",
				"message": "访问频率超限，请稍后再试",
			})

			// 中止后续处理
			c.Abort()
			return
		}

		// 输出样例: [通过] 允许客户端127.0.0.1:54321访问
		fmt.Printf("[通过] 允许客户端%s访问\n", remoteAddr)

		// 继续处理请求
		c.Next()
	})

	// 4. 添加示例路由
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "欢迎使用TCP检测SDK！\n\n这是一个简单的示例，默认限制每秒20个请求。\n尝试快速刷新页面测试速率限制功能。")
	})

	// 5. 启动服务器
	fmt.Println("服务器已启动，访问 http://localhost:8081")
	fmt.Println("提示: 快速刷新页面可以触发速率限制")
	r.Run(":8081")
}
