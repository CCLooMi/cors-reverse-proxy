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
		if c.Request.Method == "OPTIONS" {
			hd := c.Writer.Header()
			for key, value := range conf.Cfg.Header {
				hd.Set(key, value)
			}
			hd.Set("Access-Control-Allow-Methods", c.Request.Method)
			hd.Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
			hd.Set("Access-Control-Allow-Credentials", "true")
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
			//domain := removePortFromDomain(c.Request.Host)
			proxy = httputil.NewSingleHostReverseProxy(proxyURL)
			proxy.ModifyResponse = func(r *http.Response) error {
				for key, value := range conf.Cfg.Header {
					r.Header.Set(key, value)
				}
				r.Header.Set("Access-Control-Allow-Methods", c.Request.Method)
				r.Header.Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
				hv := r.Header.Get("Access-Control-Allow-Credentials")
				if len(hv) == 0 {
					r.Header.Set("Access-Control-Allow-Credentials", "true")
				}
				//setCookies := r.Header.Values("Set-Cookie")
				//for idx, cookie := range setCookies {
				//	fmt.Println(domain, cookie)
				//	cookie = strings.Replace(cookie, "SameSite=Lax", "SameSite=None", 1)
				//	setCookies[idx] = cookie + "; Secure; Domain=" + domain
				//}
				return nil
			}
			proxyMap[targetURL.Host] = proxy
		}
		proxyMapMutex.Unlock()
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

//func removePortFromDomain(domain string) string {
//	parts := strings.Split(domain, ":")
//	return parts[0]
//}
