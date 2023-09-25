GOMOD=$(shell test -f "go.work" && echo "readonly" || echo "vendor")

cli:
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/parse-lcsh cmd/parse-lcsh/main.go
	go build -mod $(GOMOD) -ldflags="-s -w" -o bin/parse-lcnaf cmd/parse-lcnaf/main.go
