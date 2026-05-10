package main

import (
	"github.com/notes-in-the-cloud/notes-cloud-api-gateway/internal/middlewares"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/notes-in-the-cloud/notes-cloud-api-gateway/internal/proxy"
	"github.com/notes-in-the-cloud/notes-cloud-jwt-utils/accesstoken"
)

func main() {
	jwtCfg, err := accesstoken.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load jwt config: %v", err)
	}

	internalToken := getEnv("INTERNAL_SERVICE_TOKEN", "")
	if internalToken == "" {
		log.Fatal("INTERNAL_TOKEN is required")
	}

	allowedOrigins := splitCSV(getEnv(
		"ALLOWED_ORIGINS",
		"http://localhost:3000,http://localhost:5173,http://127.0.0.1:3000,http://127.0.0.1:5173",
	))

	jwtService := accesstoken.NewService(&realTime{}, *jwtCfg, jwt.SigningMethodHS256)

	serviceProxy, err := proxy.NewProxy(proxy.ServiceURLs{
		AuthURL:     getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		NotesURL:    getEnv("NOTES_SERVICE_URL", "http://localhost:8082"),
		TodoURL:     getEnv("TODO_SERVICE_URL", "http://localhost:8085"),
		SharingURL:  getEnv("SHARING_SERVICE_URL", "http://localhost:8083"),
		ReminderURL: getEnv("REMINDER_SERVICE_URL", "http://localhost:8084"),
	})
	if err != nil {
		log.Fatalf("failed to create service proxy: %v", err)
	}

	router := proxy.NewRouter(
		serviceProxy,
		jwtService,
		internalToken,
		allowedOrigins,
	)

	routerWithCORS := middlewares.CORS(router, allowedOrigins)

	addr := ":" + getEnv("SERVER_PORT", "8090")
	log.Printf("API Gateway starting on %s", addr)

	if err := http.ListenAndServe(addr, routerWithCORS); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmedPart := strings.TrimSpace(part)
		if trimmedPart == "" {
			continue
		}

		result = append(result, trimmedPart)
	}

	return result
}

type realTime struct{}

func (realTime) Now() time.Time { return time.Now() }
