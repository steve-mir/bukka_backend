DB_URL=postgresql://root:toor@localhost:5434/bukka?sslmode=disable

environ:
	export PATH="$PATH:$(go env GOPATH)/bin"
air_init:
	air init

air_run:
	air

start_ps:
	docker start postgis

start_redis:
	docker start redis

postgres:
	docker run --name postgis -p 5434:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=toor -d postgis:16-3.4-alpine

redis:
	docker run --name redis -p 6379:6379 -d redis:7.0-alpine

createdb:
	docker exec -it postgis createdb --username=root --owner=root bukka

dropdb:
	docker exec -it postgis dropdb bukka

migrate_init:
	migrate create -ext sql -dir db/migration -seq users

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

migratedown2:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 2

db_docs:
	dbdocs build docs/db.dbml

db_schema:
	dbml2sql --postgres -o docs/schema.sql docs/db.dbml

sqlc_init:
	sqlc init

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test server environ air_init air_run start_redis redis start_ps db_docs db_schema

# migrate create -ext sql -dir db/migration -seq add_user_session