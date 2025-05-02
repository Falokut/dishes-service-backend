package domain

type Dish struct {
	Id             int32
	Name           string
	Description    string
	Price          int32
	Url            string   `json:",omitempty"`
	Categories     []string `json:",omitempty"`
	RestaurantName string
}

type GetDishesRequest struct {
	Ids           string `query:"ids"`
	CategoriesIds string `query:"categoriesIds"`
	Limit         int32  `query:"limit" validate:"max=30"`
	Offset        int32  `query:"offset"`
}

type AddDishRequest struct {
	Name         string  `validate:"required,min=1"`
	Description  string  `validate:"max=256"`
	Price        int32   `validate:"gte=800"`
	Categories   []int32 `json:",omitempty"`
	Image        []byte  `json:",omitempty"`
	RestaurantId int32   `validate:"required"`
}

type AddDishResponse struct {
	Id int32
}

type EditDishRequest struct {
	Id           int32   `json:",omitempty" validate:"required"`
	Name         string  `validate:"required,min=1"`
	Description  string  `validate:"max=256"`
	Price        int32   `validate:"gte=800"`
	Categories   []int32 `json:",omitempty"`
	Image        []byte  `json:",omitempty"`
	RestaurantId int32   `validate:"required"`
}

type DeleteDishRequest struct {
	Id int32 `json:",omitempty" validate:"required"`
}
