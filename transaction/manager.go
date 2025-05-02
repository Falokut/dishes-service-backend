package transaction

import (
	"context"
	"dishes-service-backend/repository"
	"dishes-service-backend/service"
	"github.com/Falokut/go-kit/db"
)

type Manager struct {
	db db.Transactional
}

func NewManager(db db.Transactional) *Manager {
	return &Manager{
		db: db,
	}
}

type registerTx struct {
	repository.User
}

func (m Manager) RegisterUserTx(ctx context.Context, registerUserTx func(ctx context.Context, tx service.RegisterTx) error) error {
	return m.db.RunInTransaction(ctx,
		func(ctx context.Context, tx *db.Tx) error {
			return registerUserTx(ctx,
				registerTx{
					User: repository.NewUser(tx),
				},
			)
		},
	)
}

type dishTx struct {
	repository.Dish
}

func (m Manager) AddDishTx(ctx context.Context, addDishTx func(ctx context.Context, tx service.AddDishTx) error) error {
	return m.db.RunInTransaction(ctx,
		func(ctx context.Context, tx *db.Tx) error {
			return addDishTx(ctx,
				dishTx{
					Dish: repository.NewDish(tx),
				},
			)
		},
	)
}

func (m Manager) EditDishTx(ctx context.Context, editDishTx func(ctx context.Context, tx service.EditDishTx) error) error {
	return m.db.RunInTransaction(ctx,
		func(ctx context.Context, tx *db.Tx) error {
			return editDishTx(ctx,
				dishTx{
					Dish: repository.NewDish(tx),
				},
			)
		},
	)
}

type processOrderTx struct {
	repository.Dish
	repository.Order
}

func (m Manager) ProcessOrderTx(ctx context.Context, orderTx func(ctx context.Context, tx service.ProcessOrderTx) error) error {
	return m.db.RunInTransaction(ctx,
		func(ctx context.Context, tx *db.Tx) error {
			return orderTx(ctx,
				processOrderTx{
					Dish:  repository.NewDish(tx),
					Order: repository.NewOrder(tx),
				},
			)
		},
	)
}

type orderingAllowedTx struct {
	repository.Order
}

func (m Manager) SetOrderingAllowedTx(ctx context.Context, orderTx func(ctx context.Context, tx service.OrderingAllowTx) error) error {
	return m.db.RunInTransaction(ctx,
		func(ctx context.Context, tx *db.Tx) error {
			return orderTx(ctx,
				orderingAllowedTx{
					Order: repository.NewOrder(tx),
				},
			)
		},
	)
}
