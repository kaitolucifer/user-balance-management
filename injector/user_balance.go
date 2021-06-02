package injector

import (
	"github.com/kaitolucifer/user-balance-management/domain"
	"github.com/kaitolucifer/user-balance-management/infrastructure"
	"github.com/kaitolucifer/user-balance-management/presentation"
	"github.com/kaitolucifer/user-balance-management/usecase"
)

func InjectDatabase(dsn string) infrastructure.DB {
	db := infrastructure.NewDatabase(dsn)
	return *db
}

func InjectRepository(db infrastructure.DB) domain.UserBalanceRepository {
	repo := infrastructure.NewUserBalanceRepository(db)
	return repo
}

func InjectUsecase(repo domain.UserBalanceRepository) domain.UserBalanceUsecase {
	usecase := usecase.NewUserBalanceUsecase(repo)
	return usecase
}

func InjectHandler(usecase domain.UserBalanceUsecase, app *presentation.App) *presentation.UserBalanceHandler {
	handler := presentation.NewUserBalanceHander(usecase, app)
	return handler
}
