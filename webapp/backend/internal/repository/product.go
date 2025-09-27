package repository

import (
	"backend/internal/model"
	"context"
	"fmt"
	"strings"
)

type ProductRepository struct{ db DBTX }

func NewProductRepository(db DBTX) *ProductRepository { return &ProductRepository{db: db} }

// where/order ビルド（一覧と件数で共有）
func buildWhereAndOrder(req model.ListRequest) (whereSQL string, args []interface{}, orderSQL string) {
	clauses := []string{}
	args = []interface{}{}

	search := strings.TrimSpace(req.Search)
	if search != "" {
		clauses = append(clauses, "(name LIKE ? OR description LIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like)
	}

	if len(clauses) > 0 {
		whereSQL = " WHERE " + strings.Join(clauses, " AND ")
	}

	// ソートはホワイトリスト
	sortField := "product_id"
	switch req.SortField {
	case "product_id", "name", "value", "weight", "image", "description":
		sortField = req.SortField
	}
	sortOrder := "ASC"
	if strings.ToUpper(req.SortOrder) == "DESC" {
		sortOrder = "DESC"
	}

	orderSQL = fmt.Sprintf(" ORDER BY %s %s", sortField, sortOrder)
	return
}

func (r *ProductRepository) ListProducts(ctx context.Context, userID int, req model.ListRequest) ([]model.Product, int, error) {
	if req.PageSize <= 0 || req.PageSize > 200 {
		req.PageSize = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// 同じWHERE/ORDERを一覧と件数で共有
	whereSQL, args, orderSQL := buildWhereAndOrder(req)

	// 一覧は“薄い”列で返す（image/descriptionは詳細API推奨）
	listSQL := `
		SELECT product_id, name, value, weight, image, description
		FROM products` + whereSQL + orderSQL + ` LIMIT ? OFFSET ?`
	listArgs := append(append([]interface{}{}, args...), req.PageSize, req.Offset)

	var products []model.Product
	if err := r.db.SelectContext(ctx, &products, listSQL, listArgs...); err != nil {
		return nil, 0, err
	}

	// 件数：検索なしはカウンタ表、検索ありは必ず COUNT(*)
	var total int
	if strings.TrimSpace(req.Search) == "" {
		if err := r.db.GetContext(ctx, &total, `SELECT total FROM product_counters WHERE id=1`); err != nil {
			// カウンタ取得に失敗した場合は最後の手段として厳密COUNT
			if err2 := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM products`); err2 != nil {
				return nil, 0, err // もとのエラーを返す
			}
		}
	} else {
		countSQL := `SELECT COUNT(*) FROM products` + whereSQL
		if err := r.db.GetContext(ctx, &total, countSQL, args...); err != nil {
			return nil, 0, err
		}
	}

	return products, total, nil
}
