package assembly

import (
	"context"
	"time"

	"dishes-service-backend/bot"
	bcontroller "dishes-service-backend/bot/controller"
	broutes "dishes-service-backend/bot/routes"
	bot_service "dishes-service-backend/bot/service"
	"dishes-service-backend/conf"
	"dishes-service-backend/controller"
	"dishes-service-backend/repository"
	"dishes-service-backend/routes"
	"dishes-service-backend/service"
	"dishes-service-backend/service/events"
	"dishes-service-backend/service/payment"
	"dishes-service-backend/service/payment/expiration"
	telegram_payment "dishes-service-backend/service/payment/telegram"
	"dishes-service-backend/transaction"

	"github.com/Falokut/go-kit/db"
	"github.com/Falokut/go-kit/http/client"
	"github.com/Falokut/go-kit/http/endpoint"
	"github.com/Falokut/go-kit/http/endpoint/hlog"
	"github.com/Falokut/go-kit/http/router"
	"github.com/Falokut/go-kit/log"
	"github.com/Falokut/go-kit/tg_botx"
	brouter "github.com/Falokut/go-kit/tg_botx/router"
	"github.com/pkg/errors"
	"github.com/txix-open/bgjob"
)

type DB interface {
	db.DB
	db.Transactional
}

type Locator struct {
	logger   *log.Adapter
	db       DB
	bgJobCli *bgjob.Client
	fileCli  *client.Client
	tgBot    *tg_botx.Bot
}

func NewLocator(
	db DB,
	bgJobCli *bgjob.Client,
	fileCli *client.Client,
	tgBot *tg_botx.Bot,
	logger *log.Adapter,
) Locator {
	return Locator{
		db:       db,
		bgJobCli: bgJobCli,
		fileCli:  fileCli,
		tgBot:    tgBot,
		logger:   logger,
	}
}

type Config struct {
	BotRouter  *brouter.Router
	HttpRouter *router.Router
	Workers    []*bgjob.Worker
}

// nolint:funlen
func (l Locator) LocatorConfig(ctx context.Context, cfg conf.Remote) (*Config, error) {
	txRunner := transaction.NewManager(l.db)

	userRepo := repository.NewUser(l.db)
	secretRepo := repository.NewSecret(cfg.App.AdminSecret)
	adminEvents := events.NewAdminEvents(l.tgBot)
	userService := service.NewUser(userRepo, txRunner, secretRepo, adminEvents)
	userBotContr := bcontroller.NewUser(userService)

	authService := service.NewAuth(cfg.Auth, cfg.Bot.Token, userRepo)
	authCtrl := controller.NewAuth(authService)

	fileRepo := repository.NewFile(l.fileCli, cfg.Images.BaseImagePath)
	dishRepo := repository.NewDish(l.db)
	dishService := service.NewDish(dishRepo, txRunner, fileRepo, l.logger)
	dishCtrl := controller.NewDish(dishService)

	dishesCategoriesRepo := repository.NewDishCategory(l.db)
	dishesCategoriesService := service.NewDishCategory(dishesCategoriesRepo)
	dishesCategoriesCtrl := controller.NewDishCategory(dishesCategoriesService)

	authMiddleware := routes.NewAuthMiddleware(cfg.Auth.Access.Secret)
	orderRepo := repository.NewOrder(l.db)
	paymentBot := bot.NewPaymentBot(cfg.Bot.PaymentToken, l.tgBot.Api(), orderRepo)
	telegramWorkerService := telegram_payment.NewWorker(paymentBot)
	telegramController := telegram_payment.NewWorkerController(telegramWorkerService)

	observer := payment.NewObserver(l.logger)
	telegramWorker := bgjob.NewWorker(
		l.bgJobCli,
		telegram_payment.WorkerQueue,
		telegramController,
		bgjob.WithPollInterval(5*time.Second), // nolint:mnd
		bgjob.WithObserver(observer),
	)

	paymentExpirationDelay := time.Minute * time.Duration(cfg.Payment.ExpirationDelayMinutes)
	expirationService := expiration.NewExpiration(l.bgJobCli, paymentExpirationDelay)
	expirationWorkerService := expiration.NewWorker(orderRepo)
	expirationController := expiration.NewWorkerController(expirationWorkerService)

	paymentMethods := payment.NewPaymentMethods(userRepo, l.bgJobCli)
	paymentService := payment.NewPayment(l.logger, paymentMethods, expirationService)

	orderService := service.NewOrder(paymentService, orderRepo, txRunner)
	orderCtrl := controller.NewOrder(orderService)

	restaurantRepo := repository.NewRestaurant(l.db)
	restaurantService := service.NewRestaurant(restaurantRepo)
	restaurantCtrl := controller.NewRestaurant(restaurantService)

	hrouter := routes.Router{
		Auth:         authCtrl,
		Dish:         dishCtrl,
		DishCategory: dishesCategoriesCtrl,
		Order:        orderCtrl,
		Restaurant:   restaurantCtrl,
	}

	orderUserService := bot_service.NewOrderUserService(l.tgBot, userRepo, orderRepo)
	orderCsvExporter := service.NewCsvOrderExporter(orderRepo)
	orderBotContrl := bcontroller.NewOrder(orderService, orderUserService, orderCsvExporter)
	botControllers := broutes.Controllers{
		User:  userBotContr,
		Order: orderBotContrl,
	}
	botAdminAuth := broutes.NewAdminAuth(userRepo)
	brouter := broutes.InitRoutes(
		botControllers,
		brouter.DefaultMiddlewares(l.logger),
		botAdminAuth.AdminAuth,
	)

	expirationWorker := bgjob.NewWorker(
		l.bgJobCli,
		expiration.WorkerQueue,
		expirationController,
		bgjob.WithPollInterval(5*time.Second), // nolint:mnd
		bgjob.WithObserver(observer),
	)
	err := broutes.RegisterRoutes(ctx, l.tgBot, userRepo)
	if err != nil {
		return nil, errors.WithMessage(err, "register bot routes")
	}
	return &Config{
		BotRouter:  brouter,
		HttpRouter: hrouter.Handler(authMiddleware, endpoint.DefaultWrapper(l.logger, hlog.Log(l.logger, true))),
		Workers: []*bgjob.Worker{
			telegramWorker,
			expirationWorker,
		},
	}, nil
}
