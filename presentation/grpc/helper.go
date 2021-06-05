package presentation

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgconn"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func handleError(err error) *status.Status {
	var st *status.Status
	var pgErr *pgconn.PgError
	if err == nil {
		st = status.New(codes.OK, "")
	} else if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			st = status.New(codes.AlreadyExists, "transaction_id must be unique")
		default:
			st = status.New(codes.Internal, "database error")
		}
	} else {
		if err == sql.ErrNoRows {
			st = status.New(codes.NotFound, "user_id not found")
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
