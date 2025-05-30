export CGO_ENABLED := "0"

build-server: build-proto
    go build -o ./bin/ent-server github.com/google/ent/cmd/ent-server

run-server: build-server
    ./bin/ent-server -config=ent-server.toml

build-api:
    go build -o ./bin/ent-api github.com/google/ent/cmd/ent-api

run-api: build-api
    ./bin/ent-api -config=ent-api.toml

build-web:
    go build -o ./bin/ent-web github.com/google/ent/cmd/ent-web

run-web: build-web
    ./bin/ent-web -config=ent-web.toml

build-proto:
    protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto

install-cli:
    go build -o ./bin/ent ./cmd/ent

test-cli: install-cli
    ./bin/ent get --digest=sha2-256:fba3120c26c61a92490736f9afa329faed0117bdb6e579011039dc4e34071507
