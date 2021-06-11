package presentation

import (
	"context"
	"errors"
	"log"

	"github.com/kaitolucifer/user-balance-management/domain"
	"github.com/kaitolucifer/user-balance-management/presentation/grpc/proto"
)

// App アプリケーションが持つコンポーネントや設定を格納
type App struct {
	InfoLog  *log.Logger
	ErrorLog *log.Logger
}

// GrpcUserBalanceHander usecaseとアプリケーション設定を格納
type GrpcUserBalanceHander struct {
	usecase domain.UserBalanceUsecase
	App     *App
}

// NewGrpcUserBalanceHander 新しいgRPCハンドラを作成
func NewGrpcUserBalanceHander(usecase domain.UserBalanceUsecase, app *App) *GrpcUserBalanceHander {
	return &GrpcUserBalanceHander{
		usecase: usecase,
		App:     app,
	}
}

// GetBalanceByUserID ユーザーIDでの残高を取得するRPC
func (h *GrpcUserBalanceHander) GetBalanceByUserID(ctx context.Context, req *proto.GetUserBalanceRequest) (*proto.GetUserBalanceResponse, error) {
	resp := &proto.GetUserBalanceResponse{}
	var err error
	if req.UserId == "" {
		err = errors.New("user_id is empty")
	} else {
		balance, newErr := h.usecase.GetBalance(req.UserId)
		if newErr == nil {
			resp = &proto.GetUserBalanceResponse{
				Balance: int32(balance),
			}
		} else {
			err = newErr
		}
	}

	if err != nil {
		h.App.ErrorLog.Println(err)
	}

	st := handleError(err)
	return resp, st.Err()
}

// ChangeBalanceByUserID ユーザーIDで残高を更新するハンドラ
func (h *GrpcUserBalanceHander) ChangeBalanceByUserID(ctx context.Context, req *proto.ChangeUserBalanceRequest) (*proto.EmptyResponse, error) {
	resp := &proto.EmptyResponse{}
	var err error
	if req.UserId == "" {
		err = errors.New("user_id is empty")
	} else if req.TransactionId == "" {
		err = errors.New("transaction_id is empty")
	} else {
		if req.Amount > 0 {
			err = h.usecase.AddBalance(req.UserId, int(req.Amount), req.TransactionId)
		} else if req.Amount < 0 {
			err = h.usecase.ReduceBalance(req.UserId, -int(req.Amount), req.TransactionId)
		} else {
			err = errors.New("amount can't be 0")
		}
	}

	if err != nil {
		h.App.ErrorLog.Println(err)
	}

	st := handleError(err)
	return resp, st.Err()
}

// AddAllUserBalance 残高を一斉に加算するハンドラ
func (h *GrpcUserBalanceHander) AddAllUserBalance(ctx context.Context, req *proto.AddAllUserBalanceRequest) (*proto.EmptyResponse, error) {
	resp := &proto.EmptyResponse{}

	var err error
	if req.TransactionId == "" {
		err = errors.New("transaction_id is empty")
	} else if req.Amount <= 0 {
		err = errors.New("amount must be positive")
	} else {
		err = h.usecase.AddAllUserBalance(int(req.Amount), req.TransactionId)
	}

	if err != nil {
		h.App.ErrorLog.Println(err)
	}

	st := handleError(err)
	return resp, st.Err()
}
