package repository

import (
	"context"
	"database/sql"
	"os"
	"sub-service/internal/testutil"
	"testing"

	"github.com/stretchr/testify/require"
)

var testDB *sql.DB
var teardown func()

func TestMain(m *testing.M) {
	ctx := context.Background()

	tdb, err := testutil.SetupPostgres(ctx)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", tdb.URI)
	if err != nil {
		panic(err)
	}

	err = testutil.RunGooseMigrations(db, "../../internal/infrastructure/database/migrations")
	if err != nil {
		panic(err)
	}

	testDB = db
	teardown = func() {
		db.Close()
		tdb.Teardown()
	}

	code := m.Run() // RUN all the tests

	teardown()
	os.Exit(code)
}

func TruncateTables(t *testing.T, db *sql.DB) {
	_, err := db.Exec(`TRUNCATE TABLE subscriptions RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func beforeEach(t *testing.T) {
	TruncateTables(t, testDB)
}

func TestSubscription_CheckEndDateConstraint(t *testing.T) {
	beforeEach(t)

	_, err := testDB.Exec(`
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
	`,
		"Netflix",
		10,
		"018f8f6e-0000-0000-0000-000000000000",
		"2026-02-01",
		"2026-01-01", // end < start
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "check_end_date_after_start_date")
}

func TestSubscription_NoOverlappingConstraint(t *testing.T) {
	beforeEach(t)

	userID := "018f8f6e-0000-0000-0000-000000000000"

	// первая подписка
	_, err := testDB.Exec(`
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
	`,
		"Netflix",
		10,
		userID,
		"2026-01-01",
		"2026-03-01",
	)
	require.NoError(t, err)

	// перекрывающаяся подписка
	_, err = testDB.Exec(`
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
	`,
		"Netflix",
		10,
		userID,
		"2026-02-01", // overlap
		"2026-04-01",
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "no_overlapping_subscriptions")
}

func TestSubscription_NoOverlap_Success(t *testing.T) {
	beforeEach(t)

	userID := "018f8f6e-0000-0000-0000-000000000000"

	_, err := testDB.Exec(`
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
	`,
		"Netflix",
		10,
		userID,
		"2026-01-01",
		"2026-02-01",
	)
	require.NoError(t, err)

	// сразу после окончания — ок
	_, err = testDB.Exec(`
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
	`,
		"Netflix",
		10,
		userID,
		"2026-02-01",
		"2026-03-01",
	)

	require.NoError(t, err)
}

func TestSubscription_Overlap_WithInfinity(t *testing.T) {
	beforeEach(t)

	userID := "018f8f6e-0000-0000-0000-000000000000"

	// без end_date → infinity
	_, err := testDB.Exec(`
		INSERT INTO subscriptions (service_name, price, user_id, start_date)
		VALUES ($1, $2, $3, $4)
	`,
		"Netflix",
		10,
		userID,
		"2026-01-01",
	)
	require.NoError(t, err)

	// любая следующая пересечётся
	_, err = testDB.Exec(`
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
	`,
		"Netflix",
		10,
		userID,
		"2026-06-01",
		"2026-07-01",
	)

	require.Error(t, err)
}
