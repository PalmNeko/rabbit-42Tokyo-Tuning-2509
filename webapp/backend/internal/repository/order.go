package repository

import (
	"backend/internal/model"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type OrderRepository struct {
	db DBTX
}

func NewOrderRepository(db DBTX) *OrderRepository {
	return &OrderRepository{db: db}
}

// 注文を作成し、生成された注文IDを返す
func (r *OrderRepository) Create(ctx context.Context, order *model.Order) (string, error) {
	query := `INSERT INTO orders (user_id, product_id, shipped_status, created_at) VALUES (?, ?, 'shipping', NOW())`
	result, err := r.db.ExecContext(ctx, query, order.UserID, order.ProductID)
	if err != nil {
		return "", err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", id), nil
}

// 複数の注文IDのステータスを一括で更新
// 主に配送ロボットが注文を引き受けた際に一括更新をするために使用
func (r *OrderRepository) UpdateStatuses(ctx context.Context, orderIDs []int64, newStatus string) error {
	if len(orderIDs) == 0 {
		return nil
	}
	query, args, err := sqlx.In("UPDATE orders SET shipped_status = ? WHERE order_id IN (?)", newStatus, orderIDs)
	if err != nil {
		return err
	}
	query = r.db.Rebind(query)
	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

// 配送中(shipped_status:shipping)の注文一覧を取得
func (r *OrderRepository) GetShippingOrders(ctx context.Context) ([]model.Order, error) {
	var orders []model.Order
	query := `
        SELECT
            o.order_id,
            p.weight,
            p.value
        FROM orders o
        JOIN products p ON o.product_id = p.product_id
        WHERE o.shipped_status = 'shipping'
    `
	err := r.db.SelectContext(ctx, &orders, query)
	return orders, err
}

// 注文履歴一覧を取得()
func (r *OrderRepository) ListOrders(ctx context.Context, userID int, req model.ListRequest) ([]model.Order, int, error) {
	if req.PageSize <= 0 {
		req.PageSize = 50
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	if req.Page > 0 && (req.Offset == 0 || (req.Page-1)*req.PageSize != req.Offset) {
		req.Offset = (req.Page - 1) * req.PageSize
	}

	//当てはまらないカラムを弾くためのホワイトリスト
	sortCols := map[string]string{
		"order_id":       "o.order_id",
		"product_name":   "p.name",
		"created_at":     "o.created_at",
		"shipped_status": "o.shipped_status",
		"arrived_at":     "o.arrived_at",
	}
	sortColumn, ok := sortCols[strings.ToLower(req.SortField)]
	if !ok {
		sortColumn = "o.order_id"
	}
	sortDirection := "ASC"
	if strings.ToUpper(req.SortOrder) == "DESC" {
		sortDirection = "DESC"
	}

	baseFromWhere := `
		FROM orders o
		JOIN products p ON p.product_id = o.product_id
		WHERE o.user_id = ?
	`
	queryArgs := []any{userID}

	search := strings.TrimSpace(req.Search)
	if search != "" {
		if req.Type == "prefix" {
			baseFromWhere += " AND p.name LIKE ?"
			queryArgs = append(queryArgs, search+"%")
		} else {
			baseFromWhere += " AND p.name LIKE ?"
			queryArgs = append(queryArgs, "%"+search+"%")
		}
	}

	// JOIN
	dataSQL := fmt.Sprintf(`
		SELECT
			o.order_id,
			o.product_id,
			p.name AS product_name,
			o.shipped_status,
			o.created_at,
			o.arrived_at
		%s
		ORDER BY %s %s, o.order_id ASC
		LIMIT ? OFFSET ?`, baseFromWhere, sortColumn, sortDirection)

	dataArgs := append(append([]any{}, queryArgs...), req.PageSize, req.Offset)

	type orderRow struct {
		OrderID       int64        `db:"order_id"`
		ProductID     int          `db:"product_id"`
		ProductName   string       `db:"product_name"`
		ShippedStatus string       `db:"shipped_status"`
		CreatedAt     sql.NullTime `db:"created_at"`
		ArrivedAt     sql.NullTime `db:"arrived_at"`
	}

	var ordersRaw []orderRow
	if err := r.db.SelectContext(ctx, &ordersRaw, dataSQL, dataArgs...); err != nil {
		return nil, 0, err
	}

	//ここでリターン用の代入
	orders := make([]model.Order, 0, len(ordersRaw))
	for _, row := range ordersRaw {
		orders = append(orders, model.Order{
			OrderID:       row.OrderID,
			ProductID:     row.ProductID,
			ProductName:   row.ProductName,
			ShippedStatus: row.ShippedStatus,
			CreatedAt:     row.CreatedAt.Time, // NULL の可能性があるなら model 側を sql.NullTime に
			ArrivedAt:     row.ArrivedAt,
		})
	}

	// 総件数
	countSQL := `SELECT COUNT(*) ` + baseFromWhere
	var total int
	if err := r.db.GetContext(ctx, &total, countSQL, queryArgs...); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
