.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock
postgres:
	@docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=12345 -d postgres
createdb:
	@docker exec -it postgres12 createdb --username=root --owner=root simple_bank
migrateup:
	@migrate -path db/migration -database "postgresql://postgres:12345@localhost:5432/simple_bank?sslmode=disable" -verbose up
migratedown:
	@migrate -path db/migration -database "postgresql://postgres:12345@localhost:5432/simple_bank?sslmode=disable" -verbose down
dropdb:
	@docker exec -it postgres12 dropdb simple_bank
test:
	@go test -v -cover ./...
server:
	@go run main.go
mock:
	@mockgen -package mockdb -destination db/mock/store.go mockgen -destination db/mock/store.go github.com/vietquan-37/simplebank/db/sqlc Store

sqlc:
	@sqlc generate
