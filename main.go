package main

import (
	"context"
	"database/sql"
	"os"

	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/vietquan-37/simplebank/api"
	"github.com/vietquan-37/simplebank/db/sqlc"
	"github.com/vietquan-37/simplebank/gapi"
	"github.com/vietquan-37/simplebank/mail"
	"github.com/vietquan-37/simplebank/pb"
	"github.com/vietquan-37/simplebank/util"
	"github.com/vietquan-37/simplebank/worker"
	"google.golang.org/grpc/reflection"
)

// this is to connect the test with db
func main() {

	config, err := util.LoadConfig(".")
	if config.Environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load from configuration")
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to db")
	}
	//run migration
	runDBMigration(config.MigrationURL, config.DBSource)
	store := sqlc.NewStore(conn)
	redisOpt := asynq.RedisClientOpt{
		Addr: config.RedisAddress,
	}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	go runTaskProcessor(redisOpt, store, config)
	go runGatewayServer(config, store, taskDistributor)
	runGrpcServer(config, store, taskDistributor)
}
func runGrpcServer(config util.Config, store sqlc.Store, taskDistributor worker.TaskDistributor) {

	server, err := gapi.NewServer(config, store, taskDistributor)

	if err != nil {
		log.Fatal().Err(err).Msg("cannot create to server:")
	}
	grpcLogger := grpc.UnaryInterceptor(gapi.GrpcLogger)
	grpcServer := grpc.NewServer(grpcLogger)
	pb.RegisterSimpleBankServer(grpcServer, server)
	reflection.Register(grpcServer) // this allow grpc client to explore what rpcs are available to call them
	listener, err := net.Listen("tcp", config.GRPCAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}
	log.Info().Msgf("start  gRPC server server at %s", listener.Addr().String())
	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to server:")
	}

}
func runTaskProcessor(redisOpt asynq.RedisClientOpt, store sqlc.Store, config util.Config) {
	mailSender := mail.NewEmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	processor := worker.NewRedisTaskProcessor(redisOpt, store, mailSender)
	log.Info().Msg("start task processor")
	err := processor.Start()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start processor")
	}
}
func runGatewayServer(config util.Config, store sqlc.Store, taskDistributor worker.TaskDistributor) {
	//create grpc server
	server, err := gapi.NewServer(config, store, taskDistributor)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create to server:")
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
		log.Fatal().Err(err).Msg("cannot register handler server:")
	}
	//this mux will receive http requests from client
	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)
	fs := http.FileServer(http.Dir("./swagger"))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fs))
	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create listener")
	}
	log.Info().Msgf("start HTTP gateway server at %s", listener.Addr().String())
	handler := gapi.HtppLogger(mux)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to HTTP gateway:")
	}

}

func runGinServer(config util.Config, store sqlc.Store) {
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create to server:")
	}
	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to server:")
	}
}
func runDBMigration(migrationUrl string, dbSource string) {
	migration, err := migrate.New(migrationUrl, dbSource)
	if err != nil {
		log.Fatal().Err(err).Msg("cannot create new migrate instance: ")
	}
	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("fail to  run migrate up: ")
	}
	log.Info().Msg("db migrate successfully")
}
