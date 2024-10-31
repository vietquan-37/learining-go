.PHONY: postgres createdb dropdb migrateup migratedown sqlc test server mock migrateup1 migratedown1 proto evans
postgres:
	@docker run --name postgres12 -p 5431:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=12345 -d postgres
createdb:
	@docker exec -it postgres12 createdb --username=postgres --owner=postgres simple_bank
migrateup:
	@migrate -path db/migration -database "postgresql://postgres:12345@localhost:5431/simple_bank?sslmode=disable" -verbose up
migratedown:
	@migrate -path db/migration -database "postgresql://postgres:12345@localhost:5431/simple_bank?sslmode=disable" -verbose down
migrateup1:
	@migrate -path db/migration -database "postgresql://postgres:12345@localhost:5431/simple_bank?sslmode=disable" -verbose up 1
migratedown1:
	@migrate -path db/migration -database "postgresql://postgres:12345@localhost:5431/simple_bank?sslmode=disable" -verbose down 1	
dropdb:
	@docker exec -it postgres12 dropdb simple_bank
test:
	@go test -v -cover ./...
server:
	@go run main.go
mock:
	@mockgen -package mockdb -destination db/mock/store.go  github.com/vietquan-37/simplebank/db/sqlc Store
proto:
	@if exist pb\*.go del /Q pb\*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb  --grpc-gateway_opt paths=source_relative \
    proto/*.proto
evans:
	@evans --host localhost --port 9090 -r repl
sqlc:
	@sqlc generate
