package domain

type Restaurant struct {
	Id   int32
	Name string
}

type AddRestaurantRequest struct {
	Name string `validate:"required,min=1"`
}

type AddRestaurantResponse struct {
	Id int32
}

type RenameRestaurantRequest struct {
	Id   int32  `validate:"required" json:",omitempty"`
	Name string `validate:"required"`
}

type DeleteRestaurantRequest struct {
	Id int32
}

type GetDishesRestaurant struct {
	Id int32
}
