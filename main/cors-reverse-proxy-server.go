package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"test/conf"
)

func main() {
	// 设置logrus为默认的日志打印库
	gin.DefaultWriter = logrus.StandardLogger().Writer()
	logLevel, err := logrus.ParseLevel(conf.Cfg.LogLevel)
	if err != nil {
		logLevel = logrus.DebugLevel
	}
	logrus.SetLevel(logLevel)

	r := gin.Default()
	// Set CORS headers
	r.Use(func(c *gin.Context) {
		for key, value := range conf.Cfg.Header {
			c.Writer.Header().Set(key, value)
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})
	// Create a map to store the reverse proxy objects
	proxyMap := make(map[string]*httputil.ReverseProxy)
	var proxyMapMutex sync.Mutex

	// Define a route that handles all incoming requests
	r.Any("/*path", func(c *gin.Context) {
		// Get the requested path from the URL
		requestPath := c.Param("path")

		// Create a new URL by combining the target base URL and the requested path
		corsUrl := "https:/" + requestPath
		targetURL, err := url.Parse(corsUrl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse target URL"})
			return
		}

		// Copy headers from host_conf
		hostConfHds := conf.Cfg.HostConf[targetURL.Host].Header
		for key, value := range hostConfHds {
			c.Request.Header.Set(key, value)
		}

		// Create a reverse proxy instance
		// last "/" is very important,if not set Path!
		proxyURL, _ := url.Parse(targetURL.Scheme + "://" + targetURL.Host)
		proxyURL.Path = "/"

		proxyMapMutex.Lock()
		proxy, ok := proxyMap[targetURL.String()]
		if !ok {
			proxy = httputil.NewSingleHostReverseProxy(proxyURL)
			proxyMap[targetURL.Host] = proxy
		}
		proxyMapMutex.Unlock()

		// Copy headers from the incoming request to the proxy request
		for key, values := range c.Request.Header {
			for _, value := range values {
				c.Request.Header.Set(key, value)
			}
		}
		// Modify the request before it's sent to the target server
		c.Request.Host = targetURL.Host
		c.Request.URL.Scheme = targetURL.Scheme
		c.Request.URL.Path = targetURL.Path
		c.Request.URL.Host = targetURL.Host
		c.Request.RequestURI = targetURL.Path + "?" + c.Request.URL.RawQuery
		c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))
		c.Request.Header.Set("Referer", corsUrl)

		// Serve the request using the reverse proxy
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	if conf.Cfg.EnableHttps {
		err = http.ListenAndServeTLS(":"+conf.Cfg.Port, conf.Cfg.CertFile, conf.Cfg.KeyFile, r)
	} else {
		err = r.Run(":" + conf.Cfg.Port)
	}
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
