package api

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/t4RG3T21/GoBigTech/services/notification/internal/service"
)

type AlertHandler struct {
	notificationService *service.NotificationService
	logger              *zap.Logger
}

func NewAlertHandler(notificationService *service.NotificationService, logger *zap.Logger) *AlertHandler {
	return &AlertHandler{
		notificationService: notificationService,
		logger:              logger,
	}
}

// HandleAlert обрабатывает webhook запросы от Alertmanager
func (h *AlertHandler) HandleAlert(w http.ResponseWriter, r *http.Request) {
	var alertMsg service.AlertMessage
	if err := json.NewDecoder(r.Body).Decode(&alertMsg); err != nil {
		h.logger.Error("Failed to decode alert message", zap.Error(err))
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	h.logger.Info("Received alert from Alertmanager",
		zap.Int("alerts_count", len(alertMsg.Alerts)))

	if err := h.notificationService.HandleAlert(r.Context(), alertMsg); err != nil {
		h.logger.Error("Failed to process alert", zap.Error(err))
		http.Error(w, "Failed to process alert", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
