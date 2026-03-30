package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"WBTech_L3.4/internal/api/handler"
	"WBTech_L3.4/internal/api/server"
	"WBTech_L3.4/internal/config"
	"WBTech_L3.4/internal/kafka"
	"WBTech_L3.4/internal/processor"
	"WBTech_L3.4/internal/repository"
	"WBTech_L3.4/internal/service"
	"WBTech_L3.4/internal/storage"
	"github.com/wb-go/wbf/zlog"
)

func Run() {
	os.Setenv("TZ", "UTC")

	zlog.InitConsole()
	logger := zlog.Logger
	cfg := config.Load()

	db, err := repository.NewPostgresDB(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to init database")
	}
	defer db.Master.Close()
	for _, s := range db.Slaves {
		defer s.Close()
	}

	repo := repository.NewRepository(db)

	fileStorage := storage.NewLocalStorage(cfg.StoragePath)

	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers, cfg.KafkaTopic, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create kafka producer")
	}

	imgProcessor := processor.NewProcessor(fileStorage, cfg.WatermarkPath, cfg.WatermarkText, 3, logger)

	kafkaConsumer, err := kafka.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroup, repo, imgProcessor, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create kafka consumer")
	}
	go func() {
		kafkaConsumer.Start(context.Background())
	}()

	svc := service.NewService(repo, fileStorage, kafkaProducer)

	srv := new(server.Server)
	handlers := handler.NewHandler(svc, logger)

	go func() {
		logger.Info().Str("addr", cfg.HTTPAddr).Msg("starting http server")
		if err = srv.Run(cfg.HTTPAddr, handlers.InitRoutes()); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Err(err).Msg("http server error")
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logger.Info().Msg("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("http server shutdown error")
	}
}
