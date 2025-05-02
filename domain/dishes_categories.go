package domain

type DishCategory struct {
	Id   int32
	Name string
}

type AddCategoryRequest struct {
	Name string `validate:"required,min=1"`
}

type AddCategoryResponse struct {
	Id int32
}

type RenameCategoryRequest struct {
	Id   int32  `validate:"required" json:",omitempty"`
	Name string `validate:"required"`
}

type DeleteCategoryRequest struct {
	Id int32
}

type GetDishesCategory struct {
	Id int32
}
