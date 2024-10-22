package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"golang.org/x/time/rate"
	"net/http"
	"sync"
	"time"
)

type TcpRateLimiter struct {
	lock  *sync.Mutex
	cache *expirable.LRU[string, *rate.Limiter]

	// 每分钟允许访问次数
	Limit rate.Limit

	// 允许的突发大小
	Burst int
}

func NewTcpRateLimiter() *TcpRateLimiter {
	return &TcpRateLimiter{
		lock: &sync.Mutex{},
		// make cache with 1 hour TTL and 1000000 max keys
		cache: expirable.NewLRU[string, *rate.Limiter](1000000, nil, time.Hour*1),
		Limit: 20,
		Burst: 20,
	}
}

func (x *TcpRateLimiter) Visit(tcpRemoteAddress string) bool {
	x.lock.Lock()
	defer x.lock.Unlock()
	limiter, ok := x.cache.Get(tcpRemoteAddress)
	if !ok {
		limiter = rate.NewLimiter(x.Limit, x.Burst)
		x.cache.Add(tcpRemoteAddress, limiter)
	}
	return limiter.Allow()
}

var tcpRateLimiter *TcpRateLimiter = NewTcpRateLimiter()

func tcpRateLimitMiddleware(c *gin.Context) {

	// 2024-10-22 22:47:52
	// golang里不太好拿到tcp连接的句柄
	// 但是注意到有个特性（Mac OS），刚刚被用过的端口短时间内不会再被重复使用
	// 所以这里就先拿客户端的 ip:port 作为唯一标识，先逻辑上认为作用基本等同
	remoteAddress := c.Request.RemoteAddr

	if tcpRateLimiter.Visit(remoteAddress) {
		fmt.Println("allow ", remoteAddress)
		c.Next()
	} else {
		fmt.Println("fuck  ", remoteAddress)
		c.AbortWithError(403, fmt.Errorf("fuck you, spider"))
	}

}

func main() {

	router := gin.Default()
	router.Use(tcpRateLimitMiddleware)

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, everything looks fine, you have access to the system.")
	})

	router.Run(":8080")

}
