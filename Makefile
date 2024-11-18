
client: cmd/client/main.go
	go build -o client cmd/client/main.go

server: cmd/server/main.go cmd/server/handler/queue.go lib/queue/queue.go
	go build -o server cmd/server/main.go