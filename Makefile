cli:
	go build -mod vendor -o bin/parse-lcsh cmd/parse-lcsh/main.go
	go build -mod vendor -o bin/parse-lcnaf cmd/parse-lcnaf/main.go
