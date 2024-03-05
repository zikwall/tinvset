package main

import (
	"context"
	"log"
	"os"

	"github.com/tinkoff/invest-api-go-sdk/investgo"
	"github.com/urfave/cli/v2"

	"github.com/bataloff/tiknkoff/config"
	"github.com/bataloff/tiknkoff/internal/instance"
	"github.com/bataloff/tiknkoff/internal/repositories"
	"github.com/bataloff/tiknkoff/internal/usecases"
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

	investCfg, err := investgo.LoadConfig(cfg.InvestConfigFilePath)
	if err != nil {
		return err
	}

	single, err := instance.New(appContext, &instance.Options{
		Database: cfg.Database,
		Invest:   investCfg,
	})
	if err != nil {
		return err
	}

	defer func() {
		single.Shutdown(func(err error) {
			single.Logger.Sugar().Error(err)
		})
	}()

	orderBookRepo, err := repositories.NewOrderBookRepository(appContext, single.Pool)
	if err != nil {
		return err
	}

	useCase := usecases.NewOrderBookUseCase(orderBookRepo, single.Client.Tinkoff())

	await, stop := signal.Notifier(func() {
		single.Logger.Sugar().Info("receive stop signal, start shutdown process..")
	})

	if cfg.PingInvest {
		if err = useCase.Ping(appContext); err != nil {
			return err
		}
	}

	go func() {
		if err = useCase.Sync(appContext); err != nil {
			stop(err)
		}
	}()

	return await()
}
