package injector

import (
	"github.com/kaitolucifer/user-balance-management/domain"
	"github.com/kaitolucifer/user-balance-management/infrastructure"
	RestfulHandler "github.com/kaitolucifer/user-balance-management/presentation/restful"
	GrpcHandler "github.com/kaitolucifer/user-balance-management/presentation/grpc"
	"github.com/kaitolucifer/user-balance-management/usecase"
)

// InjectDatabase DBを注入
func InjectDatabase(dsn string) infrastructure.DB {
	db := infrastructure.NewDatabase(dsn)
	return *db
}

// InjectRepository repositoryを注入
func InjectRepository(db infrastructure.DB) domain.UserBalanceRepository {
	repo := infrastructure.NewUserBalanceRepository(db)
	return repo
}

// InjectUsecase usecaseを注入
func InjectUsecase(repo domain.UserBalanceRepository) domain.UserBalanceUsecase {
	usecase := usecase.NewUserBalanceUsecase(repo)
	return usecase
}

// InjectRestfulHandler RESTful handlerを注入
func InjectRestfulHandler(usecase domain.UserBalanceUsecase, app *RestfulHandler.App) *RestfulHandler.RestfulUserBalanceHandler {
	handler := RestfulHandler.NewRestfulUserBalanceHander(usecase, app)
	return handler
}

// InjectGrpcHandler grpc handlerを注入
func InjectGrpcHandler(usecase domain.UserBalanceUsecase, app *GrpcHandler.App) *GrpcHandler.GrpcUserBalanceHander {
	handler := GrpcHandler.NewGrpcUserBalanceHander(usecase, app)
	return handler
}
