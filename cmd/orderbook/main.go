package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/tinkoff/invest-api-go-sdk/investgo"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/bataloff/tiknkoff/config"
	"github.com/bataloff/tiknkoff/internal/repositories"
	"github.com/bataloff/tiknkoff/internal/usecases"
	"github.com/bataloff/tiknkoff/pkg/database/sqlite"
	"github.com/bataloff/tiknkoff/pkg/signal"
)

func main() {
	application := cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "config-file",
				Required: true,
				Usage:    "YAML config filepath",
				EnvVars:  []string{"TINKOFF_INVEST_CONFIG_FILE"},
				FilePath: "/srv/tinkoff_invest_secrets/config_file",
			},
		},
		Action: Main,
		After: func(c *cli.Context) error {
			log.Println("stopped")
			return nil
		},
	}

	if err := application.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}

func Main(ctx *cli.Context) error {
	appContext, cancel := context.WithCancel(ctx.Context)
	defer func() {
		cancel()
	}()

	cfg, err := config.New(ctx.String("config-file"))
	if err != nil {
		return err
	}

	// загружаем конфигурацию для сдк из .yaml файла
	investCfg, err := investgo.LoadConfig(cfg.InvestConfigFilePath)
	if err != nil {
		return err
	}

	// ВСЕ ЭТО МОЖНО КУДА-ТО ВЫНЕСТИ
	// сдк использует для внутреннего логирования investgo.Logger
	// для примера передадим uber.zap
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	zapConfig.EncoderConfig.TimeKey = "time"
	l, err := zapConfig.Build()
	logger := l.Sugar()
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf(err.Error())
		}
	}()
	// ВСЕ ЭТО МОЖНО КУДА-ТО ВЫНЕСТИ

	client, err := investgo.NewClient(appContext, investCfg, logger)
	if err != nil {
		logger.Fatalf("client creating error %v", err.Error())
	}

	// чтобы не было много таких defer, можно обернуть в DropWrap (pkg), который будет вызывать метод Drop
	// после регистрации дропера автоматически при завершении приложения, см. доку по этому пакету там же.
	defer func() {
		logger.Infof("closing client connection")
		err := client.Stop()
		if err != nil {
			logger.Errorf("client shutdown error %v", err.Error())
		}
	}()

	connect, err := sqlite.New(appContext, cfg.Database.Path, cfg.Database.Debug)
	if err != nil {
		return err
	}
	defer func() {
		_ = connect.Drop()
	}()

	orderBookRepo, err := repositories.NewOrderBookRepository(appContext, connect)
	if err != nil {
		return err
	}

	useCase := usecases.NewOrderBookUseCase(orderBookRepo, client)

	await, stop := signal.Notifier(func() {
		// call before shutdown application
	})

	go func() {
		if err = useCase.Sync(appContext); err != nil {
			stop(err)
		}
	}()

	return await()
}
