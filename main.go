package main

import (
	"gateway/route"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"gateway/cors"
	"gateway/jwt"
)

func newReverseProxy(target string) *httputil.ReverseProxy {
	targetUrl, err := url.Parse(target)
	if err != nil {
		return nil
	}

	return httputil.NewSingleHostReverseProxy(targetUrl)
}

func main() {
	// 初始化 etcd 以及读取路由配置
	route.InitEtcd()

	target := "http://127.0.0.1:9000"
	proxy := newReverseProxy(target)

	jwtHandler := jwt.NewLoginJWTMiddleware().
		IgnorePath("/user/send-code").
		IgnorePath("/user/verify-code")

	http.HandleFunc("/", cors.CORS(jwtHandler.CheckLogin(proxy)))

	// 启动网关服务器
	log.Println("Starting gateway on port 8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
