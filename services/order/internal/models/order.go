package models

// Order - доменная модель заказа
// Не зависит от внешних фреймворков и протоколов
type Order struct {
	ID     string
	UserID string
	Status string // "created", "paid", "cancelled"
	Items  []OrderItem
	Total  float64
}

type OrderItem struct {
	ProductID string
	Quantity  int
	Price     float64
}

// Методы доменной логики
func (o *Order) CalculateTotal() {
	total := 0.0
	for _, item := range o.Items {
		total += item.Price * float64(item.Quantity)
	}
	o.Total = total
}

func (o *Order) CanBeCancelled() bool {
	return o.Status == "created"
}
