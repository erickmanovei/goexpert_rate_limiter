FROM golang:1.22.1

WORKDIR /app

COPY . .

# RUN go mod download
# RUN go build -o main .

EXPOSE 8080

CMD ["sh", "-c", "cd ./cmd/ratelimiter && go run main.go"]
