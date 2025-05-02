// //nolint:noctx,funlen,dupl
package tests_test

import (
	"database/sql"
	"dishes-service-backend/assembly"
	"dishes-service-backend/conf"
	"dishes-service-backend/domain"
	"dishes-service-backend/entity"
	"dishes-service-backend/repository"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Falokut/go-kit/db"
	"github.com/Falokut/go-kit/http/client"
	"github.com/Falokut/go-kit/jwt"
	"github.com/Falokut/go-kit/test"
	"github.com/Falokut/go-kit/test/dbt"
	"github.com/Falokut/go-kit/test/fake"
	"github.com/Falokut/go-kit/test/tgt"
	"github.com/stretchr/testify/suite"
	"github.com/txix-open/bgjob"
)

type DishSuite struct {
	suite.Suite
	test             *test.Test
	adminAccessToken string
	userAccessToken  string

	restaurantName string
	restaurantId   int32

	db       *dbt.TestDb
	dishRepo repository.Dish
	cli      *client.Client
}

func TestDish(t *testing.T) {
	t.Parallel()
	suite.Run(t, &DishSuite{})
}

func (t *DishSuite) SetupTest() {
	test, _ := test.New(t.T())
	t.test = test
	t.db = dbt.New(test, db.WithMigrationRunner("../migrations", test.Logger()))
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
		"@admin",
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

	dishCategoriesRepo := repository.NewDishCategory(t.db.Client)
	for _, category := range []string{
		"Горячее",
		"Холодное",
		"Напиток",
		"Острое",
		"Рыба",
		"Вегетарианское",
		"Мясное",
	} {
		_, err := dishCategoriesRepo.AddCategory(t.T().Context(), category)
		t.Require().NoError(err)
	}

	dishRestaurantsRepo := repository.NewRestaurant(t.db.Client)
	t.restaurantName = fake.It[string]()
	t.restaurantId, err = dishRestaurantsRepo.InsertRestaurant(t.T().Context(), t.restaurantName)
	t.Require().NoError(err)
}

func (t *DishSuite) Test_List_ByLimitOffset_HappyPath() {
	var addDish = entity.InsertDish{
		Name:         fake.It[string](),
		Description:  fake.It[string](),
		Price:        350,
		ImageId:      fake.It[string](),
		RestaurantId: t.restaurantId,
	}
	dishId, err := t.dishRepo.InsertDish(t.T().Context(), &addDish)
	t.Require().NoError(err)

	err = t.dishRepo.InsertDishCategories(t.T().Context(), dishId, []int32{1, 2})
	t.Require().NoError(err)

	var dishes []domain.Dish
	_, err = t.cli.Get("/dishes").
		QueryParams(map[string]any{
			"limit":  1,
			"offset": 0,
		}).
		JsonResponseBody(&dishes).
		Do(t.T().Context())
	t.Require().NoError(err)
	t.Require().Len(dishes, 1)

	dish := dishes[0]
	t.Require().Equal(addDish.Name, dish.Name)
	t.Require().Equal(addDish.Description, dish.Description)
	t.Require().Equal("my_image_path/image-dish/"+addDish.ImageId, dish.Url)
	t.Require().ElementsMatch([]string{"Горячее", "Холодное"}, dish.Categories)
	t.Require().Equal(t.restaurantName, dish.RestaurantName)
}

func (t *DishSuite) Test_List_ByIds_HappyPath() {
	var addDish = entity.InsertDish{
		Name:         fake.It[string](),
		Description:  fake.It[string](),
		Price:        350,
		ImageId:      fake.It[string](),
		RestaurantId: t.restaurantId,
	}
	t.insertDishWithCategories(addDish, []int32{1, 4})

	addDish = entity.InsertDish{
		Name:         fake.It[string](),
		Description:  fake.It[string](),
		Price:        544,
		ImageId:      fake.It[string](),
		RestaurantId: t.restaurantId,
	}
	id := t.insertDishWithCategories(addDish, []int32{3, 2})

	var dishes []domain.Dish
	_, err := t.cli.Get("/dishes").
		JsonResponseBody(&dishes).
		QueryParams(map[string]any{
			"ids": fmt.Sprint(id),
		}).
		StatusCodeToError().
		Do(t.T().Context())
	t.Require().NoError(err)

	t.Require().Len(dishes, 1)
	dish := dishes[0]

	t.Require().EqualValues(id, dish.Id)
	t.Require().Equal(addDish.Name, dish.Name)
	t.Require().Equal(addDish.Description, dish.Description)
	t.Require().Equal("my_image_path/image-dish/"+addDish.ImageId, dish.Url)
	t.Require().ElementsMatch([]string{"Напиток", "Холодное"}, dish.Categories)
	t.Require().Equal(t.restaurantName, dish.RestaurantName)
}

func (t *DishSuite) Test_List_ByCategories_HappyPath() {
	addDish := entity.InsertDish{
		Name:         fake.It[string](),
		Description:  fake.It[string](),
		Price:        350,
		ImageId:      fake.It[string](),
		RestaurantId: t.restaurantId,
	}
	t.insertDishWithCategories(addDish, []int32{4, 5})

	addDish = entity.InsertDish{
		Name:         fake.It[string](),
		Description:  fake.It[string](),
		Price:        544,
		ImageId:      fake.It[string](),
		RestaurantId: t.restaurantId,
	}
	t.insertDishWithCategories(addDish, []int32{4, 2})

	var dishes []domain.Dish
	_, err := t.cli.Get("/dishes").
		JsonResponseBody(&dishes).
		QueryParams(map[string]any{
			"categoriesIds": "2,4",
			"limit":         2,
		}).
		StatusCodeToError().
		Do(t.T().Context())
	t.Require().NoError(err)
	t.Require().Len(dishes, 1)

	dish := dishes[0]
	t.Require().EqualValues(2, dish.Id)
	t.Require().Equal(addDish.Name, dish.Name)
	t.Require().Equal(addDish.Description, dish.Description)
	t.Require().Equal("my_image_path/image-dish/"+addDish.ImageId, dish.Url)
	t.Require().ElementsMatch([]string{"Острое", "Холодное"}, dish.Categories)

	_, err = t.cli.Get("/dishes").
		JsonResponseBody(&dishes).
		QueryParams(map[string]any{
			"categoriesIds": "6,2",
			"limit":         2,
		}).
		Do(t.T().Context())
	t.Require().NoError(err)
	t.Require().Empty(dishes)
}

func (t *DishSuite) Test_AddDish_HappyPath() {
	req := domain.AddDishRequest{
		Name:         fake.It[string](),
		Description:  fake.It[string](),
		Price:        900,
		Categories:   []int32{5, 6},
		RestaurantId: t.restaurantId,
	}
	resp := domain.AddDishResponse{}
	_, err := t.cli.Post("/dishes").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(req).
		JsonResponseBody(&resp).
		Do(t.T().Context())
	t.Require().NoError(err)

	dishes, err := t.dishRepo.GetDishesByIds(t.T().Context(), []int32{resp.Id})
	t.Require().NoError(err)
	t.Require().Len(dishes, 1)

	dish := dishes[0]
	t.Require().Equal(resp.Id, dish.Id)
	t.Require().Equal(req.Name, dish.Name)
	t.Require().Equal(req.Price, dish.Price)
	t.Require().Empty(dish.ImageId)

	var categoriesIds []int32
	t.db.Must().SelectContext(t.T().Context(), &categoriesIds, "SELECT category_id FROM dish_categories WHERE dish_id=$1", resp.Id)
	t.Require().ElementsMatch(req.Categories, categoriesIds)
}

func (t *DishSuite) Test_EditDish_HappyPath() {
	req := domain.AddDishRequest{
		Name:         fake.It[string](),
		Description:  fake.It[string](),
		Price:        900,
		Categories:   []int32{5, 6},
		RestaurantId: t.restaurantId,
	}
	resp := domain.AddDishResponse{}
	_, err := t.cli.Post("/dishes").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(req).
		JsonResponseBody(&resp).
		Do(t.T().Context())
	t.Require().NoError(err)

	dishes, err := t.dishRepo.GetDishesByIds(t.T().Context(), []int32{resp.Id})
	t.Require().NoError(err)
	t.Require().Len(dishes, 1)

	dish := dishes[0]
	t.Require().Equal(resp.Id, dish.Id)
	t.Require().Equal(req.Name, dish.Name)
	t.Require().Equal(req.Price, dish.Price)
	t.Require().Empty(dish.ImageId)

	var categoriesIds []int32
	t.db.Must().SelectContext(t.T().Context(), &categoriesIds, "SELECT category_id FROM dish_categories WHERE dish_id=$1", resp.Id)
	t.Require().ElementsMatch(req.Categories, categoriesIds)

	editDishReq := domain.EditDishRequest{
		Name:         fake.It[string](),
		Description:  fake.It[string](),
		Price:        800,
		Categories:   []int32{1, 2, 5},
		RestaurantId: t.restaurantId,
	}

	_, err = t.cli.Post(fmt.Sprintf("dishes/edit/%d", resp.Id)).
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(editDishReq).
		StatusCodeToError().
		Do(t.T().Context())
	t.Require().NoError(err)

	dish = entity.Dish{}
	t.db.Must().SelectRow(t.T().Context(), &dish, "SELECT id, name, description, price FROM dish WHERE id=$1", resp.Id)

	expectedDish := entity.Dish{
		Id:          resp.Id,
		Name:        editDishReq.Name,
		Description: editDishReq.Description,
		Price:       editDishReq.Price,
	}
	t.Require().Equal(expectedDish, dish)

	t.db.Must().SelectContext(t.T().Context(), &categoriesIds, "SELECT category_id FROM dish_categories WHERE dish_id=$1", expectedDish.Id)
	t.Require().ElementsMatch(editDishReq.Categories, categoriesIds)
}

func (t *DishSuite) Test_DeleteDish_HappyPath() {
	req := domain.AddDishRequest{
		Name:         fake.It[string](),
		Description:  fake.It[string](),
		Price:        900,
		Categories:   []int32{5, 6},
		RestaurantId: t.restaurantId,
	}
	_, err := t.cli.Post("/dishes").
		Header(domain.AuthHeaderName, t.adminAccessToken).
		JsonRequestBody(req).
		Do(t.T().Context())
	t.Require().NoError(err)

	var ids []int32
	t.db.Must().SelectContext(t.T().Context(), &ids,
		"SELECT id FROM dish WHERE name=$1 AND description=$2 AND price=$3 AND image_id=$4",
		req.Name, req.Description, req.Price, "")
	t.Require().Len(ids, 1)

	var categoriesIds []int32
	t.db.Must().SelectContext(t.T().Context(), &categoriesIds, "SELECT category_id FROM dish_categories WHERE dish_id=$1", ids[0])
	t.Require().ElementsMatch(req.Categories, categoriesIds)

	err = t.cli.Delete(fmt.Sprintf("dishes/delete/%d", ids[0])).
		Header(domain.AuthHeaderName, t.adminAccessToken).
		DoWithoutResponse(t.T().Context())
	t.Require().NoError(err)

	var dish entity.Dish
	err = t.db.Get(&dish, "SELECT id, name, description, price FROM dish WHERE id=$1", ids[0])
	t.Require().ErrorIs(err, sql.ErrNoRows)

	t.db.Must().SelectContext(t.T().Context(), &categoriesIds, "SELECT category_id FROM dish_categories WHERE dish_id=$1", ids[0])
	t.Require().ElementsMatch([]int32{}, categoriesIds)
}

func (t *DishSuite) Test_AddDish_Forbidden() {
	addDishReq := domain.AddDishRequest{
		Name:         fake.It[string](),
		Description:  fake.It[string](),
		Price:        800,
		Categories:   []int32{5, 6},
		RestaurantId: t.restaurantId,
	}

	resp, err := t.cli.Post("/dishes").
		JsonRequestBody(addDishReq).
		Header(domain.AuthHeaderName, t.userAccessToken).
		Do(t.T().Context())
	t.Require().NoError(err)
	t.Require().Equal(http.StatusForbidden, resp.StatusCode())

	var ids []int32
	t.db.Must().SelectContext(t.T().Context(), &ids,
		"SELECT id FROM dish WHERE name=$1 AND description=$2 AND price=$3 AND image_id=$4",
		addDishReq.Name, addDishReq.Description, addDishReq.Price, "")
	t.Require().Empty(ids)
}

func (t *DishSuite) insertDishWithCategories(dish entity.InsertDish, categories []int32) int32 {
	dishId, err := t.dishRepo.InsertDish(t.T().Context(), &dish)
	t.Require().NoError(err)

	err = t.dishRepo.InsertDishCategories(t.T().Context(), dishId, categories)
	t.Require().NoError(err)
	return dishId
}

func getConfig() conf.Remote {
	return conf.Remote{
		Images: conf.Images{
			BaseImagePath: "my_image_path",
		},
		App: conf.App{
			AdminSecret: "secret",
		},
		Auth: conf.Auth{
			Access: conf.JwtToken{
				TtlHours: 1,
				Secret:   fake.It[string](),
			},
		},
	}
}
