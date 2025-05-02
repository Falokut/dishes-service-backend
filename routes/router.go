// nolint:funlen
package routes

import (
	"dishes-service-backend/controller"
	"net/http"

	"github.com/Falokut/go-kit/cluster"
	"github.com/Falokut/go-kit/http/endpoint"
	"github.com/Falokut/go-kit/http/router"
)

const (
	withAdminAuthKey = "extra-with-admin-auth"
	withUserAuthKey  = "extra-with-user-auth"
)

type Router struct {
	Auth         controller.Auth
	Dish         controller.Dish
	DishCategory controller.DishCategory
	Order        controller.Order
	Restaurant   controller.Restaurant
}

func (r Router) Handler(authMiddleware AuthMiddleware, wrapper endpoint.Wrapper) *router.Router {
	mux := router.New()
	for _, desc := range EndpointDescriptors(r) {
		endpointWrapper := wrapper
		switch {
		case desc.Extra[withAdminAuthKey]:
			endpointWrapper = wrapper.WithMiddlewares(authMiddleware.AdminAuthToken())
		case desc.Extra[withUserAuthKey]:
			endpointWrapper = wrapper.WithMiddlewares(authMiddleware.UserAuthToken())
		}
		mux.Handler(desc.HttpMethod, desc.Path, endpointWrapper.Endpoint(desc.Handler))
	}

	return mux
}

func EndpointDescriptors(r Router) []cluster.EndpointDescriptor {
	return []cluster.EndpointDescriptor{
		{
			HttpMethod: http.MethodGet,
			Path:       "/dishes",
			Handler:    r.Dish.List,
		},
		{
			HttpMethod: http.MethodPost,
			Path:       "/dishes",
			Handler:    r.Dish.AddDish,
			Extra:      map[string]any{withAdminAuthKey: true},
		},
		{
			HttpMethod: http.MethodPost,
			Path:       "/dishes/edit/:id",
			Handler:    r.Dish.EditDish,
			Extra:      map[string]any{withAdminAuthKey: true},
		},
		{
			HttpMethod: http.MethodDelete,
			Path:       "/dishes/delete/:id",
			Handler:    r.Dish.DeleteDish,
			Extra:      map[string]any{withAdminAuthKey: true},
		},
		{
			HttpMethod: http.MethodGet,
			Path:       "/dishes/all_categories",
			Handler:    r.DishCategory.GetAllCategories,
		},
		{
			HttpMethod: http.MethodGet,
			Path:       "/dishes/categories",
			Handler:    r.DishCategory.GetDishCategory,
		},
		{
			HttpMethod: http.MethodGet,
			Path:       "/dishes/categories/:id",
			Handler:    r.DishCategory.GetCategory,
		},
		{
			HttpMethod: http.MethodPost,
			Path:       "/dishes/categories",
			Handler:    r.DishCategory.AddCategory,
			Extra:      map[string]any{withAdminAuthKey: true},
		},
		{
			HttpMethod: http.MethodPost,
			Path:       "/dishes/categories/:id",
			Handler:    r.DishCategory.RenameCategory,
			Extra:      map[string]any{withAdminAuthKey: true},
		},
		{
			HttpMethod: http.MethodDelete,
			Path:       "/dishes/categories/:id",
			Handler:    r.DishCategory.DeleteCategory,
			Extra:      map[string]any{withAdminAuthKey: true},
		},
		{
			HttpMethod: http.MethodGet,
			Path:       "/restaurants",
			Handler:    r.Restaurant.GetAllRestaurants,
		},
		{
			HttpMethod: http.MethodGet,
			Path:       "/restaurants/:id",
			Handler:    r.Restaurant.GetRestaurant,
		},
		{
			HttpMethod: http.MethodPost,
			Path:       "/restaurants",
			Handler:    r.Restaurant.AddRestaurant,
			Extra:      map[string]any{withAdminAuthKey: true},
		},
		{
			HttpMethod: http.MethodPost,
			Path:       "/restaurants/:id",
			Handler:    r.Restaurant.RenameRestaurant,
			Extra:      map[string]any{withAdminAuthKey: true},
		},
		{
			HttpMethod: http.MethodDelete,
			Path:       "/restaurants/:id",
			Handler:    r.Restaurant.DeleteRestaurant,
			Extra:      map[string]any{withAdminAuthKey: true},
		},

		{
			HttpMethod: http.MethodPost,
			Path:       "/orders",
			Handler:    r.Order.ProcessOrder,
			Extra:      map[string]any{withUserAuthKey: true},
		},
		{
			HttpMethod: http.MethodGet,
			Path:       "/orders/my",
			Handler:    r.Order.GetUserOrders,
			Extra:      map[string]any{withUserAuthKey: true},
		},
		{
			HttpMethod: http.MethodPost,
			Path:       "/auth/login_by_telegram",
			Handler:    r.Auth.LoginByTelegram,
		},
		{
			HttpMethod: http.MethodGet,
			Path:       "/auth/refresh_access_token",
			Handler:    r.Auth.RefreshAccessToken,
		},
		{
			HttpMethod: http.MethodGet,
			Path:       "/auth/user_role",
			Handler:    r.Auth.GetUserRole,
		},
	}
}
