package presentation

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func handleError(err error) *status.Status {
	var st *status.Status
	if err == nil {
		st = status.New(codes.OK, "")
	} else {
		if err.Error() == "database error" {
			st = status.New(codes.Internal, "database error")
		} else if err.Error() == "transaction_id must be unique" {
			st = status.New(codes.AlreadyExists, "transaction_id must be unique")
		} else if err.Error() == "user not found" {
			st = status.New(codes.NotFound, "user not found")
		} else if err.Error() == "balance insufficient" {
			st = status.New(codes.FailedPrecondition, "user balance is insufficient")
		} else if err.Error() == "update failed" {
			// データ競合が発生
			st = status.New(codes.Unavailable, "update failed, please retry")
		} else if err.Error() == "transaction_id is empty" {
			st = status.New(codes.InvalidArgument, err.Error())
		} else if err.Error() == "user_id is empty" {
			st = status.New(codes.InvalidArgument, err.Error())
		} else if err.Error() == "amount must be positive" {
			st = status.New(codes.InvalidArgument, err.Error())
		} else if err.Error() == "amount can't be 0" {
			st = status.New(codes.InvalidArgument, err.Error())
		} else {
			st = status.New(codes.Internal, "internal server error")
		}
	}

	return st
}
