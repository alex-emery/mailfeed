FILES = $(shell find . -name '*.go'")

SQL_SOURCE = $(shell find ./sql/ -name '*.sql')
SQL_GENERATED = $(shell find ./database/sqlc -name '*.sql.go')

$(SQL_GENERATED): $(SQL_SOURCE)
	sqlc generate

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: build
build:
	go build -o bin/mailfeed .