package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	myv "sub-service/internal/infrastructure/validator"
	"sub-service/internal/model"
	"sub-service/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Service interface {
	CreateSubscription(ctx context.Context, input model.CreateSubscription) (*model.Subscription, error)
	GetSubscription(ctx context.Context, id uint) (*model.Subscription, error)
	ListSubscriptions(ctx context.Context, query model.ListSubscriptionsQuery) (*model.Page[*model.Subscription], error)
	UpdateSubscription(ctx context.Context, id uint, input model.UpdateSubscription) (*model.Subscription, error)
	DeleteSubscription(ctx context.Context, id uint) error
	GetTotalCost(ctx context.Context, query model.TotalCostReq) (int, error)
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

type HandlerFunction func(w http.ResponseWriter, r *http.Request) error

func Adapt(h HandlerFunction) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			handleHTTPError(w, err)
		}
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
func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) error {
	const op = "Handler.CreateSubscription"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to create a new Subscription")

	var input model.CreateSubscriptionInput

	if err := h.validator.BindAndValidate(r, &input); err != nil {
		log.ErrorContext(r.Context(), "invalid request body", "error", err)
		return err
	}

	model, err := input.ToDomain()
	if err != nil {
		log.ErrorContext(r.Context(), "error conversion to domain", "error", err)
		return err
	}

	subscription, err := h.svc.CreateSubscription(r.Context(), *model)
	if err != nil {
		log.ErrorContext(r.Context(), "failed to create subscription", "error", err)
		return err
	}

	return writeJSON(w, http.StatusCreated, subscription)
}

// GetSubscription godoc
// @Summary Получение подписки по ID
// @Description Возвращает информацию о подписке по её ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "ID подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /subscriptions/{id} [get]
func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) error {
	const op = "Handler.GetSubscription"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to get Subscription with ID")

	id, err := h.parseUintParam(r, "id")
	if err != nil {
		log.ErrorContext(r.Context(), "invalid parameter", "message", err.Error())
		return err
	}

	subscription, err := h.svc.GetSubscription(r.Context(), id)

	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			log.WarnContext(r.Context(), "subscription not found", "id", id)
		} else {
			log.ErrorContext(r.Context(), "failed to get subscription", "error", err)
		}

		return err
	}

	return writeJSON(w, http.StatusOK, subscription)
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
func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) error {
	const op = "Handler.ListSubscriptions"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to list Subscriptions")

	query := model.ListSubscriptionsQuery{
		UserID:      nilIfEmpty(h.getQueryParam(r, "user_id")),
		ServiceName: nilIfEmpty(h.getQueryParam(r, "service_name")),
	}

	// TODO: make  size and page optional

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
		return err
	}

	subscriptions, err := h.svc.ListSubscriptions(r.Context(), query)
	if err != nil {
		log.ErrorContext(r.Context(), "failed to list subscriptions", "error", err)
		return err
	}

	return writeJSON(w, http.StatusOK, subscriptions)
}

// UpdateSubscription godoc
// @Summary Обновление подписки
// @Description Обновляет информацию о подписке по её ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "ID подписки"
// @Param subscription body model.UpdateSubscriptionInput true "Данные для обновления подписки"
// @Success 200 {object} model.Subscription
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /subscriptions/{id} [put]
func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) error {
	const op = "Handler.UpdateSubscription"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to update Subscription")

	id, err := h.parseUintParam(r, "id")
	if err != nil {
		log.ErrorContext(r.Context(), "invalid parameter", "message", err.Error())
		return err
	}

	// TODO: implement nullable fields

	var input model.UpdateSubscriptionInput
	if err := h.validator.BindAndValidate(r, &input); err != nil {
		log.ErrorContext(r.Context(), "invalid request body", "error", err)
		return err
	}

	domain, err := input.ToDomain()
	if err != nil {
		log.ErrorContext(r.Context(), "error conversion to domain", "error", err)
		return err
	}

	subscription, err := h.svc.UpdateSubscription(r.Context(), id, *domain)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			log.WarnContext(r.Context(), "subscription not found", "id", id)
		} else {
			log.ErrorContext(r.Context(), "failed to update subscription", "error", err)
		}

		return err
	}

	return writeJSON(w, http.StatusOK, subscription)
}

// DeleteSubscription godoc
// @Summary Удаление подписки
// @Description Удаляет подписку по её ID
// @Tags subscriptions
// @Param id path int true "ID подписки"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) error {
	const op = "Handler.DeleteSubscription"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to delete Subscription")

	id, err := h.parseUintParam(r, "id")
	if err != nil {
		log.ErrorContext(r.Context(), "invalid parameter", "message", err.Error())
		return err
	}

	if err := h.svc.DeleteSubscription(r.Context(), id); err != nil {

		if errors.Is(err, model.ErrNotFound) {
			log.WarnContext(r.Context(), "subscription not found", "id", id)
		} else {
			log.ErrorContext(r.Context(), "failed to delete subscription", "error", err)
		}
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
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
func (h *Handler) GetTotalCost(w http.ResponseWriter, r *http.Request) error {
	const op = "Handler.GetTotalCost"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to get total cost")

	query := model.TotalCostQuery{
		UserID:      h.getQueryParam(r, "user_id"),
		ServiceName: h.getQueryParam(r, "service_name"),
		StartDate:   h.getQueryParam(r, "start_date"),
		EndDate:     h.getQueryParam(r, "end_date"),
	}

	if err := h.validator.ValidateQuery(query); err != nil {
		log.ErrorContext(r.Context(), "invalid request query")
		return err
	}

	model, err := query.ToDomain()
	if err != nil {
		log.ErrorContext(r.Context(), "error conversion to domain", "error", err)
		return err
	}

	totalCost, err := h.svc.GetTotalCost(r.Context(), *model)
	if err != nil {
		log.ErrorContext(r.Context(), "failed to get total cost", "error", err)
		return err
	}

	return writeJSON(w, http.StatusOK, TotalCostResponse{TotalCost: totalCost})
}

// GetUUID godoc
// @Summary Генерация UUID v7 (для тестирования)
// @Description Возвращает строку UUID
// @Tags util
// @Produce text/plain
// @Success 200 {string} string "some_uuid"
// @Failure 500 {object} ErrorResponse
// @Router /uuid [get]
func (h *Handler) GetUUID(w http.ResponseWriter, r *http.Request) error {
	const op = "Handler.GetUUID"
	log := h.logger.With("op", op)
	log.DebugContext(r.Context(), "attempting to get UUID")

	generatedUUID, err := uuid.NewV7()

	if err != nil {
		log.ErrorContext(r.Context(), "failed to generate UUID")
		return err
	}

	uuidStr := generatedUUID.String()

	ctx := r.Context()

	w.Header().Set("Content-Type", "text/plain")
	// w.Write([]byte(uuidStr))

	// для тестирования работы middleware.Timeout()
	select {
	case <-time.After(1 * time.Second):
		_, err = w.Write([]byte(uuidStr))
		return err

	case <-ctx.Done():
		log.ErrorContext(ctx, ctx.Err().Error())
		return ctx.Err()
	}
}

// Вспомогательные функции

type ErrorResponse struct {
	Error  string                  `json:"error"`
	Errors []model.ValidationError `json:"errors,omitempty"`
}

type TotalCostResponse struct {
	TotalCost int `json:"total_cost"`
}

func (h *Handler) getQueryParam(r *http.Request, key string) string {
	value := r.URL.Query().Get(key)
	return strings.TrimSpace(value)
}

func (h *Handler) parseUintParam(r *http.Request, key string) (uint, error) {
	idStr := chi.URLParam(r, key)
	if idStr == "" {
		return 0, fmt.Errorf("%w: missing ID parameter", model.ErrInvalidArgument)
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid %s parameter", model.ErrInvalidArgument, key)
	}

	return uint(id), nil

}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
