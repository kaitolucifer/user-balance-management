package presentation

import (
	"context"
	"testing"

	"github.com/kaitolucifer/user-balance-management/presentation/grpc/proto"
)

var grpcHealthCheckHandler *HealthCheckHandler

func TestCheck(t *testing.T) {
	ctx := context.Background()
	req := &proto.HealthCheckRequest{}
	resp, err := grpcHealthCheckHandler.Check(ctx, req)
	if err != nil {
		t.Fatalf("expect no error but got [%s]", err)
	}
	if resp.Status != proto.HealthCheckResponse_SERVING {
		t.Errorf("expect status [%d] but got [%s]", proto.HealthCheckResponse_SERVING, resp.Status)
	}
}

