package presentation

import (
	"context"

	"github.com/kaitolucifer/user-balance-management/presentation/grpc/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HealthCheckHandler struct{}

func (h *HealthCheckHandler) Check(context.Context, *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	return &proto.HealthCheckResponse{
		Status: proto.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthCheckHandler) Watch(*proto.HealthCheckRequest, proto.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "watch is not implemented.")
}
