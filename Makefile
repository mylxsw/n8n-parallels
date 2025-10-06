
run:
	go run cmd/server/main.go

build:
	go build -o build/server/n8n-parallels-server cmd/server/main.go

.PHONY: run build