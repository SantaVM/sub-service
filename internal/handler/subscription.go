package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"sub-service/internal/model"
	"sub-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	svc    *service.SubscriptionService
	logger *slog.Logger
}

func New(svc *service.SubscriptionService, logger *slog.Logger) *Handler {
	return &Handler{
		svc:    svc,
		logger: logger,
	}
}

// CreateSubscription godoc
// @Summary Создание новой подписки
// @Description Создает новую запись о подписке пользователя
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body model.CreateSubscriptionInput true "Данные для создания подписки"
// @Success 201 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /subscriptions [post]
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var input model.CreateSubscriptionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	subscription, err := h.svc.CreateSubscription(r.Context(), input)
	if err != nil {
		h.logger.Error("failed to create subscription", "error", err)

		var status int

		switch {
		case errors.Is(err, model.ErrSubscriptionOverlap):
			status = http.StatusConflict

		case errors.Is(err, model.ErrInvalidDateRange):
			status = http.StatusBadRequest

		default:
			status = http.StatusInternalServerError
		}

		h.errorResponse(w, status, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusCreated, subscription)
}

// GetSubscription godoc
// @Summary Получение подписки по ID
// @Description Возвращает информацию о подписке по её ID
// @Tags subscriptions
// @Produce json
// @Param id path string true "ID подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "missing ID parameter")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid subscription ID format")
		return
	}

	subscription, err := h.svc.GetSubscription(r.Context(), uint(id))
	if err != nil {
		h.logger.Error("failed to get subscription", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to get subscription")
		return
	}

	if subscription == nil {
		h.errorResponse(w, http.StatusNotFound, "subscription not found")
		return
	}

	h.jsonResponse(w, http.StatusOK, subscription)
}

// ListSubscriptions godoc
// @Summary Список подписок
// @Description Возвращает список подписок с фильтрацией и пагинацией
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param size query int false "Количество записей на странице" default(10)
// @Param page query int false "Страница (начиная с 1)" default(1)
// @Success 200 {object} model.SubscriptionPageResponse
// @Failure 400 {object} ErrorResponse
// @Router /subscriptions [get]
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	query := model.ListSubscriptionsQuery{
		UserID:      h.getQueryParam(r, "user_id"),
		ServiceName: h.getQueryParam(r, "service_name"),
	}

	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil {
			query.Size = size
		}
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			query.Page = page
		}
	}

	subscriptions, err := h.svc.ListSubscriptions(r.Context(), query)
	if err != nil {
		h.logger.Error("failed to list subscriptions", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to list subscriptions")
		return
	}

	h.jsonResponse(w, http.StatusOK, subscriptions)
}

// UpdateSubscription godoc
// @Summary Обновление подписки
// @Description Обновляет информацию о подписке по её ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path string true "ID подписки"
// @Param subscription body model.UpdateSubscriptionInput true "Данные для обновления подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "missing ID parameter")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid subscription ID format")
		return
	}

	var input model.UpdateSubscriptionInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	subscription, err := h.svc.UpdateSubscription(r.Context(), uint(id), input)
	if err != nil {
		h.logger.Error("failed to update subscription", "error", err)

		var status int

		switch {
		case errors.Is(err, model.ErrSubscriptionOverlap):
			status = http.StatusConflict

		case errors.Is(err, model.ErrInvalidDateRange):
			status = http.StatusBadRequest

		default:
			status = http.StatusInternalServerError
		}

		h.errorResponse(w, status, err.Error())
		return
	}

	if subscription == nil {
		h.errorResponse(w, http.StatusNotFound, "subscription not found")
		return
	}

	h.jsonResponse(w, http.StatusOK, subscription)
}

// DeleteSubscription godoc
// @Summary Удаление подписки
// @Description Удаляет подписку по её ID
// @Tags subscriptions
// @Param id path string true "ID подписки"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		h.errorResponse(w, http.StatusBadRequest, "missing ID parameter")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid subscription ID format")
		return
	}

	if err := h.svc.DeleteSubscription(r.Context(), uint(id)); err != nil {
		h.logger.Error("failed to delete subscription", "reason", err)
		if strings.Contains(err.Error(), "not found") {
			h.errorResponse(w, http.StatusNotFound, err.Error())
		} else {
			h.errorResponse(w, http.StatusInternalServerError, "failed to delete subscription")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetTotalCost godoc
// @Summary Подсчет суммарной стоимости подписок
// @Description Возвращает суммарную стоимость всех подписок за выбранный период с фильтрацией
// @Description Если end_date не задан, то считается за период в 1 календарный месяц
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "ID пользователя"
// @Param service_name query string false "Название сервиса"
// @Param start_date query string true "Дата начала периода" default(01-2026)
// @Param end_date query string false "Дата окончания периода (не включая). При отсутствии принимается период в 1 месяц" default(02-2026)
// @Success 200 {object} TotalCostResponse
// @Failure 400 {object} ErrorResponse
// @Router /subscriptions/total [get]
func (h *Handler) GetTotalCost(w http.ResponseWriter, r *http.Request) {
	query := model.TotalCostQuery{
		UserID:      h.getQueryParam(r, "user_id"),
		ServiceName: h.getQueryParam(r, "service_name"),
	}

	if query.UserID != nil {
		_, err := uuid.Parse(*query.UserID)
		if err != nil {
			h.logger.Error("parsing failed for user_id", "error", err)
			h.errorResponse(w, http.StatusBadRequest, "invalid user_id format, expected UUID")
			return
		}
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if err := query.StartDate.Parse(startDateStr); err != nil {
			h.logger.Error("parsing failed for start_date", "error", err)
			h.errorResponse(w, http.StatusBadRequest, "invalid start_date format, expected MM-YYYY")
			return
		}
	} else {
		h.logger.Error("start_date is required", "error", startDateStr)
		h.errorResponse(w, http.StatusBadRequest, "start_date is required")
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if err := query.EndDate.Parse(endDateStr); err != nil {
			h.logger.Error("parsing failed for end_date", "value", err)
			h.errorResponse(w, http.StatusBadRequest, "invalid end_date format, expected MM-YYYY")
			return
		}
	} else {
		query.EndDate.Time = query.StartDate.Time.AddDate(0, 1, 0)
	}

	// Проверка обязательных параметров и бизнес-ограничений
	if !query.StartDate.Time.Before(query.EndDate.Time) {
		h.logger.Error("start_date must be before end_date", "error", "GetTotalCost")
		h.errorResponse(w, http.StatusBadRequest, "start_date must be before end_date")
		return
	}

	// TODO: ??
	// if (query.UserID == nil || *query.UserID == "") && (query.ServiceName == nil || *query.ServiceName == "") {
	// 	h.logger.Error("user_id or service_name must not be empty", "error", "GetTotalCost")
	// 	h.errorResponse(w, http.StatusBadRequest, "user_id or service_name must not be empty")
	// 	return
	// }

	totalCost, err := h.svc.GetTotalCost(r.Context(), query)
	if err != nil {
		h.logger.Error("failed to get total cost", "error", err)
		h.errorResponse(w, http.StatusInternalServerError, "failed to calculate total cost")
		return
	}

	h.jsonResponse(w, http.StatusOK, TotalCostResponse{TotalCost: totalCost})
}

// GetUUID godoc
// @Summary Генерация UUID v7 (для тестирования)
// @Description Возвращает строку UUID
// @Tags util
// @Produce text/plain
// @Success 200 {string} string "some_uuid"
// @Failure 500 {object} ErrorResponse
// @Router /uuid [get]
func (h *Handler) GetUUID(w http.ResponseWriter, r *http.Request) {
	generatedUUID, err := uuid.NewV7()

	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "failed to generate UUID")
		return
	}

	uuidStr := generatedUUID.String()

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(uuidStr))
}

// Вспомогательные функции

type ErrorResponse struct {
	Error string `json:"error"`
}

type TotalCostResponse struct {
	TotalCost int `json:"total_cost"`
}

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) errorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

func (h *Handler) getQueryParam(r *http.Request, key string) *string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil
	}
	return &value
}
