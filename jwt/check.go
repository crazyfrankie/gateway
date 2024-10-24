package jwt

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt"

	"gateway/route"
)

type Claim struct {
	jwt.StandardClaims
	SSID string
	Role string
}

type LoginJWTMiddleware struct {
	ignorePaths map[string]struct{}
	adminPaths  map[string]struct{}
}

func NewLoginJWTMiddleware() *LoginJWTMiddleware {
	return &LoginJWTMiddleware{
		ignorePaths: make(map[string]struct{}),
		adminPaths:  make(map[string]struct{}),
	}
}

func (l *LoginJWTMiddleware) IgnorePath(path string) *LoginJWTMiddleware {
	l.ignorePaths[path] = struct{}{}
	return l
}

func (l *LoginJWTMiddleware) AdminPath(path string) *LoginJWTMiddleware {
	l.adminPaths[path] = struct{}{}
	return l
}

func (l *LoginJWTMiddleware) CheckLogin(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 首先检查路径是否需要忽略 JWT 验证
		if _, ok := l.ignorePaths[r.URL.Path]; ok {
			// 如果是忽略路径，直接替换路径并转发请求
			if mappedPath, ok := route.PathMappings[r.URL.Path]; ok {
				r.URL.Path = mappedPath // 替换路径
			}
			next.ServeHTTP(w, r) // 继续执行请求
			return
		}

		tokenHeader := ExtractToken(r)
		if tokenHeader == "" {
			http.Error(w, "missing or malformed token", http.StatusUnauthorized)
			return
		}

		// 解析 Token
		claim, err := ParseToken(tokenHeader)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// 检查管理员路径
		if _, ok := l.adminPaths[r.URL.Path]; ok && !(claim.Role == "admin") {
			http.Error(w, "access denied for non-merchants", http.StatusUnauthorized)
			return
		}

		// 在上下文中设置 claims 信息，以便后续的处理逻辑使用
		ctx := context.WithValue(r.Context(), "claims", claim)
		r = r.WithContext(ctx)

		// 处理路径映射
		if mappedPath, ok := route.PathMappings[r.URL.Path]; ok {
			r.URL.Path = mappedPath // 替换路径
		} else {
			http.NotFound(w, r) // 未找到匹配路径
			return
		}

		next.ServeHTTP(w, r) // 继续执行请求
	}
}

func ExtractToken(r *http.Request) string {
	tokenHeader := r.Header.Get("Authorization")
	// 检查请求头中是否包含 Token
	if tokenHeader == "" {
		return ""
	}
	return tokenHeader
}

func ParseToken(tokenString string) (*Claim, error) {
	// 解析 token
	token, err := jwt.ParseWithClaims(tokenString, &Claim{}, func(token *jwt.Token) (interface{}, error) {
		// 确保 token 使用的签名方法是我们期望的
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// 返回用于验证的密钥
		return []byte("KsS2X1CgFT4bi3BRRIxLk5jjiUBj8wxE"), nil
	})

	if claims, ok := token.Claims.(*Claim); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}
}
