package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

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
	go runGatewayServer(config, store)
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
func runGatewayServer(config util.Config, store sqlc.Store) {

	server, err := gapi.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create to server:", err)
	}
	jsonOption := runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: true,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	})
	//grpcmux will handle the HTTP request from the client and convert it to gRPC.
	grpcMux := runtime.NewServeMux(jsonOption)
	//avoid server to do an unnecessary work
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//register a handler server for grpc gateway to call function of grpc server which is inital in first can call fuction in
	err = pb.RegisterSimpleBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal("cannot register handler server:", err)
	}
	//this mux will receive http requests from client
	mux := http.NewServeMux()
	//convert in grpc
	mux.Handle("/", grpcMux)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal("cannot create listener")
	}
	log.Printf("start HTTP gateway server at %s", listener.Addr().String())
	err = http.Serve(listener, mux)
	if err != nil {
		log.Fatal("cannot connect to HTTP gateway:", err)
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
