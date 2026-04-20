package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	myv "sub-service/internal/infrastructure/validator"
	"sub-service/internal/model"
	"sub-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Service interface {
	CreateSubscription(ctx context.Context, input model.CreateSubscriptionInput) (*model.Subscription, error)
	GetSubscription(ctx context.Context, id uint) (*model.Subscription, error)
	ListSubscriptions(ctx context.Context, query model.ListSubscriptionsQuery) (*model.Page[*model.Subscription], error)
	UpdateSubscription(ctx context.Context, id uint, input model.UpdateSubscriptionInput) (*model.Subscription, error)
	DeleteSubscription(ctx context.Context, id uint) error
	GetTotalCost(ctx context.Context, query model.TotalCostQuery) (int, error)
}

var _ Service = (*service.SubscriptionService)(nil)

type Handler struct {
	svc       Service
	logger    *slog.Logger
	validator *myv.Validator
}

func New(
	svc *service.SubscriptionService,
	logger *slog.Logger,
	validator *myv.Validator,
) *Handler {
	return &Handler{
		svc:       svc,
		logger:    logger,
		validator: validator,
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
	const op = "Handler.CreateSubscription"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to create a new Subscription")

	var input model.CreateSubscriptionInput
	// if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
	// 	log.ErrorContext(r.Context(), "invalid request body")
	// 	h.errorResponse(w, http.StatusBadRequest, "invalid request body")
	// 	return
	// }
	if err := h.validator.BindAndValidate(r, &input); err != nil {
		log.ErrorContext(r.Context(), "invalid request body")
		h.validationError(w, err)
		return
	}

	subscription, err := h.svc.CreateSubscription(r.Context(), input)
	if err != nil {
		log.ErrorContext(r.Context(), "failed to create subscription", "error", err)

		var status int

		switch {
		case errors.Is(err, model.ErrSubscriptionOverlap):
			status = http.StatusConflict

		case errors.Is(err, model.ErrInvalidDateRange):
			status = http.StatusBadRequest

		default:
			status = http.StatusInternalServerError
		}

		h.errorMessage(w, status, err.Error())
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
	const op = "Handler.GetSubscription"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to get Subscription with ID")

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		log.ErrorContext(r.Context(), "missing ID parameter")
		h.errorMessage(w, http.StatusBadRequest, "missing ID parameter")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		log.ErrorContext(r.Context(), "invalid subscription ID format", "id", idStr)
		h.errorMessage(w, http.StatusBadRequest, "invalid subscription ID format")
		return
	}

	subscription, err := h.svc.GetSubscription(r.Context(), uint(id))
	if err != nil {
		log.ErrorContext(r.Context(), "failed to get subscription", "error", err)
		h.errorMessage(w, http.StatusInternalServerError, "failed to get subscription")
		return
	}

	if subscription == nil {
		log.WarnContext(r.Context(), "subscription not found", "id", id)
		h.errorMessage(w, http.StatusNotFound, "subscription not found")
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
// @Param size query int true "Количество записей на странице (максимум 100)" default(10)
// @Param page query int true "Страница (начиная с 1)" default(1)
// @Success 200 {object} model.SubscriptionPageResponse
// @Failure 400 {object} ErrorResponse
// @Router /subscriptions [get]
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.ListSubscriptions"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to list Subscriptions")

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

	if err := h.validator.ValidateQuery(query); err != nil {
		log.ErrorContext(r.Context(), "invalid request query")
		h.validationError(w, err)
		return
	}

	subscriptions, err := h.svc.ListSubscriptions(r.Context(), query)
	if err != nil {
		log.ErrorContext(r.Context(), "failed to list subscriptions", "error", err)
		h.errorMessage(w, http.StatusInternalServerError, "failed to list subscriptions")
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
	const op = "Handler.UpdateSubscription"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to update Subscription")

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		log.ErrorContext(r.Context(), "missing ID parameter")
		h.errorMessage(w, http.StatusBadRequest, "missing ID parameter")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		log.ErrorContext(r.Context(), "invalid subscription ID format", "id", idStr)
		h.errorMessage(w, http.StatusBadRequest, "invalid subscription ID format")
		return
	}

	var input model.UpdateSubscriptionInput
	if err := h.validator.BindAndValidate(r, &input); err != nil {
		log.ErrorContext(r.Context(), "invalid request body", "error", err)
		h.validationError(w, err)
		return
	}

	subscription, err := h.svc.UpdateSubscription(r.Context(), uint(id), input)
	if err != nil {
		log.ErrorContext(r.Context(), "failed to update subscription", "error", err)

		var status int

		switch {
		case errors.Is(err, model.ErrSubscriptionOverlap):
			status = http.StatusConflict

		case errors.Is(err, model.ErrInvalidDateRange):
			status = http.StatusBadRequest

		default:
			status = http.StatusInternalServerError
		}

		h.errorMessage(w, status, err.Error())
		return
	}

	if subscription == nil {
		log.WarnContext(r.Context(), "subscription not found", "id", id)
		h.errorMessage(w, http.StatusNotFound, "subscription not found")
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
	const op = "Handler.DeleteSubscription"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to delete Subscription")

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		log.ErrorContext(r.Context(), "missing ID parameter")
		h.errorMessage(w, http.StatusBadRequest, "missing ID parameter")
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		log.ErrorContext(r.Context(), "invalid subscription ID format", "id", idStr)
		h.errorMessage(w, http.StatusBadRequest, "invalid subscription ID format")
		return
	}

	if err := h.svc.DeleteSubscription(r.Context(), uint(id)); err != nil {

		if strings.Contains(err.Error(), "not found") {
			log.WarnContext(r.Context(), "subscription not found", "id", id)
			h.errorMessage(w, http.StatusNotFound, err.Error())
		} else {
			log.ErrorContext(r.Context(), "failed to delete subscription", "error", err)
			h.errorMessage(w, http.StatusInternalServerError, "failed to delete subscription")
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
	const op = "Handler.GetTotalCost"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to get total cost")

	startDate := h.getQueryParam(r, "start_date")
	if startDate == nil {
		h.errorMessage(w, http.StatusBadRequest, "start_date is required")
		return
	}

	query := model.TotalCostQuery{
		UserID:      h.getQueryParam(r, "user_id"),
		ServiceName: h.getQueryParam(r, "service_name"),
		StartDate:   *startDate,
		EndDate:     h.getQueryParam(r, "end_date"),
	}

	if err := h.validator.ValidateQuery(query); err != nil {
		log.ErrorContext(r.Context(), "invalid request query")
		h.validationError(w, err)
		return
	}

	totalCost, err := h.svc.GetTotalCost(r.Context(), query)
	if err != nil {
		log.ErrorContext(r.Context(), "failed to get total cost", "error", err)
		h.errorMessage(w, http.StatusInternalServerError, "failed to calculate total cost")
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
	const op = "Handler.GetUUID"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to get UUID")

	generatedUUID, err := uuid.NewV7()

	if err != nil {
		log.ErrorContext(r.Context(), "failed to generate UUID")
		h.errorMessage(w, http.StatusInternalServerError, "failed to generate UUID")
		return
	}

	uuidStr := generatedUUID.String()

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(uuidStr))
}

// Вспомогательные функции

type ErrorResponse struct {
	Error  string                  `json:"error"`
	Errors []model.ValidationError `json:"errors,omitempty"`
}

type TotalCostResponse struct {
	TotalCost int `json:"total_cost"`
}

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	const op = "Handler.jsonResponse"
	log := h.logger.With("op", op)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error("failed to encode response", "error", err)
	}
}

func (h *Handler) errorResponse(w http.ResponseWriter, status int, resp ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) errorMessage(w http.ResponseWriter, status int, message string) {
	h.errorResponse(w, status, ErrorResponse{
		Error: message,
	})
}

func (h *Handler) validationError(w http.ResponseWriter, err error) {
	h.errorResponse(w, http.StatusBadRequest, ErrorResponse{
		Errors: h.validator.FormatErrors(err),
	})
}

func (h *Handler) getQueryParam(r *http.Request, key string) *string {
	value := r.URL.Query().Get(key)
	if value == "" {
		return nil
	}
	return &value
}
