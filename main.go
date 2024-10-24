package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"

	"github.com/vietquan-37/simplebank/api"
	"github.com/vietquan-37/simplebank/db/sqlc"
	"github.com/vietquan-37/simplebank/util"
)

// this is to connect the test with db
func main() {

	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load from configuration")
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db")
	}

	store := sqlc.NewStore(conn)
	server := api.NewServer(store)
	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot connect to server:", err)
	}
}
