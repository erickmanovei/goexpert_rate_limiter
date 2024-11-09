package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/erickmanovei/goexpert_rate_limiter/internal/middleware"
	"github.com/erickmanovei/goexpert_rate_limiter/internal/ratelimiter"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Erro ao carregar o arquivo .env")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
	})

	rateLimiter := ratelimiter.NewRateLimiter(rdb)
	mux := http.NewServeMux()
	mux.Handle("/", middleware.RateLimiterMiddleware(rateLimiter)(http.HandlerFunc(handleRequest)))

	log.Println("Servidor rodando na porta 8080")
	http.ListenAndServe(":8080", mux)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Request aceita"))
}
