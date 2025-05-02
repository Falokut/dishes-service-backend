//nolint:noctx,funlen
package tests_test

import (
	"dishes-service-backend/assembly"
	"dishes-service-backend/conf"
	"dishes-service-backend/entity"
	"time"

	"dishes-service-backend/repository"
	"testing"

	"github.com/Falokut/go-kit/test/fake"
	"github.com/Falokut/go-kit/test/tgt"
	"github.com/Falokut/go-kit/tg_bot"
	"github.com/google/uuid"

	"github.com/Falokut/go-kit/db"
	"github.com/Falokut/go-kit/test"
	"github.com/Falokut/go-kit/test/dbt"
	"github.com/stretchr/testify/suite"
	"github.com/txix-open/bgjob"
)

type UserBotSuite struct {
	suite.Suite
	test          *test.Test
	db            *db.Client
	userRepo      repository.User
	botServerMock *tgt.BotServerMock
}

func TestUserBot(t *testing.T) {
	t.Parallel()
	suite.Run(t, &UserBotSuite{})
}

func (t *UserBotSuite) SetupTest() {
	test, _ := test.New(t.T())
	t.test = test
	db := dbt.New(test, db.WithMigrationRunner("../migrations", test.Logger()))
	t.userRepo = repository.NewUser(db.Client)
	t.db = db.Client

	bgjobDb := bgjob.NewPgStore(t.db.DB.DB)
	bgjobCli := bgjob.NewClient(bgjobDb)
	tgBot, botMockServerMock := tgt.TestBot(t.test)

	locator := assembly.NewLocator(t.db, bgjobCli, nil, tgBot, t.test.Logger())
	locatorCfg, err := locator.LocatorConfig(t.T().Context(), conf.Remote{App: conf.App{AdminSecret: "secret"}})
	t.Require().NoError(err)

	tgBot.UpgradeMux(t.T().Context(), locatorCfg.BotRouter)
	t.botServerMock = botMockServerMock
}

func (t *UserBotSuite) Test_RegisterUser_HappyPath() {
	username := fake.It[string]()
	firstName := fake.It[string]()
	telegramId := fake.It[int64]()
	updates := []tg_bot.Update{
		{
			UpdateId: 1,
			Message: &tg_bot.Message{
				Entities: []tg_bot.MessageEntity{
					{
						Type:   "bot_command",
						Length: len("/start"),
					},
				},
				Text: "/start",
				Chat: &tg_bot.Chat{
					Id: telegramId,
				},
				From: &tg_bot.User{
					UserName:  username,
					FirstName: firstName,
					Id:        telegramId,
				},
			},
		},
	}
	t.botServerMock.SetUpdates(updates)
	time.Sleep(time.Second)
	userId, err := t.userRepo.GetUserIdByTelegramId(t.T().Context(), telegramId)
	t.Require().NoError(err)

	user, err := t.userRepo.GetUserInfo(t.T().Context(), userId)
	t.Require().NoError(err)
	t.Require().Equal(username, user.Username)
	t.Require().Equal(firstName, user.Name)
	t.Require().False(user.Admin)
}

func (t *UserBotSuite) Test_PassBySecret_HappyPath() {
	username := fake.It[string]()
	firstName := fake.It[string]()
	telegramId := fake.It[int64]()
	updates := []tg_bot.Update{
		{
			UpdateId: 1,
			Message: &tg_bot.Message{
				Entities: []tg_bot.MessageEntity{
					{
						Type:   "bot_command",
						Length: len("/start"),
					},
				},
				Text: "/start",
				Chat: &tg_bot.Chat{
					Id: telegramId,
				},
				From: &tg_bot.User{
					UserName:  username,
					FirstName: firstName,
					Id:        telegramId,
				},
			},
		},
		{
			UpdateId: 2,
			Message: &tg_bot.Message{
				Entities: []tg_bot.MessageEntity{
					{
						Type:   "bot_command",
						Length: len("/pass_by_secret"),
					},
				},
				Text: "/pass_by_secret secret",
				Chat: &tg_bot.Chat{
					Id: telegramId,
				},
				From: &tg_bot.User{
					UserName:  username,
					FirstName: firstName,
					Id:        telegramId,
				},
			},
		},
	}
	t.botServerMock.SetUpdates(updates)
	time.Sleep(time.Second)
	userId, err := t.userRepo.GetUserIdByTelegramId(t.T().Context(), telegramId)
	t.Require().NoError(err)

	user, err := t.userRepo.GetUserInfo(t.T().Context(), userId)
	t.Require().NoError(err)
	t.Require().Equal(username, user.Username)
	t.Require().Equal(firstName, user.Name)
	t.Require().True(user.Admin)
}

func (t *UserBotSuite) Test_RemoveAdminByUsername_HappyPath() {
	firstUserId := uuid.NewString()
	firstChatId := fake.It[int64]()
	firstUsername := fake.It[string]()
	userId, err := t.userRepo.InsertUser(t.T().Context(), entity.InsertUser{
		Id:       firstUserId,
		Username: firstUsername,
		Name:     fake.It[string](),
	})
	t.Require().NoError(err)

	err = t.userRepo.InsertUserTelegram(t.T().Context(), userId, entity.Telegram{
		ChatId: firstChatId,
		UserId: firstChatId,
	})
	t.Require().NoError(err)

	err = t.userRepo.SetUserAdminStatus(t.T().Context(), firstUsername, true)
	t.Require().NoError(err)

	secondUserId := uuid.NewString()
	secondChatId := fake.It[int64]()
	secondUsername := fake.It[string]()
	userId, err = t.userRepo.InsertUser(t.T().Context(), entity.InsertUser{
		Id:       secondUserId,
		Username: secondUsername,
		Name:     fake.It[string](),
	})
	t.Require().NoError(err)

	err = t.userRepo.InsertUserTelegram(t.T().Context(), userId, entity.Telegram{
		ChatId: secondChatId,
		UserId: secondChatId,
	})
	t.Require().NoError(err)

	t.botServerMock.SetUpdates([]tg_bot.Update{
		{
			UpdateId: 1,
			Message: &tg_bot.Message{
				Entities: []tg_bot.MessageEntity{
					{
						Type:   "bot_command",
						Length: len("/remove_admin"),
					},
				},
				Text: "/remove_admin " + secondUsername,
				Chat: &tg_bot.Chat{
					Id: firstChatId,
				},
				From: &tg_bot.User{
					UserName: firstUsername,
					Id:       firstChatId,
				},
			},
		},
	})
	time.Sleep(time.Second)

	user, err := t.userRepo.GetUserInfo(t.T().Context(), secondUserId)
	t.Require().NoError(err)
	t.Require().False(user.Admin)
}
