package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service/utils"
	"context"
	"log"
)

type RobotService struct {
	store *repository.Store
}

func NewRobotService(store *repository.Store) *RobotService {
	return &RobotService{store: store}
}

func (s *RobotService) GenerateDeliveryPlan(ctx context.Context, robotID string, capacity int) (*model.DeliveryPlan, error) {
	plan := model.DeliveryPlan{
		RobotID: robotID,
		Orders:  make([]model.Order, 0),
	}

	err := utils.WithTimeout(ctx, func(ctx context.Context) error {
		return s.store.ExecTx(ctx, func(txStore *repository.Store) error {
			orders, err := txStore.OrderRepo.GetShippingOrders(ctx)
			if err != nil {
				return err
			}
			// plan, err = selectOrdersForDelivery(ctx, orders, robotID, capacity)
			sel, err := selectOrdersForDelivery(ctx, orders, robotID, capacity)
			if err != nil {
				return err
			}

			plan.RobotID = sel.RobotID
			plan.TotalWeight = sel.TotalWeight
			plan.TotalValue = sel.TotalValue
			if sel.Orders != nil {
				plan.Orders = sel.Orders
			} else {
				plan.Orders = make([]model.Order, 0)
			}

			if len(plan.Orders) > 0 {
				orderIDs := make([]int64, len(plan.Orders))
				for i, order := range plan.Orders {
					orderIDs[i] = order.OrderID
				}

				if err := txStore.OrderRepo.UpdateStatuses(ctx, orderIDs, "delivering"); err != nil {
					return err
				}
				log.Printf("Updated status to 'delivering' for %d orders", len(orderIDs))
			}
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (s *RobotService) UpdateOrderStatus(ctx context.Context, orderID int64, newStatus string) error {
	return utils.WithTimeout(ctx, func(ctx context.Context) error {
		return s.store.OrderRepo.UpdateStatuses(ctx, []int64{orderID}, newStatus)
	})
}

func selectOrdersForDelivery(ctx context.Context, orders []model.Order, robotID string, robotCapacity int) (model.DeliveryPlan, error) {
	n := len(orders)

	if n > 20 {
		return selectOrdersForDeliveryDP(ctx, orders, robotID, robotCapacity)
	} else {
		return selectOrdersForDeliveryDFS(ctx, orders, robotID, robotCapacity)
	}
}

func selectOrdersForDeliveryDFS(ctx context.Context, orders []model.Order, robotID string, robotCapacity int) (model.DeliveryPlan, error) {
	n := len(orders)
	bestValue := 0
	var bestSet []model.Order
	steps := 0
	checkEvery := 16384

	var dfs func(i, curWeight, curValue int, curSet []model.Order) bool
	dfs = func(i, curWeight, curValue int, curSet []model.Order) bool {
		if curWeight > robotCapacity {
			return false
		}
		steps++
		if checkEvery > 0 && steps%checkEvery == 0 {
			select {
			case <-ctx.Done():
				return true
			default:
			}
		}
		if i == n {
			if curValue > bestValue {
				bestValue = curValue
				bestSet = append([]model.Order{}, curSet...)
			}
			return false
		}

		if dfs(i+1, curWeight, curValue, curSet) {
			return true
		}

		order := orders[i]
		return dfs(i+1, curWeight+order.Weight, curValue+order.Value, append(curSet, order))
	}

	canceled := dfs(0, 0, 0, nil)
	if canceled {
		return model.DeliveryPlan{}, ctx.Err()
	}

	var totalWeight int
	for _, o := range bestSet {
		totalWeight += o.Weight
	}

	return model.DeliveryPlan{
		RobotID:     robotID,
		TotalWeight: totalWeight,
		TotalValue:  bestValue,
		Orders:      bestSet,
	}, nil
}

func selectOrdersForDeliveryDP(ctx context.Context, orders []model.Order, robotID string, robotCapacity int) (model.DeliveryPlan, error) {
	n := len(orders)
	if n == 0 {
		return model.DeliveryPlan{
			RobotID:     robotID,
			TotalWeight: 0,
			TotalValue:  0,
			Orders:      []model.Order{},
		}, nil
	}

	log.Printf("Using DP algorithm for %d orders with capacity %d", n, robotCapacity)

	// DPテーブル: dp[i][w] = i番目までの注文でw重量以下での最大価値
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, robotCapacity+1)
	}

	// DP計算
	for i := 1; i <= n; i++ {
		// 100回に1回コンテキストチェック
		if i%100 == 0 {
			select {
			case <-ctx.Done():
				return model.DeliveryPlan{}, ctx.Err()
			default:
			}
		}

		order := orders[i-1]
		for w := 0; w <= robotCapacity; w++ {
			// 取らない場合
			dp[i][w] = dp[i-1][w]

			// 取る場合（重量制約内なら）
			if order.Weight <= w {
				takeValue := dp[i-1][w-order.Weight] + order.Value
				if takeValue > dp[i][w] {
					dp[i][w] = takeValue
				}
			}
		}
	}

	// 解の復元
	selectedOrders := []model.Order{}
	totalWeight := 0
	i, w := n, robotCapacity

	for i > 0 && w > 0 {
		if dp[i][w] != dp[i-1][w] {
			selectedOrders = append(selectedOrders, orders[i-1])
			totalWeight += orders[i-1].Weight
			w -= orders[i-1].Weight
		}
		i--
	}

	log.Printf("DP completed: selected %d orders, total weight %d, total value %d",
		len(selectedOrders), totalWeight, dp[n][robotCapacity])

	return model.DeliveryPlan{
		RobotID:     robotID,
		TotalWeight: totalWeight,
		TotalValue:  dp[n][robotCapacity],
		Orders:      selectedOrders,
	}, nil
}
