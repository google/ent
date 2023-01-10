export CGO_ENABLED := "0"

run-server: build-server
    ./ent-server -config=ent-server.toml

build-server:
    go build github.com/google/ent/cmd/ent-server

run-web: build-web
    ./ent-web -config=ent-web.toml

build-web:
    go build github.com/google/ent/cmd/ent-web

build-proto:
    protoc --go_out=. --go_opt=paths=source_relative proto/*.proto
