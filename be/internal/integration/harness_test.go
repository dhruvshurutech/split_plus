package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dhruvsaxena1998/splitplus/internal/app"
	"github.com/dhruvsaxena1998/splitplus/internal/db/sqlc"
)

type testHarness struct {
	t      *testing.T
	pool   *pgxpool.Pool
	server *httptest.Server
	client *http.Client
}

func setupTestDB(t *testing.T) (*pgxpool.Pool, *sqlc.Queries) {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL_TEST")
	if dsn == "" {
		t.Skip("DATABASE_URL_TEST is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to connect to test DB: %v", err)
	}

	// Deterministic state each test run
	if err := resetSchema(ctx, pool); err != nil {
		t.Fatalf("failed to reset schema: %v", err)
	}
	if err := ensureExtensions(ctx, pool); err != nil {
		t.Fatalf("failed to ensure extensions: %v", err)
	}
	if err := applyMigrations(ctx, pool); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}

	return pool, sqlc.New(pool)
}

func setupServer(t *testing.T, pool *pgxpool.Pool, queries *sqlc.Queries) *httptest.Server {
	t.Helper()
	a := app.New(pool, queries)
	return httptest.NewServer(a.Router)
}

func newHarness(t *testing.T) *testHarness {
	t.Helper()
	pool, queries := setupTestDB(t)
	srv := setupServer(t, pool, queries)

	h := &testHarness{
		t:      t,
		pool:   pool,
		server: srv,
		client: srv.Client(),
	}

	t.Cleanup(func() {
		srv.Close()
		pool.Close()
	})

	return h
}

type standardResponse struct {
	Status bool            `json:"status"`
	Data   json.RawMessage `json:"data"`
	Error  *struct {
		Message any `json:"message"`
	} `json:"error"`
}

func (h *testHarness) doJSON(method, path string, body any, headers map[string]string) (int, standardResponse) {
	h.t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			h.t.Fatalf("encode body: %v", err)
		}
	}

	req, err := http.NewRequest(method, h.server.URL+path, &buf)
	if err != nil {
		h.t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		h.t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	var sr standardResponse
	_ = json.NewDecoder(resp.Body).Decode(&sr)
	return resp.StatusCode, sr
}

func (h *testHarness) authHeader(userID string) map[string]string {
	return map[string]string{
		"X-User-ID": userID,
	}
}
