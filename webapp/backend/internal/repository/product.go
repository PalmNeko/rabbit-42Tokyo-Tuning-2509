package repository

import (
	"backend/internal/model"
	"context"
)

type ProductRepository struct {
	db DBTX
}

func NewProductRepository(db DBTX) *ProductRepository {
	return &ProductRepository{db: db}
}

// 商品一覧を全件取得し、アプリケーション側でページング処理を行う
func (r *ProductRepository) ListProducts(ctx context.Context, userID int, req model.ListRequest) ([]model.Product, int, error) {
	var products []model.Product
	baseQuery := `
		SELECT product_id, name, value, weight, image, description
		FROM products
	`
	where := ""
	args := []interface{}{}
	if req.Search != "" {
		where = " WHERE MATCH(name, description) AGAINST(? IN BOOLEAN MODE)"
		args = append(args, req.Search)
	}

	if req.PageSize <= 0 || req.PageSize > 200 {
		req.PageSize = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// デフォルトのソート設定
	sortField := "product_id"
	sortOrder := "ASC"

	// 有効なソートフィールドの検証
	validSortFields := map[string]bool{
		"product_id": true,
		"name":       true,
		"value":      true,
		"weight":     true,
	}

	if req.SortField != "" && validSortFields[req.SortField] {
		sortField = req.SortField
	}

	if req.SortOrder == "DESC" {
		sortOrder = "DESC"
	}

	baseQuery += where + " ORDER BY " + sortField + " " + sortOrder + " LIMIT ? OFFSET ?"
	dataArgs := append(append([]interface{}{}, args...), req.PageSize, req.Offset)

	err := r.db.SelectContext(ctx, &products, baseQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}

	countQuery := "SELECT COUNT(*) FROM products" + where
	var total int
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
