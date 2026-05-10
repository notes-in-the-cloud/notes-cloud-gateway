package main

import (
	"log"
	"net/http"
	"os"
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

	jwtService := accesstoken.NewService(&realTime{}, *jwtCfg, jwt.SigningMethodHS256)

	serviceProxy, err := proxy.NewServiceProxy(proxy.ServiceURLs{
		AuthURL:     getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		NotesURL:    getEnv("NOTES_SERVICE_URL", "http://localhost:8082"),
		TodoURL:     getEnv("TODO_SERVICE_URL", "http://localhost:8085"),
		SharingURL:  getEnv("SHARING_SERVICE_URL", "http://localhost:8083"),
		ReminderURL: getEnv("REMINDER_SERVICE_URL", "http://localhost:8084"),
	})
	if err != nil {
		log.Fatalf("failed to create service proxy: %v", err)
	}

	router := proxy.NewRouter(serviceProxy, jwtService)

	addr := ":" + getEnv("SERVER_PORT", "8090")
	log.Printf("API Gateway starting on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type realTime struct{}

func (realTime) Now() time.Time { return time.Now() }
