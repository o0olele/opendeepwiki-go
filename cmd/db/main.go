package main

import (
	"github.com/o0olele/opendeepwiki-go/internal/database"
	"github.com/o0olele/opendeepwiki-go/internal/database/dao"
	"go.uber.org/zap"
)

func main() {
	if err := database.InitDB("./data/sqlite/opendeepwiki.db"); err != nil {
		zap.L().Error("Failed to initialize database", zap.Error(err))
	}

	repoDao := dao.NewRepositoryDAO()

	repoDao.UpdateRepositoryStatus(1, 3)
}
