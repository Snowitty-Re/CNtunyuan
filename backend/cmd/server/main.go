package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/Snowitty-Re/CNtunyuan/internal/config"
	"github.com/Snowitty-Re/CNtunyuan/internal/di"
	"github.com/Snowitty-Re/CNtunyuan/internal/interfaces/http/router"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
)

func main() {
	var (
		configPath = flag.String("config", "config/config.yaml", "config file path")
		migrate    = flag.Bool("migrate", false, "run database migration")
		init       = flag.Bool("init", false, "initialize base data")
	)
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Load config failed: %v", err)
	}

	logger.Init(&cfg.Log)

	container, err := di.NewContainer(cfg)
	if err != nil {
		logger.Fatal("Create container failed", logger.Err(err))
	}

	if *migrate {
		if err := runMigration(container.DB); err != nil {
			logger.Fatal("Migration failed", logger.Err(err))
		}
		logger.Info("Migration completed")
		return
	}

	if *init {
		if err := initData(container.DB); err != nil {
			logger.Fatal("Init data failed", logger.Err(err))
		}
		logger.Info("Data initialized")
		return
	}

	r := router.NewRouter(
		container.AuthHandler,
		container.UserHandler,
		container.AuthMiddleware,
	)
	r.Setup()

	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info("Server starting", logger.String("address", addr))

	if err := r.GetEngine().Run(addr); err != nil {
		logger.Fatal("Server start failed", logger.Err(err))
	}
}

func runMigration(db interface{}) error {
	return nil
}

func initData(db interface{}) error {
	return nil
}
