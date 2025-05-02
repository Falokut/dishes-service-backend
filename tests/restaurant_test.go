// nolint:noctx,funlen,dupl
package tests_test

import (
	"context"
	"dishes-service-backend/assembly"
	"dishes-service-backend/domain"
	"dishes-service-backend/entity"
	"dishes-service-backend/repository"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Falokut/go-kit/db"
	"github.com/Falokut/go-kit/http/apierrors"
	"github.com/Falokut/go-kit/http/client"
	"github.com/Falokut/go-kit/json"
	"github.com/Falokut/go-kit/jwt"
	"github.com/Falokut/go-kit/test"
	"github.com/Falokut/go-kit/test/dbt"
	"github.com/Falokut/go-kit/test/fake"
	"github.com/Falokut/go-kit/test/tgt"
	"github.com/stretchr/testify/suite"
	"github.com/txix-open/bgjob"
)

type RestaurantSuite struct {
	suite.Suite
	test             *test.Test
	adminAccessToken string

	db              *dbt.TestDb
	restaurantsRepo repository.Restaurant
	dishRepo        repository.Dish
	cli             *client.Client
}

func TestDishRestaurants(t *testing.T) {
	t.Parallel()
	suite.Run(t, &RestaurantSuite{})
}

// nolint:dupl
func (t *RestaurantSuite) SetupTest() {
	test, _ := test.New(t.T())
	t.test = test
	t.db = dbt.New(test, db.WithMigrationRunner("../migrations", test.Logger()))
	t.restaurantsRepo = repository.NewRestaurant(t.db.Client)
	t.dishRepo = repository.NewDish(t.db.Client)

	bgjobDb := bgjob.NewPgStore(t.db.Client.DB.DB)
	bgjobCli := bgjob.NewClient(bgjobDb)
	tgBot, _ := tgt.TestBot(test)

	cfg := getConfig()
	locator := assembly.NewLocator(t.db, bgjobCli, nil, tgBot, t.test.Logger())
	locatorCfg, err := locator.LocatorConfig(t.T().Context(), cfg)
	t.Require().NoError(err)

	server := httptest.NewServer(locatorCfg.HttpRouter)
	t.cli = client.NewWithClient(server.Client())
	t.cli.GlobalRequestConfig().BaseUrl = fmt.Sprintf("http://%s", server.Listener.Addr())

	var userId string
	t.db.Must().SelectRow(t.T().Context(),
		&userId,
		`INSERT INTO users(username,name,admin)
		VALUES($1,$2,$3)
		RETURNING id;`,
		"@test",
		"test",
		true,
	)

	accessTokenTtl := time.Hour * time.Duration(cfg.Auth.Access.TtlHours)
	jwtGen, err := jwt.GenerateToken(cfg.Auth.Access.Secret, accessTokenTtl, &entity.TokenUserInfo{
		UserId:   userId,
		RoleName: domain.AdminRoleName,
	})
	t.Require().NoError(err)

	t.adminAccessToken = domain.BearerToken + " " + jwtGen.Token

	for _, category := range []string{
		"Додо",
		"Вкусно",
	} {
		_, err := t.restaurantsRepo.InsertRestaurant(context.Background(), category)
		t.Require().NoError(err)
	}
}

func (t *RestaurantSuite) Test_GetAllRestaurants_HappyPath() {
	var restaurants []domain.Restaurant
	_, err := t.cli.Get("/restaurants").
		JsonResponseBody(&restaurants).
		Do(context.Background())
	t.Require().NoError(err)

	expectedRestaurants := []domain.Restaurant{
		{
			Id:   1,
			Name: "Додо",
		},
		{
			Id:   2,
			Name: "Вкусно",
		},
	}
	t.Require().ElementsMatch(expectedRestaurants, restaurants)
}

func (t *RestaurantSuite) Test_GetRestaurant_HappyPath() {
	const restaurantId = 2
	var category domain.Restaurant
	_, err := t.cli.Get(fmt.Sprintf("/restaurants/%d", restaurantId)).
		JsonResponseBody(&category).
		Do(context.Background())
	t.Require().NoError(err)
	t.Require().Equal(domain.Restaurant{Id: restaurantId, Name: "Вкусно"}, category)
}

func (t *RestaurantSuite) Test_InsertRestaurant_HappyPath() {
	categoryName := fake.It[string]()
	req := domain.AddRestaurantRequest{Name: categoryName}
	var resp domain.AddRestaurantResponse
	_, err := t.cli.Post("/restaurants").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(req).
		JsonResponseBody(&resp).
		Do(context.Background())
	t.Require().NoError(err)

	category, err := t.restaurantsRepo.GetRestaurant(context.Background(), resp.Id)
	t.Require().NoError(err)
	t.Require().Equal(categoryName, category.Name)
}

func (t *RestaurantSuite) Test_DeleteRestaurant_HappyPath() {
	categoryName := fake.It[string]()
	req := domain.AddRestaurantRequest{Name: categoryName}
	var resp domain.AddRestaurantResponse
	_, err := t.cli.Post("/restaurants").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(req).
		JsonResponseBody(&resp).
		Do(context.Background())
	t.Require().NoError(err)

	category, err := t.restaurantsRepo.GetRestaurant(context.Background(), resp.Id)
	t.Require().NoError(err)
	t.Require().Equal(categoryName, category.Name)

	err = t.cli.Delete(fmt.Sprintf("restaurants/%d", resp.Id)).
		Header(domain.AuthHeaderName, t.adminAccessToken).
		DoWithoutResponse(context.Background())
	t.Require().NoError(err)

	getResp, err := t.cli.Get(fmt.Sprintf("/restaurants/%d", resp.Id)).
		Do(context.Background())
	t.Require().NoError(err)
	t.Require().EqualValues(http.StatusNotFound, getResp.StatusCode())

	respBody, err := getResp.Body()
	t.Require().NoError(err)

	var errorResp apierrors.Error
	err = json.Unmarshal(respBody, &errorResp)
	t.Require().NoError(err)
	t.Require().EqualValues(domain.ErrCodeRestaurantNotFound, errorResp.ErrorCode)
}

func (t *RestaurantSuite) Test_RenameRestaurant_HappyPath() {
	categoryName := fake.It[string]()
	req := domain.AddRestaurantRequest{Name: categoryName}
	var resp domain.AddRestaurantResponse
	_, err := t.cli.Post("/restaurants").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(req).
		JsonResponseBody(&resp).
		Do(context.Background())
	t.Require().NoError(err)

	category, err := t.restaurantsRepo.GetRestaurant(context.Background(), resp.Id)
	t.Require().NoError(err)
	t.Require().Equal(categoryName, category.Name)

	newRestaurantName := fake.It[string]()
	renameReq := domain.RenameRestaurantRequest{Name: newRestaurantName}
	_, err = t.cli.Post(fmt.Sprintf("/restaurants/%d", resp.Id)).
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(renameReq).
		StatusCodeToError().
		Do(context.Background())
	t.Require().NoError(err)

	category, err = t.restaurantsRepo.GetRestaurant(context.Background(), resp.Id)
	t.Require().NoError(err)
	t.Require().Equal(newRestaurantName, category.Name)
}

func (t *RestaurantSuite) Test_RenameRestaurant_Conflict() {
	categoryName := fake.It[string]()
	req := domain.AddRestaurantRequest{Name: categoryName}
	var resp domain.AddRestaurantResponse
	_, err := t.cli.Post("/restaurants").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(req).
		JsonResponseBody(&resp).
		Do(context.Background())
	t.Require().NoError(err)

	category, err := t.restaurantsRepo.GetRestaurant(context.Background(), resp.Id)
	t.Require().NoError(err)
	t.Require().Equal(categoryName, category.Name)

	categoryName = fake.It[string]()
	_, err = t.cli.Post("/restaurants").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(domain.AddRestaurantRequest{Name: categoryName}).
		Do(context.Background())
	t.Require().NoError(err)

	renameReq := domain.RenameRestaurantRequest{Name: categoryName}
	renameResp, err := t.cli.Post(fmt.Sprintf("/restaurants/%d", resp.Id)).
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(renameReq).
		Do(context.Background())
	t.Require().NoError(err)

	_, err = t.restaurantsRepo.GetRestaurant(context.Background(), resp.Id)
	t.Require().NoError(err)
	t.Require().EqualValues(http.StatusConflict, renameResp.StatusCode())
}
