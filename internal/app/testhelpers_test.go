package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"car-marketplace/internal/config"
	"car-marketplace/internal/infrastructure/database/sqlite"
	"car-marketplace/migrations"
)

func testConfig(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{
		Server: config.ServerConfig{Port: "0", Mode: "test"},
		Database: config.DatabaseConfig{
			Path:            filepath.Join(t.TempDir(), "test.db"),
			MaxOpenConns:    5,
			MaxIdleConns:    5,
			ConnMaxLifetime: time.Minute,
		},
		JWT: config.JWTConfig{
			SecretKey:       strings.Repeat("s", 40),
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: time.Hour,
		},
		CORS: config.CORSConfig{AllowedOrigins: []string{"http://localhost:3000"}},
	}
}

// newTestApp builds the real application on a fresh, seeded database.
func newTestApp(t *testing.T, cfg *config.Config) *App {
	t.Helper()
	if cfg == nil {
		cfg = testConfig(t)
	}
	ctx := context.Background()
	a, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("app.New: %v", err)
	}
	t.Cleanup(func() { a.Close() })
	if err := sqlite.Seed(ctx, a.DB, migrations.Seeds); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return a
}

type result struct {
	Status int
	Header http.Header
	Raw    string
	Body   map[string]any
}

// data returns the "data" object of the response.
func (r result) data() map[string]any {
	d, _ := r.Body["data"].(map[string]any)
	return d
}

// client sends requests to the app, carrying cookies like a browser does.
type client struct {
	t       *testing.T
	app     *App
	cookies map[string]*http.Cookie
	bearer  string
	headers map[string]string
}

func newClient(t *testing.T, a *App) *client {
	return &client{t: t, app: a, cookies: map[string]*http.Cookie{}, headers: map[string]string{}}
}

func (c *client) do(method, path string, body any) result {
	c.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			c.t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	if c.bearer != "" {
		req.Header.Set("Authorization", "Bearer "+c.bearer)
	}
	for _, ck := range c.cookies {
		req.AddCookie(ck)
	}

	rec := httptest.NewRecorder()
	c.app.Handler.ServeHTTP(rec, req)

	resp := rec.Result()
	for _, ck := range resp.Cookies() {
		if ck.MaxAge < 0 || ck.Value == "" {
			delete(c.cookies, ck.Name)
		} else {
			c.cookies[ck.Name] = ck
		}
	}

	res := result{Status: rec.Code, Header: rec.Header(), Raw: rec.Body.String()}
	_ = json.Unmarshal(rec.Body.Bytes(), &res.Body)
	return res
}

var userSeq atomic.Int64

// newUser registers and logs in a fresh individual user, returning a
// cookie-authenticated client and the user's ID.
func newUser(t *testing.T, a *App) (*client, int64) {
	t.Helper()
	c := newClient(t, a)
	email := fmt.Sprintf("user%d@example.com", userSeq.Add(1))
	res := c.do("POST", "/api/v1/auth/register", map[string]any{
		"email": email, "password": "Password123", "user_type": "individual",
		"first_name": "Test", "last_name": "User",
	})
	if res.Status != http.StatusCreated {
		t.Fatalf("register: %d %s", res.Status, res.Raw)
	}
	id := int64(res.data()["id"].(float64))

	res = c.do("POST", "/api/v1/auth/login", map[string]any{"email": email, "password": "Password123"})
	if res.Status != http.StatusOK {
		t.Fatalf("login: %d %s", res.Status, res.Raw)
	}
	return c, id
}

// validListing returns a create-listing body using seeded make 1 / model 1.
func validListing() map[string]any {
	return map[string]any{
		"make_id": 1, "model_id": 1, "year": 2020, "body_type_id": 1, "fuel_type_id": 1,
		"transmission_type_id": 1, "color_id": 1, "condition_id": 1, "mileage": 1000,
		"engine_size": 2.0, "price": 15000, "currency": "USD",
		"title": "Well kept family sedan", "description": "One owner, full service history available.",
		"city": "Austin", "state": "TX", "country": "USA",
	}
}

func createListing(t *testing.T, c *client, overrides map[string]any) int64 {
	t.Helper()
	body := validListing()
	for k, v := range overrides {
		body[k] = v
	}
	res := c.do("POST", "/api/v1/cars", body)
	if res.Status != http.StatusCreated {
		t.Fatalf("create listing: %d %s", res.Status, res.Raw)
	}
	return int64(res.data()["id"].(float64))
}
