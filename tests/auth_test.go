//nolint:noctx,funlen
package tests_test

import (
	"dishes-service-backend/assembly"
	"dishes-service-backend/domain"
	"dishes-service-backend/entity"
	"time"

	"dishes-service-backend/repository"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/Falokut/go-kit/jwt"
	"github.com/Falokut/go-kit/test/tgt"

	"github.com/Falokut/go-kit/db"
	"github.com/Falokut/go-kit/http/client"
	"github.com/Falokut/go-kit/test"
	"github.com/Falokut/go-kit/test/dbt"
	"github.com/stretchr/testify/suite"
	"github.com/txix-open/bgjob"
)

type AuthSuite struct {
	suite.Suite
	test             *test.Test
	adminAccessToken string
	userAccessToken  string

	db       *dbt.TestDb
	userRepo repository.User
	cli      *client.Client
}

func TestAuth(t *testing.T) {
	t.Parallel()
	suite.Run(t, &AuthSuite{})
}

// nolint:dupl
func (t *AuthSuite) SetupTest() {
	test, _ := test.New(t.T())
	t.test = test
	t.db = dbt.New(test, db.WithMigrationRunner("../migrations", test.Logger()))
	t.userRepo = repository.NewUser(t.db)

	bgjobDb := bgjob.NewPgStore(t.db.Client.DB.DB)
	bgjobCli := bgjob.NewClient(bgjobDb)
	tgBot, _ := tgt.TestBot(test)

	cfg := getConfig()
	locator := assembly.NewLocator(t.db, bgjobCli, nil, tgBot, t.test.Logger())
	locatorCfg, err := locator.LocatorConfig(t.T().Context(), cfg)
	t.Require().NoError(err)

	server := httptest.NewServer(locatorCfg.HttpRouter)
	t.cli = client.NewWithClient(server.Client())
	t.cli.GlobalRequestConfig().BaseUrl = fmt.Sprintf("http://%s", server.Listener.Addr().String())

	var userId string
	t.db.Must().SelectRow(t.T().Context(),
		&userId,
		`INSERT INTO users(username,name,admin)
		VALUES($1,$2,$3)
		RETURNING id;`,
		"@admin",
		"test",
		true,
	)

	accessTokenTtl := time.Duration(cfg.Auth.Access.TtlHours) * time.Hour
	jwtGen, err := jwt.GenerateToken(cfg.Auth.Access.Secret, accessTokenTtl, &entity.TokenUserInfo{
		UserId:   userId,
		RoleName: domain.AdminRoleName,
	})
	t.Require().NoError(err)

	t.adminAccessToken = domain.BearerToken + " " + jwtGen.Token

	t.db.Must().SelectRow(t.T().Context(),
		&userId,
		`INSERT INTO users(username,name,admin)
		VALUES($1,$2,$3)
		RETURNING id;`,
		"@user",
		"test",
		false,
	)
	jwtGen, err = jwt.GenerateToken(cfg.Auth.Access.Secret, accessTokenTtl, &entity.TokenUserInfo{
		UserId:   userId,
		RoleName: domain.UserRoleName,
	})
	t.Require().NoError(err)

	t.userAccessToken = domain.BearerToken + " " + jwtGen.Token
}

func (t *AuthSuite) Test_UserRole_HappyPath() {
	var resp domain.UserRoleResponse
	_, err := t.cli.Get("/auth/user_role").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		StatusCodeToError().
		JsonResponseBody(&resp).
		Do(t.T().Context())
	t.Require().NoError(err)
	t.Require().Equal(domain.AdminRoleName, resp.RoleName)

	_, err = t.cli.Get("/auth/user_role").
		Header(domain.AuthHeaderName, t.userAccessToken).
		StatusCodeToError().
		JsonResponseBody(&resp).
		Do(t.T().Context())
	t.Require().NoError(err)
	t.Require().Equal(domain.UserRoleName, resp.RoleName)
}
