// 这个示例展示了TCP检测SDK的高级用法
// 包括端口递增检测和动态配置
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/JSREI/tcp-rate-limit-goat/pkg/sdk"
	"github.com/gin-gonic/gin"
)

// 创建一个全局状态管理器，用于模拟动态调整配置
type DetectorManager struct {
	// 速率限制检测器
	rateDetector *sdk.RateLimitDetector

	// 端口递增检测器
	portDetector *sdk.PortIncrementDetector

	// 当前限制模式
	limitMode string
}

// 创建一个新的检测器管理器
func NewDetectorManager() *DetectorManager {
	return &DetectorManager{
		rateDetector: sdk.Default.RateLimit(),
		portDetector: sdk.Default.PortIncrement(),
		limitMode:    "normal", // 初始为普通模式
	}
}

// 应用当前的限制模式
func (m *DetectorManager) ApplyLimitMode(mode string) {
	m.limitMode = mode

	switch mode {
	case "normal": // 普通模式 - 默认配置
		m.rateDetector.SetLimit(20)
		m.rateDetector.SetBurst(20)
		m.portDetector.SetCutoff(3)
		fmt.Println("[配置] 已切换到普通保护模式 - 每秒限制20个请求")

	case "strict": // 严格模式 - 更严格的限制
		m.rateDetector.SetLimit(5)
		m.rateDetector.SetBurst(2)
		m.portDetector.SetCutoff(2)
		fmt.Println("[配置] 已切换到严格保护模式 - 每秒限制5个请求")

	case "relaxed": // 宽松模式 - 更宽松的限制
		m.rateDetector.SetLimit(50)
		m.rateDetector.SetBurst(30)
		m.portDetector.SetCutoff(5)
		fmt.Println("[配置] 已切换到宽松保护模式 - 每秒限制50个请求")
	}
}

func main() {
	// 创建检测器管理器
	manager := NewDetectorManager()

	// 设置控制台输出
	fmt.Println("=== TCP检测SDK高级示例 ===")
	fmt.Println("演示检测器的高级功能和动态配置")
	fmt.Println("============================")

	// 创建Gin路由器
	r := gin.Default()

	// 添加全局中间件
	r.Use(func(c *gin.Context) {
		remoteAddr := c.Request.RemoteAddr
		startTime := time.Now()

		// 速率限制检测
		if manager.rateDetector.Detect(remoteAddr) {
			// 输出样例: [速率限制] IP:127.0.0.1:12345 模式:strict 结果:拒绝
			fmt.Printf("[速率限制] IP:%s 模式:%s 结果:拒绝\n",
				remoteAddr, manager.limitMode)

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "访问频率超限",
				"mode":        manager.limitMode,
				"retry_after": 5, // 建议5秒后重试
			})
			return
		}

		// 端口递增检测
		if manager.portDetector.Detect(remoteAddr) {
			// 输出样例: [端口检测] IP:127.0.0.1 模式:normal 结果:拒绝
			fmt.Printf("[端口检测] IP:%s 模式:%s 结果:拒绝\n",
				remoteAddr, manager.limitMode)

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "检测到可疑的端口递增行为",
				"mode":  manager.limitMode,
			})
			return
		}

		// 继续处理请求
		c.Next()

		// 请求完成后记录处理时间
		// 输出样例: [请求] IP:127.0.0.1:12345 路径:/api/users 耗时:15.2ms
		elapsed := time.Since(startTime).Milliseconds()
		fmt.Printf("[请求] IP:%s 路径:%s 耗时:%dms\n",
			remoteAddr, c.Request.URL.Path, elapsed)
	})

	// 首页路由
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "TCP检测SDK高级示例\n\n"+
			"当前模式: "+manager.limitMode+"\n\n"+
			"可用路由:\n"+
			"- /mode/normal  切换到普通模式\n"+
			"- /mode/strict  切换到严格模式\n"+
			"- /mode/relaxed 切换到宽松模式\n"+
			"- /api/test     测试API\n")
	})

	// 模式切换路由
	modeGroup := r.Group("/mode")
	{
		modeGroup.GET("/normal", func(c *gin.Context) {
			manager.ApplyLimitMode("normal")
			c.String(http.StatusOK, "已切换到普通保护模式")
		})

		modeGroup.GET("/strict", func(c *gin.Context) {
			manager.ApplyLimitMode("strict")
			c.String(http.StatusOK, "已切换到严格保护模式")
		})

		modeGroup.GET("/relaxed", func(c *gin.Context) {
			manager.ApplyLimitMode("relaxed")
			c.String(http.StatusOK, "已切换到宽松保护模式")
		})
	}

	// 测试API路由
	r.GET("/api/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "success",
			"mode":      manager.limitMode,
			"timestamp": time.Now().Unix(),
			"message":   "API访问成功",
		})
	})

	// 启动服务器
	fmt.Println("服务器已启动，访问 http://localhost:8082")
	fmt.Println("提示: 访问 /mode/... 路径可以切换保护模式")
	r.Run(":8082")
}
