package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Dev-noauth: allow all requests, set a dummy user_id
		if _, exists := c.Get("user_id"); !exists {
			c.Set("user_id", "dev-user")
		}
		c.Next()
	}
}

// ---- JWKS cache ----
type jwksKey struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}
type jwksSet struct {
	Keys []jwksKey `json:"keys"`
}

var (
	jwksCache     jwksSet
	jwksCacheOnce sync.Once
	jwksCachedAt  time.Time
	jwksMu        sync.RWMutex
)

func jwksKeyFunc(c *gin.Context) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		// Разрешаем только RS256
		if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Проверки iss/aud до подписи (часть клиентов делает после) — читаем claims безопасно
		claims, _ := token.Claims.(jwt.MapClaims)
		expectedIss := os.Getenv("JWT_ISSUER")
		expectedAud := os.Getenv("JWT_AUDIENCE")
		if expectedIss != "" {
			if iss, _ := claims["iss"].(string); iss != expectedIss {
				return nil, errors.New("invalid issuer")
			}
		}
		if expectedAud != "" {
			if aud, _ := claims["aud"].(string); aud != expectedAud {
				return nil, errors.New("invalid audience")
			}
		}

		kid, _ := token.Header["kid"].(string)
		pub := fetchJWKSAndFindKey(c, kid)
		if pub == nil {
			return nil, errors.New("jwks key not found")
		}
		return pub, nil
	}
}

func fetchJWKSAndFindKey(c *gin.Context, kid string) *rsa.PublicKey {
	url := os.Getenv("JWKS_URL")
	if url == "" {
		// dev default: user-service publishes on 8081
		url = "http://user-service:8081/.well-known/jwks.json"
	}
	// refresh cache every 5 minutes
	jwksMu.RLock()
	valid := time.Since(jwksCachedAt) < 5*time.Minute && len(jwksCache.Keys) > 0
	jwksMu.RUnlock()
	if !valid {
		resp, err := http.Get(url)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()
		var set jwksSet
		if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
			return nil
		}
		jwksMu.Lock()
		jwksCache = set
		jwksCachedAt = time.Now()
		jwksMu.Unlock()
	}
	jwksMu.RLock()
	defer jwksMu.RUnlock()
	for _, k := range jwksCache.Keys {
		if k.Kid == kid && k.Kty == "RSA" && (k.Alg == "RS256" || k.Alg == "") {
			nBytes, errN := base64RawURLDecode(k.N)
			eBytes, errE := base64RawURLDecode(k.E)
			if errN != nil || errE != nil {
				return nil
			}
			n := new(big.Int).SetBytes(nBytes)
			var eInt int
			if len(eBytes) == 3 {
				eInt = int(eBytes[0])<<16 | int(eBytes[1])<<8 | int(eBytes[2])
			} else if len(eBytes) == 1 {
				eInt = int(eBytes[0])
			} else {
				return nil
			}
			return &rsa.PublicKey{N: n, E: eInt}
		}
	}
	return nil
}

func base64RawURLDecode(s string) ([]byte, error) {
	// local decoder to avoid extra deps
	// base64.RawURLEncoding expects no padding, so use it directly
	return base64.RawURLEncoding.DecodeString(s)
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		// Поддержим список через env FRONTEND_ORIGINS (CSV), дефолт: localhost:3000 и 127.0.0.1:3000
		allowed := os.Getenv("FRONTEND_ORIGINS")
		if allowed == "" {
			allowed = "http://localhost:3000,http://127.0.0.1:3000"
		}
		allowSet := map[string]struct{}{}
		for _, o := range strings.Split(allowed, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				allowSet[o] = struct{}{}
			}
		}
		if _, ok := allowSet[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Vary", "Origin")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
