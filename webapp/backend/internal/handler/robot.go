package handler

import (
	"backend/internal/model"
	"backend/internal/service"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
)

type RobotHandler struct {
	RobotSvc *service.RobotService
}

func NewRobotHandler(robotSvc *service.RobotService) *RobotHandler {
	return &RobotHandler{RobotSvc: robotSvc}
}

// 配送計画を取得
func (h *RobotHandler) GetDeliveryPlan(w http.ResponseWriter, r *http.Request) {
	robotID := "robot-001"

	capacityStr := r.URL.Query().Get("capacity")
	if capacityStr == "" {
		http.Error(w, "Query parameter 'capacity' is required", http.StatusBadRequest)
		return
	}
	capacity, err := strconv.Atoi(capacityStr)
	if err != nil {
		http.Error(w, "Query parameter 'capacity' must be an integer", http.StatusBadRequest)
		return
	}

	plan, err := h.RobotSvc.GenerateDeliveryPlan(r.Context(), robotID, capacity)
	if err != nil {
		log.Printf("Failed to generate delivery plan: %v", err)
		http.Error(w, "Failed to create delivery plan", http.StatusInternalServerError)
		return
	}

	if plan.Orders == nil {
		plan.Orders = make([]model.Order, 0)
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(plan)
}

// 配送完了時に注文ステータスを更新
func (h *RobotHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	var req model.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.RobotSvc.UpdateOrderStatus(r.Context(), req.OrderID, req.NewStatus)
	if err != nil {
		log.Printf("Failed to update order status for order %d: %v", req.OrderID, err)
		http.Error(w, "Failed to update order status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Order status updated"))
}
