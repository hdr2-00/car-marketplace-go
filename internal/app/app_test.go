package app

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// ---------- auth ----------

func TestProtectedRouteAcceptsAccessTokenCookie(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a) // authenticated by login cookies only

	if res := c.do("GET", "/api/v1/auth/profile", nil); res.Status != http.StatusOK {
		t.Fatalf("profile with cookie: %d %s", res.Status, res.Raw)
	}
}

func TestRefreshTokenCannotBeReused(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)
	old := *c.cookies["refresh_token"]

	if res := c.do("POST", "/api/v1/auth/refresh", nil); res.Status != http.StatusOK {
		t.Fatalf("first refresh: %d %s", res.Status, res.Raw)
	}

	replay := newClient(t, a)
	replay.cookies["refresh_token"] = &old
	if res := replay.do("POST", "/api/v1/auth/refresh", nil); res.Status != http.StatusUnauthorized {
		t.Fatalf("replayed refresh token: got %d, want 401", res.Status)
	}
}

func TestRefreshTokensAreNotStoredInPlaintext(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)

	raw, err := url.QueryUnescape(c.cookies["refresh_token"].Value) // gin escapes cookie values
	if err != nil {
		t.Fatal(err)
	}

	var n int
	if err := a.DB.QueryRow(`SELECT COUNT(*) FROM refresh_tokens WHERE token = ?`, raw).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("raw refresh token found in database")
	}
}

func TestAuthCookiesAreSecureWhenConfigured(t *testing.T) {
	cfg := testConfig(t)
	cfg.Server.CookieSecure = true
	a := newTestApp(t, cfg)
	c, _ := newUser(t, a)

	for _, name := range []string{"access_token", "refresh_token"} {
		if !c.cookies[name].Secure {
			t.Errorf("%s cookie is not Secure", name)
		}
	}
}

// ---------- errors ----------

func TestMissingListingReturns404(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)

	for _, tc := range []struct{ method, path string }{
		{"GET", "/api/v1/cars/999999"},
		{"PUT", "/api/v1/cars/999999"},
		{"DELETE", "/api/v1/cars/999999"},
		{"POST", "/api/v1/cars/999999/sold"},
		{"POST", "/api/v1/favorites/999999"},
	} {
		var body any
		if tc.method == "PUT" {
			body = map[string]any{"price": 1}
		}
		if res := c.do(tc.method, tc.path, body); res.Status != http.StatusNotFound {
			t.Errorf("%s %s: got %d, want 404 (%s)", tc.method, tc.path, res.Status, res.Raw)
		}
	}

	res := c.do("POST", "/api/v1/inquiries", map[string]any{"car_listing_id": 999999, "message": "Is this still available?"})
	if res.Status != http.StatusNotFound {
		t.Errorf("inquiry on missing listing: got %d, want 404", res.Status)
	}
}

func TestDuplicateVINIsConflictWithoutLeakingSQL(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)
	createListing(t, c, map[string]any{"vin": "1HGCM82633A004352"})

	res := c.do("POST", "/api/v1/cars", func() map[string]any {
		b := validListing()
		b["vin"] = "1HGCM82633A004352"
		return b
	}())
	if res.Status != http.StatusConflict {
		t.Fatalf("duplicate VIN: got %d, want 409", res.Status)
	}
	if strings.Contains(strings.ToLower(res.Raw), "constraint") {
		t.Fatalf("response leaks SQL error: %s", res.Raw)
	}
}

// ---------- listings ----------

func TestListingsWithoutVINCanCoexist(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)

	createListing(t, c, nil)
	createListing(t, c, nil)
}

func TestCreateListingAllowsOmittingOptionalFields(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)

	// validListing omits horsepower, number_of_doors and seating_capacity.
	createListing(t, c, nil)
}

func TestCreateListingRejectsModelOfAnotherMake(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)

	body := validListing()
	body["make_id"] = 2 // model 1 belongs to make 1
	if res := c.do("POST", "/api/v1/cars", body); res.Status != http.StatusBadRequest {
		t.Fatalf("make/model mismatch: got %d, want 400", res.Status)
	}

	id := createListing(t, c, nil)
	if res := c.do("PUT", fmt.Sprintf("/api/v1/cars/%d", id), map[string]any{"model_id": 6}); res.Status != http.StatusBadRequest {
		t.Fatalf("update to model of another make: got %d, want 400", res.Status)
	}
}

func TestCreateListingRejectsUnknownFeature(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)

	body := validListing()
	body["feature_ids"] = []int{1, 9999}
	if res := c.do("POST", "/api/v1/cars", body); res.Status != http.StatusBadRequest {
		t.Fatalf("unknown feature: got %d, want 400 (%s)", res.Status, res.Raw)
	}

	var n int
	a.DB.QueryRow(`SELECT COUNT(*) FROM car_listings WHERE title = ?`, body["title"]).Scan(&n)
	if n != 0 {
		t.Fatal("listing was created despite invalid feature")
	}
}

func TestStatusSoldOnlyViaMarkAsSold(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)
	id := createListing(t, c, nil)
	path := fmt.Sprintf("/api/v1/cars/%d", id)

	if res := c.do("PUT", path, map[string]any{"status": "sold"}); res.Status != http.StatusBadRequest {
		t.Fatalf("PUT status=sold: got %d, want 400", res.Status)
	}

	if res := c.do("POST", path+"/sold", nil); res.Status != http.StatusOK {
		t.Fatalf("mark as sold: %d %s", res.Status, res.Raw)
	}
	got := c.do("GET", path, nil).data()
	if got["status"] != "sold" || got["sold_at"] == nil {
		t.Fatalf("after mark as sold: status=%v sold_at=%v", got["status"], got["sold_at"])
	}
}

func TestListingIncludesRelatedEntities(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)
	id := createListing(t, c, map[string]any{"feature_ids": []int{1, 2}})

	got := c.do("GET", fmt.Sprintf("/api/v1/cars/%d", id), nil).data()
	for _, key := range []string{"make", "model", "body_type", "fuel_type", "transmission", "color", "condition"} {
		if m, ok := got[key].(map[string]any); !ok || m["name"] == "" {
			t.Errorf("listing.%s not populated: %v", key, got[key])
		}
	}
	if f, _ := got["features"].([]any); len(f) != 2 {
		t.Errorf("listing.features = %v, want 2 entries", got["features"])
	}

	list := c.do("GET", "/api/v1/cars?limit=1", nil)
	items, _ := list.Body["data"].([]any)
	if len(items) != 1 || items[0].(map[string]any)["make"] == nil {
		t.Errorf("list items missing make: %s", list.Raw)
	}
}

func TestViewCountIsFreshAndIgnoresOwner(t *testing.T) {
	a := newTestApp(t, nil)
	owner, _ := newUser(t, a)
	viewer, _ := newUser(t, a)
	id := createListing(t, owner, nil)
	path := fmt.Sprintf("/api/v1/cars/%d", id)

	views := func(c *client) float64 { return c.do("GET", path, nil).data()["views_count"].(float64) }

	if v := views(viewer); v != 1 {
		t.Fatalf("first view: views_count = %v, want 1", v)
	}
	if v := views(newClient(t, a)); v != 2 {
		t.Fatalf("anonymous view: views_count = %v, want 2", v)
	}
	if v := views(owner); v != 2 {
		t.Fatalf("owner view counted: views_count = %v, want 2", v)
	}
}

func TestSearchFindsWordsAndToleratesSpecialCharacters(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)
	createListing(t, c, map[string]any{"title": "Zanzibar special roadster"})

	res := c.do("GET", "/api/v1/cars?search=zanzib", nil)
	if res.Status != http.StatusOK || res.Body["pagination"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("search: %d %s", res.Status, res.Raw)
	}

	for _, q := range []string{`"`, `AND`, `foo*`, `(x`} {
		if res := c.do("GET", "/api/v1/cars?search="+q, nil); res.Status != http.StatusOK {
			t.Errorf("search %q: got %d", q, res.Status)
		}
	}
}

func TestLocationFilterIsCaseInsensitive(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)
	createListing(t, c, map[string]any{"city": "Springfield"})

	res := c.do("GET", "/api/v1/cars?city=SPRINGFIELD", nil)
	if res.Body["pagination"].(map[string]any)["total"].(float64) != 1 {
		t.Fatalf("city filter: %s", res.Raw)
	}
}

// ---------- images ----------

func TestImageLifecycle(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)
	id := createListing(t, c, nil)
	base := fmt.Sprintf("/api/v1/cars/%d/images", id)

	add := func(url string) int64 {
		res := c.do("POST", base, map[string]any{"image_url": url})
		if res.Status != http.StatusCreated {
			t.Fatalf("add image: %d %s", res.Status, res.Raw)
		}
		return int64(res.data()["id"].(float64))
	}
	first := add("https://example.com/1.jpg")
	second := add("https://example.com/2.jpg")

	if res := c.do("PUT", fmt.Sprintf("%s/%d/primary", base, second), nil); res.Status != http.StatusOK {
		t.Fatalf("set primary: %d %s", res.Status, res.Raw)
	}

	// An image from another listing cannot be made primary here.
	if res := c.do("PUT", fmt.Sprintf("%s/%d/primary", base, 1), nil); res.Status != http.StatusNotFound {
		t.Fatalf("foreign image primary: got %d, want 404", res.Status)
	}

	if res := c.do("DELETE", fmt.Sprintf("%s/%d", base, second), nil); res.Status != http.StatusOK {
		t.Fatalf("delete image: %d %s", res.Status, res.Raw)
	}

	images, _ := c.do("GET", fmt.Sprintf("/api/v1/cars/%d", id), nil).data()["images"].([]any)
	if len(images) != 1 {
		t.Fatalf("images after delete = %d, want 1", len(images))
	}
	img := images[0].(map[string]any)
	if int64(img["id"].(float64)) != first || img["is_primary"] != true {
		t.Fatalf("remaining image should be promoted to primary: %v", img)
	}

	stranger, _ := newUser(t, a)
	if res := stranger.do("DELETE", fmt.Sprintf("%s/%d", base, first), nil); res.Status != http.StatusForbidden {
		t.Fatalf("stranger delete image: got %d, want 403", res.Status)
	}
}

// ---------- inquiries ----------

func TestInquiriesAreScopedToParticipants(t *testing.T) {
	a := newTestApp(t, nil)
	seller, sellerID := newUser(t, a)
	buyer, _ := newUser(t, a)
	stranger, _ := newUser(t, a)
	id := createListing(t, seller, nil)

	if res := buyer.do("POST", "/api/v1/inquiries", map[string]any{"car_listing_id": id, "message": "Is this still available?"}); res.Status != http.StatusCreated {
		t.Fatalf("create inquiry: %d %s", res.Status, res.Raw)
	}

	total := func(res result) float64 { return res.Body["pagination"].(map[string]any)["total"].(float64) }

	if res := seller.do("GET", "/api/v1/inquiries", nil); res.Status != http.StatusOK || total(res) != 1 {
		t.Fatalf("seller should see received inquiry: %d %s", res.Status, res.Raw)
	}
	if res := buyer.do("GET", "/api/v1/inquiries", nil); total(res) != 1 {
		t.Fatalf("buyer should see sent inquiry: %s", res.Raw)
	}
	if res := stranger.do("GET", fmt.Sprintf("/api/v1/inquiries?seller_id=%d", sellerID), nil); res.Status != http.StatusForbidden {
		t.Fatalf("stranger reading seller's inquiries: got %d, want 403 (%s)", res.Status, res.Raw)
	}
}

// ---------- lists without pagination params ----------

func TestListEndpointsWorkWithoutPaginationParams(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)

	for _, path := range []string{"/api/v1/favorites", "/api/v1/inquiries", "/api/v1/reviews"} {
		if res := c.do("GET", path, nil); res.Status != http.StatusOK {
			t.Errorf("GET %s: %d %s", path, res.Status, res.Raw)
		}
	}
}

// ---------- reviews ----------

func TestReviewOfUnknownUserIs404(t *testing.T) {
	a := newTestApp(t, nil)
	c, _ := newUser(t, a)

	res := c.do("POST", "/api/v1/reviews", map[string]any{"reviewed_user_id": 999999, "rating": 5})
	if res.Status != http.StatusNotFound {
		t.Fatalf("review unknown user: got %d, want 404 (%s)", res.Status, res.Raw)
	}
}

// ---------- reference data ----------

func TestReferenceDataEndpoints(t *testing.T) {
	a := newTestApp(t, nil)
	c := newClient(t, a)

	res := c.do("GET", "/api/v1/reference/data", nil)
	if res.Status != http.StatusOK {
		t.Fatalf("reference data: %d", res.Status)
	}
	for _, key := range []string{"body_types", "fuel_types", "transmissions", "colors", "conditions", "features"} {
		if v, _ := res.data()[key].([]any); len(v) == 0 {
			t.Errorf("reference data %s empty", key)
		}
	}
	if res := c.do("GET", "/api/v1/reference/makes/1/models", nil); res.Status != http.StatusOK {
		t.Fatalf("models: %d", res.Status)
	}
}

// ---------- middleware ----------

func TestCORSPreflightHeaders(t *testing.T) {
	a := newTestApp(t, nil)
	c := newClient(t, a)
	c.headers["Origin"] = "http://localhost:3000"

	res := c.do("OPTIONS", "/api/v1/cars", nil)
	if got := res.Header.Get("Access-Control-Max-Age"); got != "43200" {
		t.Errorf("Access-Control-Max-Age = %q, want 43200", got)
	}
	if got := res.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Errorf("Access-Control-Allow-Origin = %q", got)
	}
	if !strings.Contains(res.Header.Get("Vary"), "Origin") {
		t.Errorf("Vary should include Origin, got %q", res.Header.Get("Vary"))
	}

	c.headers["Origin"] = "https://evil.example"
	if got := c.do("OPTIONS", "/api/v1/cars", nil).Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("disallowed origin got Allow-Origin %q", got)
	}
}

func TestRateLimitIgnoresSpoofedForwardedFor(t *testing.T) {
	cfg := testConfig(t)
	cfg.RateLimit.Enabled = true
	cfg.RateLimit.RequestsPerSec = 0.0001
	cfg.RateLimit.Burst = 2
	cfg.RateLimit.CleanupInterval = 0
	a := newTestApp(t, cfg)
	c := newClient(t, a)

	var last int
	for i := 0; i < 3; i++ {
		c.headers["X-Forwarded-For"] = fmt.Sprintf("203.0.113.%d", i+1)
		last = c.do("GET", "/api/v1/reference/makes", nil).Status
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("third request: got %d, want 429", last)
	}
}

func TestHealthCheckReportsDatabaseFailure(t *testing.T) {
	a := newTestApp(t, nil)
	c := newClient(t, a)

	if res := c.do("GET", "/health", nil); res.Status != http.StatusOK {
		t.Fatalf("healthy: %d", res.Status)
	}
	a.DB.Close()
	if res := c.do("GET", "/health", nil); res.Status != http.StatusServiceUnavailable {
		t.Fatalf("db down: got %d, want 503", res.Status)
	}
}
