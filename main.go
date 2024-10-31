package main

import (
	"database/sql"
	"log"
	"net"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"github.com/vietquan-37/simplebank/api"
	"github.com/vietquan-37/simplebank/db/sqlc"
	"github.com/vietquan-37/simplebank/gapi"
	"github.com/vietquan-37/simplebank/pb"
	"github.com/vietquan-37/simplebank/util"
	"google.golang.org/grpc/reflection"
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
	runGrpcServer(config, store)
}
func runGrpcServer(config util.Config, store sqlc.Store) {

	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create to server:", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer) // this allow grpc client to explore what rpcs are available to call them
	listener, err := net.Listen("tcp", config.GRPCAddress)
	if err != nil {
		log.Fatal("cannot create listener")
	}
	log.Printf("start gRPC server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot connect to server:", err)
	}

}
func runGinServer(config util.Config, store sqlc.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create to server:", err)
	}
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot connect to server:", err)
	}
}
