export CGO_ENABLED := "0"

run-server: build-server
    ./bin/ent-server -config=ent-server.toml

build-server:
    go build -o ./bin/ent-server github.com/google/ent/cmd/ent-server

run-web: build-web
    ./bin/ent-web -config=ent-web.toml

build-web:
    go build -o ./bin/ent-web github.com/google/ent/cmd/ent-web

build-proto:
    protoc --go_out=. --go_opt=paths=source_relative proto/*.proto

install-cli:
    go build -o ./bin/ent ./cmd/ent
