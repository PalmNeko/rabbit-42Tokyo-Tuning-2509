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

// 商品一覧を取得（インデックス最適化済み）
func (r *ProductRepository) ListProducts(ctx context.Context, userID int, req model.ListRequest) ([]model.Product, int, error) {
	var products []model.Product

	// パラメータ検証
	if req.PageSize <= 0 || req.PageSize > 200 {
		req.PageSize = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// ソートフィールドのホワイトリスト検証
	allowedSortFields := map[string]bool{
		"product_id": true,
		"name":       true,
		"value":      true,
		"weight":     true,
	}
	if !allowedSortFields[req.SortField] {
		req.SortField = "product_id"
	}
	if req.SortOrder != "ASC" && req.SortOrder != "DESC" {
		req.SortOrder = "ASC"
	}

	var query string
	var countQuery string
	var args []interface{}

	if req.Search != "" {
		// 全文検索を使用（日本語対応）
		query = `
			SELECT product_id, name, value, weight, image, description,
			       MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE) as relevance
			FROM products
			WHERE MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE)
			ORDER BY relevance DESC, ` + req.SortField + ` ` + req.SortOrder + `, product_id ASC
			LIMIT ? OFFSET ?
		`
		countQuery = `
			SELECT COUNT(*)
			FROM products
			WHERE MATCH(name, description) AGAINST(? IN NATURAL LANGUAGE MODE)
		`
		args = []interface{}{req.Search, req.Search, req.PageSize, req.Offset}

		// 全文検索でヒットしない場合のフォールバック（LIKE検索）
		var products_fulltext []model.Product
		err := r.db.SelectContext(ctx, &products_fulltext, query, args...)
		if err != nil || len(products_fulltext) == 0 {
			// LIKE検索にフォールバック（前方一致優先）
			query = `
				(
					SELECT product_id, name, value, weight, image, description, 2 as search_priority
					FROM products
					WHERE name LIKE ?
				)
				UNION ALL
				(
					SELECT product_id, name, value, weight, image, description, 1 as search_priority
					FROM products
					WHERE name LIKE ? AND name NOT LIKE ?
				)
				UNION ALL
				(
					SELECT product_id, name, value, weight, image, description, 0 as search_priority
					FROM products
					WHERE description LIKE ? AND name NOT LIKE ?
				)
				ORDER BY search_priority DESC, ` + req.SortField + ` ` + req.SortOrder + `, product_id ASC
				LIMIT ? OFFSET ?
			`
			frontMatch := req.Search + "%"
			partialMatch := "%" + req.Search + "%"
			args = []interface{}{frontMatch, partialMatch, frontMatch, partialMatch, frontMatch, req.PageSize, req.Offset}

			countQuery = `
				SELECT COUNT(*)
				FROM products
				WHERE name LIKE ? OR description LIKE ?
			`
		} else {
			products = products_fulltext
		}
	} else {
		// 検索なしの場合
		query = `
			SELECT product_id, name, value, weight, image, description
			FROM products
			ORDER BY ` + req.SortField + ` ` + req.SortOrder + `, product_id ASC
			LIMIT ? OFFSET ?
		`
		countQuery = "SELECT COUNT(*) FROM products"
		args = []interface{}{req.PageSize, req.Offset}
	}

	// 全文検索が成功していない場合のみクエリ実行
	if len(products) == 0 {
		err := r.db.SelectContext(ctx, &products, query, args...)
		if err != nil {
			return nil, 0, err
		}
	}

	// COUNT取得（検索条件に応じて）
	var total int
	var countArgs []interface{}
	if req.Search != "" {
		if len(products) > 0 && args[0] == req.Search { // 全文検索成功
			countArgs = []interface{}{req.Search}
		} else { // LIKE検索
			frontMatch := req.Search + "%"
			partialMatch := "%" + req.Search + "%"
			countArgs = []interface{}{frontMatch, partialMatch}
		}
	}
	if err := r.db.GetContext(ctx, &total, countQuery, countArgs...); err != nil {
		return nil, 0, err
	}

	return products, total, nil
}
