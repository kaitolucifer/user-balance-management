package injector

import (
	"github.com/kaitolucifer/user-balance-management/domain"
	"github.com/kaitolucifer/user-balance-management/infrastructure"
	"github.com/kaitolucifer/user-balance-management/presentation"
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

// InjectHandler handlerまたはcontrollerを注入
func InjectHandler(usecase domain.UserBalanceUsecase, app *presentation.App) *presentation.UserBalanceHandler {
	handler := presentation.NewUserBalanceHander(usecase, app)
	return handler
}
