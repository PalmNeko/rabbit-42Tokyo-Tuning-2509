package service

import (
	"context"
	"log"

	"backend/internal/model"
	"backend/internal/repository"
)

type ProductService struct {
	store *repository.Store
}

func NewProductService(store *repository.Store) *ProductService {
	return &ProductService{store: store}
}

func (s *ProductService) CreateOrders(ctx context.Context, userID int, items []model.RequestItem) ([]string, error) {
	var insertedOrderIDs []string

	err := s.store.ExecTx(ctx, func(txStore *repository.Store) error {
		// 数量0を除外
		itemsToProcess := make([]model.Order, 0, len(items))
		for _, item := range items {
			for i := 0; i < item.Quantity; i++ {
				itemsToProcess = append(itemsToProcess, model.Order{
					UserID:    userID,
					ProductID: item.ProductID,
				})
			}
		}
		if len(itemsToProcess) == 0 {
			return nil
		}

		// バルクインサート
		orderIDs, err := txStore.OrderRepo.CreateBulk(ctx, itemsToProcess)
		if err != nil {
			return err
		}
		insertedOrderIDs = append(insertedOrderIDs, orderIDs...)
		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Printf("Created %d orders for user %d", len(insertedOrderIDs), userID)
	return insertedOrderIDs, nil
}

func (s *ProductService) FetchProducts(ctx context.Context, userID int, req model.ListRequest) ([]model.Product, int, error) {
	products, total, err := s.store.ProductRepo.ListProducts(ctx, userID, req)
	return products, total, err
}
