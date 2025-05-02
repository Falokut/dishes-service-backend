package assembly

import (
	"context"
	"dishes-service-backend/conf"

	"github.com/Falokut/go-kit/cluster"
	"github.com/Falokut/go-kit/dbx"
	"github.com/Falokut/go-kit/http"
	"github.com/Falokut/go-kit/http/client"
	"github.com/Falokut/go-kit/remote"
	"github.com/Falokut/go-kit/tg_botx"

	"github.com/Falokut/go-kit/app"
	"github.com/Falokut/go-kit/bootstrap"
	"github.com/Falokut/go-kit/db"
	"github.com/Falokut/go-kit/log"
	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
)

type Assembly struct {
	boot     *bootstrap.Bootstrap
	db       *dbx.Client
	tgBot    *tg_botx.Bot
	bgjobCli *bgjob.Client
	fileCli  *client.Client
	server   *http.Server
	logger   *log.Adapter
}

func New(boot *bootstrap.Bootstrap) (*Assembly, error) {
	logger := boot.App.Logger()

	db := dbx.New(logger, db.WithMigrationRunner(boot.MigrationsDir, logger))
	boot.HealthcheckRegistry.Register("db", db)

	tgBot := tg_botx.New(logger)

	filesCli := client.New()

	server := http.NewServer(logger)
	return &Assembly{
		boot:    boot,
		db:      db,
		tgBot:   tgBot,
		fileCli: filesCli,
		server:  server,
		logger:  logger,
	}, nil
}

func (a *Assembly) ReceiveConfig(shortCtx context.Context, remoteConfig []byte) error {
	newCfg, _, err := remote.Upgrade[conf.Remote](a.boot.RemoteConfig, remoteConfig)
	if err != nil {
		a.boot.Fatal(errors.WithMessage(err, "upgrade remote config"))
	}

	err = a.db.Upgrade(shortCtx, newCfg.Db)
	if err != nil {
		a.boot.Fatal(errors.WithMessage(err, "upgrade db"))
	}

	err = a.tgBot.UpgradeConfig(shortCtx, newCfg.Bot)
	if err != nil {
		a.boot.Fatal(errors.WithMessage(err, "upgrade tg bot config"))
	}
	a.fileCli.GlobalRequestConfig().BaseUrl = newCfg.Images.BaseServicePath

	locator := NewLocator(a.db, a.bgjobCli, a.fileCli, a.tgBot, a.logger)
	cfg, err := locator.LocatorConfig(shortCtx, newCfg)
	if err != nil {
		a.boot.Fatal(errors.WithMessage(err, "locator config"))
	}

	a.tgBot.UpgradeMux(shortCtx, cfg.BotRouter)
	a.server.Upgrade(cfg.HttpRouter)
	return nil
}

func (a *Assembly) Runners() []app.Runner {
	eventHandler := cluster.NewEventHandler().
		RemoteConfigReceiver(a)
	return []app.Runner{
		app.RunnerFunc(func(_ context.Context) error {
			err := a.server.ListenAndServe(a.boot.BindingAddress)
			if err != nil {
				return errors.WithMessage(err, "listen and serve http server")
			}
			return nil
		}),
		app.RunnerFunc(func(ctx context.Context) error {
			err := a.boot.ClusterCli.Run(ctx, eventHandler)
			if err != nil {
				return errors.WithMessage(err, "run cluster client")
			}
			return nil
		}),
		app.RunnerFunc(func(ctx context.Context) error {
			err := a.tgBot.Serve(ctx)
			if err != nil {
				return errors.WithMessage(err, "serve tg bot")
			}
			return nil
		}),
	}
}

func (a *Assembly) Closers() []app.Closer {
	return []app.Closer{
		app.CloserFunc(func(_ context.Context) error {
			return a.db.Close()
		}),
		app.CloserFunc(func(_ context.Context) error {
			a.tgBot.Shutdown()
			return nil
		}),
		app.CloserFunc(func(ctx context.Context) error {
			return a.server.Shutdown(ctx)
		}),
	}
}
