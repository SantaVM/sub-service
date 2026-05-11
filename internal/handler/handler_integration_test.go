package handler

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sub-service/internal/model"
	"sub-service/internal/repository"
	"sub-service/internal/service"
	"sub-service/internal/testutil"

	"github.com/stretchr/testify/require"
)

func TestCreateSubscription_Integration(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// 1. Поднимаем БД
	testDB, err := testutil.SetupPostgres(ctx)
	require.NoError(t, err)
	defer testDB.Teardown()

	db, err := sql.Open("postgres", testDB.URI)
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	}()

	// 2. Миграции
	err = testutil.RunGooseMigrations(db, "../../internal/infrastructure/database/migrations")
	require.NoError(t, err)

	// 3. Инициализация слоёв
	repo := repository.New(db, logger)
	svc := service.New(repo, logger)

	h := newTestHandler(svc)

	// 4. HTTP запрос
	body := `{
		"user_id": "018f8f6e-0000-0000-0000-000000000000",
		"service_name": "Netflix",
		"price": 10,
		"start_date": "01-2026"
	}`

	req := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// 5. Вызов handler
	h.CreateSubscription(rr, req)

	// 6. Проверка HTTP
	require.Equal(t, http.StatusCreated, rr.Code)

	// 7. Проверка БД
	list, err := repo.List(ctx, model.ListSubscriptionsQuery{
		Page: 1,
		Size: 10,
	})
	require.NoError(t, err)

	require.Len(t, list.Content, 1)
	require.Equal(t, "Netflix", list.Content[0].ServiceName)
}
