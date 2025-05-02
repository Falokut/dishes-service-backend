// nolint:noctx,funlen,dupl
package tests_test

import (
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

type DishCategorySuite struct {
	suite.Suite
	test             *test.Test
	adminAccessToken string

	db                 *dbt.TestDb
	dishCategoriesRepo repository.DishCategory
	dishRepo           repository.Dish
	restaurantRepo     repository.Restaurant
	cli                *client.Client
}

func TestDishCategory(t *testing.T) {
	t.Parallel()
	suite.Run(t, &DishCategorySuite{})
}

// nolint:dupl
func (t *DishCategorySuite) SetupTest() {
	test, _ := test.New(t.T())
	t.test = test
	t.db = dbt.New(test, db.WithMigrationRunner("../migrations", test.Logger()))
	t.dishCategoriesRepo = repository.NewDishCategory(t.db.Client)
	t.dishRepo = repository.NewDish(t.db.Client)
	t.restaurantRepo = repository.NewRestaurant(t.db.Client)

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
	err = t.db.Get(&userId,
		`INSERT INTO users(username,name,admin)
		VALUES($1,$2,$3)
		RETURNING id;`,
		"@test",
		"test",
		true,
	)
	t.Require().NoError(err)

	accessTokenTtl := time.Duration(cfg.Auth.Access.TtlHours) * time.Hour
	jwtGen, err := jwt.GenerateToken(cfg.Auth.Access.Secret, accessTokenTtl, &entity.TokenUserInfo{
		UserId:   userId,
		RoleName: domain.AdminRoleName,
	})
	t.Require().NoError(err)

	t.adminAccessToken = domain.BearerToken + " " + jwtGen.Token

	for _, category := range []string{
		"Горячее",
		"Холодное",
		"Напиток",
		"Острое",
		"Рыба",
		"Вегетарианское",
		"Мясное",
	} {
		_, err := t.dishCategoriesRepo.AddCategory(t.T().Context(), category)
		t.Require().NoError(err)
	}
}

func (t *DishCategorySuite) Test_GetAllCategories_HappyPath() {
	var categories []domain.DishCategory
	_, err := t.cli.Get("/dishes/all_categories").
		JsonResponseBody(&categories).
		Do(t.T().Context())
	t.Require().NoError(err)
	t.Require().ElementsMatch([]domain.DishCategory{
		{
			Id:   1,
			Name: "Горячее",
		},
		{
			Id:   2,
			Name: "Холодное",
		},
		{
			Id:   3,
			Name: "Напиток",
		},
		{
			Id:   4,
			Name: "Острое",
		},
		{
			Id:   5,
			Name: "Рыба",
		},
		{
			Id:   6,
			Name: "Вегетарианское",
		},
		{
			Id:   7,
			Name: "Мясное",
		},
	}, categories)
}

func (t *DishCategorySuite) Test_GetDishCategory_HappyPath() {
	var categories []domain.DishCategory
	_, err := t.cli.Get("/dishes/categories").
		JsonResponseBody(&categories).
		Do(t.T().Context())
	t.Require().NoError(err)
	t.Require().Empty(categories)

	restaurantId, err := t.restaurantRepo.InsertRestaurant(t.T().Context(), fake.It[string]())
	t.Require().NoError(err)

	dishId, err := t.dishRepo.InsertDish(t.T().Context(), &entity.InsertDish{
		Name:         "dish",
		Description:  "desc",
		ImageId:      "image_id",
		Price:        1000,
		RestaurantId: restaurantId,
	})
	t.Require().NoError(err)

	err = t.dishRepo.InsertDishCategories(t.T().Context(), dishId, []int32{1, 2})
	t.Require().NoError(err)

	_, err = t.cli.Get("/dishes/categories").
		JsonResponseBody(&categories).
		Do(t.T().Context())
	t.Require().NoError(err)
	t.Require().ElementsMatch([]domain.DishCategory{
		{
			Id:   1,
			Name: "Горячее",
		},
		{
			Id:   2,
			Name: "Холодное",
		},
	}, categories)
}

func (t *DishCategorySuite) Test_GetCategory_HappyPath() {
	const categoryId = 3
	var category domain.DishCategory
	_, err := t.cli.Get(fmt.Sprintf("/dishes/categories/%d", categoryId)).
		JsonResponseBody(&category).
		Do(t.T().Context())
	t.Require().NoError(err)
	t.Require().Equal(domain.DishCategory{Id: categoryId, Name: "Напиток"}, category)
}

func (t *DishCategorySuite) Test_AddCategory_HappyPath() {
	categoryName := fake.It[string]()
	req := domain.AddCategoryRequest{Name: categoryName}
	var resp domain.AddCategoryResponse
	_, err := t.cli.Post("/dishes/categories").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(req).
		JsonResponseBody(&resp).
		Do(t.T().Context())
	t.Require().NoError(err)

	category, err := t.dishCategoriesRepo.GetCategory(t.T().Context(), resp.Id)
	t.Require().NoError(err)
	t.Require().Equal(categoryName, category.Name)
}

func (t *DishCategorySuite) Test_DeleteCategory_HappyPath() {
	categoryName := fake.It[string]()
	req := domain.AddCategoryRequest{Name: categoryName}
	var resp domain.AddCategoryResponse
	_, err := t.cli.Post("/dishes/categories").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(req).
		JsonResponseBody(&resp).
		Do(t.T().Context())
	t.Require().NoError(err)

	category, err := t.dishCategoriesRepo.GetCategory(t.T().Context(), resp.Id)
	t.Require().NoError(err)
	t.Require().Equal(categoryName, category.Name)

	err = t.cli.Delete(fmt.Sprintf("dishes/categories/%d", resp.Id)).
		Header(domain.AuthHeaderName, t.adminAccessToken).
		DoWithoutResponse(t.T().Context())
	t.Require().NoError(err)

	getResp, err := t.cli.Get(fmt.Sprintf("/dishes/categories/%d", resp.Id)).
		Do(t.T().Context())
	t.Require().NoError(err)
	t.Require().NoError(err)
	t.Require().EqualValues(http.StatusNotFound, getResp.StatusCode())

	respBody, err := getResp.Body()
	t.Require().NoError(err)

	var errorResp apierrors.Error
	err = json.Unmarshal(respBody, &errorResp)
	t.Require().NoError(err)
	t.Require().EqualValues(domain.ErrCodeDishCategoryNotFound, errorResp.ErrorCode)
}

func (t *DishCategorySuite) Test_RenameCategory_HappyPath() {
	categoryName := fake.It[string]()
	req := domain.AddCategoryRequest{Name: categoryName}
	var resp domain.AddCategoryResponse
	_, err := t.cli.Post("/dishes/categories").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(req).
		JsonResponseBody(&resp).
		Do(t.T().Context())
	t.Require().NoError(err)

	category, err := t.dishCategoriesRepo.GetCategory(t.T().Context(), resp.Id)
	t.Require().NoError(err)
	t.Require().Equal(categoryName, category.Name)

	newCategoryName := fake.It[string]()
	renameReq := domain.RenameCategoryRequest{Name: newCategoryName}
	_, err = t.cli.Post(fmt.Sprintf("/dishes/categories/%d", resp.Id)).
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(renameReq).
		StatusCodeToError().
		Do(t.T().Context())
	t.Require().NoError(err)

	category, err = t.dishCategoriesRepo.GetCategory(t.T().Context(), resp.Id)
	t.Require().NoError(err)
	t.Require().Equal(newCategoryName, category.Name)
}

func (t *DishCategorySuite) Test_RenameCategory_Conflict() {
	categoryName := fake.It[string]()
	req := domain.AddCategoryRequest{Name: categoryName}
	var resp domain.AddCategoryResponse
	_, err := t.cli.Post("/dishes/categories").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(req).
		JsonResponseBody(&resp).
		Do(t.T().Context())
	t.Require().NoError(err)

	category, err := t.dishCategoriesRepo.GetCategory(t.T().Context(), resp.Id)
	t.Require().NoError(err)
	t.Require().Equal(categoryName, category.Name)

	categoryName = fake.It[string]()
	_, err = t.cli.Post("/dishes/categories").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(domain.AddCategoryRequest{Name: categoryName}).
		Do(t.T().Context())
	t.Require().NoError(err)

	renameReq := domain.RenameCategoryRequest{Name: categoryName}
	renameResp, err := t.cli.Post(fmt.Sprintf("/dishes/categories/%d", resp.Id)).
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(renameReq).
		Do(t.T().Context())
	t.Require().NoError(err)

	_, err = t.dishCategoriesRepo.GetCategory(t.T().Context(), resp.Id)
	t.Require().NoError(err)
	t.Require().EqualValues(http.StatusConflict, renameResp.StatusCode())
}
