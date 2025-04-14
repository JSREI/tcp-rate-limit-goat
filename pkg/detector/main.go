// Package detector 提供TCP连接级别的访问频率限制功能
// 主要实现了一个基于IP和端口的访问速率检测和限制机制
package detector

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/time/rate"
)

// TcpRateLimiterDetector 是一个TCP级别的访问速率检测器
// 它使用基于令牌桶算法的速率限制器来控制每个IP的访问频率
type TcpRateLimiterDetector struct {
	// lock 用于保证并发安全地访问缓存
	lock *sync.Mutex

	// cache 是一个带有过期时间的LRU缓存
	// key: IP地址（包括端口）
	// value: 对应IP的令牌桶限速器
	cache *expirable.LRU[string, *rate.Limiter]

	// Limit 定义每秒允许的请求数量
	// 例如：20表示每秒最多允许20个请求
	Limit rate.Limit

	// Burst 定义允许的突发请求数量
	// 允许在短时间内超过Limit限制的请求数
	Burst int
}

// NewTcpRateLimiter 创建并初始化一个新的TCP访问频率限制器
// 默认配置：
// - 缓存大小：100万个条目
// - 缓存过期时间：1小时
// - 每秒请求限制：20个
// - 突发请求数：20个
func NewTcpRateLimiter() *TcpRateLimiterDetector {
	return &TcpRateLimiterDetector{
		lock: &sync.Mutex{},
		// 创建一个最大容量为100万的LRU缓存，条目1小时后过期
		cache: expirable.NewLRU[string, *rate.Limiter](1000000, nil, time.Hour*1),
		Limit: 20, // 每秒20个请求
		Burst: 20, // 允许20个突发请求
	}
}

// Visit 检查并限制指定TCP连接（通过IP:端口）的访问频率
// 如果允许访问，返回true；否则返回false
func (x *TcpRateLimiterDetector) Visit(tcpRemoteAddress string) bool {
	x.lock.Lock()
	defer x.lock.Unlock()

	// 尝试从缓存中获取对应地址的限速器
	limiter, ok := x.cache.Get(tcpRemoteAddress)
	if !ok {
		// 如果缓存中不存在，则为该地址创建新的限速器
		limiter = rate.NewLimiter(x.Limit, x.Burst)
		x.cache.Add(tcpRemoteAddress, limiter)
	}

	// 使用令牌桶算法检查是否允许本次访问
	return limiter.Allow()
}

// 全局TCP速率限制器实例
var tcpRateLimiter *TcpRateLimiterDetector = NewTcpRateLimiter()

// tcpRateLimitMiddleware 是Gin框架的中间件，用于实现TCP连接级别的访问频率限制
// 基于客户端的IP:端口进行访问频率控制
func tcpRateLimitMiddleware(c *gin.Context) {
	// 获取客户端的远程地址（IP:端口）
	// 注：由于获取实际TCP连接句柄有困难，这里使用IP:端口作为唯一标识
	remoteAddress := c.Request.RemoteAddr

	// 检查是否允许访问
	if tcpRateLimiter.Visit(remoteAddress) {
		fmt.Println("允许访问：", remoteAddress)
		c.Next() // 继续处理请求
	} else {
		fmt.Println("拒绝访问：", remoteAddress)
		// 返回403禁止访问错误
		c.AbortWithError(403, fmt.Errorf("访问受限，疑似爬虫行为"))
	}
}

// main 函数是一个示例程序，展示了TCP速率限制器的使用方法
func main() {
	// 创建Gin路由器
	router := gin.Default()

	// 使用TCP速率限制中间件
	router.Use(tcpRateLimitMiddleware)

	// 定义根路由，返回欢迎消息
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "系统正常，您已获得访问权限。")
	})

	// 在8080端口启动服务器
	router.Run(":8080")
}
