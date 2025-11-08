package api

import (
	"encoding/json"
	"net/http"

	api "github.com/t4RG3T21/GoBigTech/services/order/api" // сгенерированный OpenAPI
	"github.com/t4RG3T21/GoBigTech/services/order/internal/models"
	"github.com/t4RG3T21/GoBigTech/services/order/internal/service"
)

// OrderHandler - адаптер между HTTP и бизнес-логикой
type OrderHandler struct {
	orderService service.OrderServiceInterface
}

func NewOrderHandler(orderService service.OrderServiceInterface) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// PostOrders - обработчик HTTP запроса
func (h *OrderHandler) PostOrders(w http.ResponseWriter, r *http.Request) {
	var req api.PostOrdersJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Валидация
	if req.UserId == "" || len(req.Items) == 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Преобразование из DTO в доменную модель
	items := make([]models.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = models.OrderItem{
			ProductID: item.ProductId,
			Quantity:  int(item.Quantity),
			Price:     100.0, // В реальности получаем из каталога
		}
	}

	// Вызов бизнес-логики
	order, err := h.orderService.CreateOrder(r.Context(), req.UserId, items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Преобразование в ответ
	resp := convertToAPIOrder(order)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetOrdersId - обработчик получения заказа по ID
func (h *OrderHandler) GetOrdersId(w http.ResponseWriter, r *http.Request, id string) {
	order, err := h.orderService.GetOrderByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := convertToAPIOrder(order)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func convertToAPIOrder(order *models.Order) api.Order {
	id := order.ID
	userID := order.UserID
	status := order.Status

	items := make([]api.OrderItem, len(order.Items))
	for i, item := range order.Items {
		quantity := int32(item.Quantity)
		items[i] = api.OrderItem{
			ProductId: item.ProductID,
			Quantity:  quantity,
		}
	}

	return api.Order{
		Id:     &id,
		UserId: &userID,
		Status: &status,
		Items:  &items,
	}
}
