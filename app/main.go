package main

import (
	"net"
	"net/http"

	"github.com/kaitolucifer/user-balance-management/presentation/grpc/proto"
	"google.golang.org/grpc"
)

const restfulPortNumber = ":8080"
const grpcPortNumber = ":50051"

func main() {
	configApp()
	defer db.Close()
	if *useGrpc {
		listener, err := net.Listen("tcp", "0.0.0.0"+grpcPortNumber)
		if err != nil {
			grpcHandler.App.ErrorLog.Fatalf("failed to listen: %s", err)
		}
		grpcHandler.App.InfoLog.Printf("starting gRPC application on port %s\n", grpcPortNumber)
		s := grpc.NewServer()
		proto.RegisterUserBalanceServer(s, grpcHandler)
		grpcHandler.App.ErrorLog.Fatal(s.Serve(listener))
	} else {
		restfulHandler.App.InfoLog.Printf("starting RESTful application on port %s\n", restfulPortNumber)
		srv := &http.Server{
			Addr:    restfulPortNumber,
			Handler: mux,
		}

		restfulHandler.App.ErrorLog.Fatal(srv.ListenAndServe())
	}
}
