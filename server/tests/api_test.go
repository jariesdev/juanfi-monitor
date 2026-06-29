// Package tests contains feature-level (end-to-end) tests for all HTTP endpoints.
// Each test spins up the real Gin router backed by an in-memory SQLite database
// so no external dependencies (real vendo machines, files) are required.
package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jariesdev/vendoreport/app"
	"github.com/jariesdev/vendoreport/internal/database"
	"github.com/jariesdev/vendoreport/internal/models"
	"gorm.io/gorm"
)

// ── Test harness ──────────────────────────────────────────────────────────────

var (
	router      *gin.Engine
	db          *gorm.DB
	bearerToken string
	testVendoID uint
)

// TestMain sets up a shared in-memory SQLite database and seeds it with
// baseline records before running any test. It tears down afterwards.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	var err error
	// file::memory:?cache=shared allows multiple goroutines to share one DB.
	// migrate=true: fresh in-memory SQLite needs schema created from scratch.
	db, err = database.Connect("file::memory:?cache=shared", "sqlite", true)
	if err != nil {
		panic("test database: " + err.Error())
	}

	seed(db)

	// Build the router without starting cron jobs. Use a fixed test secret.
	// isProd=false so request logs are visible during test runs.
	application := app.New(db, []string{"*"}, "test-secret-key", false, false, nil)
	router = application.Router

	// Obtain an auth token once and reuse it across all tests.
	bearerToken = mustLogin("testuser", "testpass")

	os.Exit(m.Run())
}

// seed inserts the minimum set of records required by the test suite.
func seed(db *gorm.DB) {
	// User
	db.Create(&models.User{Username: "testuser", Password: "testpass", IsActive: true})

	// Vendo
	apiURL := "http://192.168.42.10:8081"
	apiKey := "testkey"
	v := &models.Vendo{
		Name:     "Test Vendo",
		APIURL:   &apiURL,
		APIKey:   &apiKey,
		IsActive: 1,
		IsOnline: true,
	}
	db.Create(v)
	testVendoID = v.ID

	// Logs
	now := time.Now()
	db.Create(&models.VendoLog{VendoID: v.ID, LogTime: now.Add(-2 * time.Minute), Description: "System Starting..."})
	db.Create(&models.VendoLog{VendoID: v.ID, LogTime: now.Add(-1 * time.Minute), Description: "Network Connected"})

	// Sales
	db.Create(&models.VendoSale{VendoID: v.ID, SaleTime: now.Add(-3 * time.Hour), MacAddress: "AA:BB:CC:DD:EE:FF", Voucher: "VOUCHER1", Amount: 15.0})
	db.Create(&models.VendoSale{VendoID: v.ID, SaleTime: now.Add(-2 * time.Hour), MacAddress: "11:22:33:44:55:66", Voucher: "VOUCHER2", Amount: 20.0})

	// VendoStatus snapshot
	totalSales := 35.0
	currentSales := 35.0
	freeHeap := 40000
	wireless := float64(-60)
	activeUsers := 2
	customerCount := 5
	db.Create(&models.VendoStatus{
		VendoID:          v.ID,
		TotalSales:       &totalSales,
		CurrentSales:     &currentSales,
		FreeHeap:         &freeHeap,
		WirelessStrength: &wireless,
		ActiveUsers:      &activeUsers,
		CustomerCount:    &customerCount,
	})

	// Withdrawal
	amount := 35.0
	db.Create(&models.Withdrawal{VendoID: v.ID, Amount: amount})

	// Unread notification
	db.Create(&models.Notification{Message: "Test notification"})
}

// mustLogin calls POST /token and returns the bearer token. Panics on failure.
func mustLogin(username, password string) string {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)

	w := doRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"})

	if w.Code != http.StatusOK {
		panic(fmt.Sprintf("mustLogin: unexpected status %d: %s", w.Code, w.Body.String()))
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	token, ok := resp["access_token"].(string)
	if !ok || token == "" {
		panic("mustLogin: no access_token in response")
	}
	return "Bearer " + token
}

// doRequest executes an HTTP request against the test router.
func doRequest(method, path string, body *strings.Reader, headers map[string]string) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req, _ = http.NewRequest(method, path, body)
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// authHeader returns the default auth headers for protected routes.
func authHeader() map[string]string {
	return map[string]string{
		"Authorization": bearerToken,
		"Content-Type":  "application/json",
	}
}

// jsonBody serialises v to a strings.Reader for use as a request body.
func jsonBody(v interface{}) *strings.Reader {
	b, _ := json.Marshal(v)
	return strings.NewReader(string(b))
}

// decodeJSON is a test helper that asserts no JSON decode error.
func decodeJSON(t *testing.T, body *bytes.Buffer, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(body).Decode(v); err != nil {
		t.Fatalf("decodeJSON: %v (body: %s)", err, body.String())
	}
}

// assertStatus fails the test when the response code does not match expected.
func assertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Errorf("expected HTTP %d, got %d\nbody: %s", expected, w.Code, w.Body.String())
	}
}

// ── Health check ─────────────────────────────────────────────────────────────

func TestHealthCheck(t *testing.T) {
	w := doRequest(http.MethodGet, "/", nil, nil)
	assertStatus(t, w, http.StatusOK)

	var resp map[string]string
	decodeJSON(t, w.Body, &resp)
	if resp["message"] == "" {
		t.Error("expected non-empty message field")
	}
}

// ── Authentication ────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	form := url.Values{}
	form.Set("username", "testuser")
	form.Set("password", "testpass")

	w := doRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)

	if _, ok := resp["access_token"]; !ok {
		t.Error("response missing access_token")
	}
	if resp["token_type"] != "bearer" {
		t.Errorf("expected token_type 'bearer', got %v", resp["token_type"])
	}
	if _, ok := resp["user"]; !ok {
		t.Error("response missing user object")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	// Seed a dedicated user so this test doesn't depend on global seed order.
	db.Create(&models.User{Username: "wrongpwduser", Password: "correctpass", IsActive: true})

	form := url.Values{}
	form.Set("username", "wrongpwduser")
	form.Set("password", "badpass")

	w := doRequest(http.MethodPost, "/token", strings.NewReader(form.Encode()),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
	assertStatus(t, w, http.StatusBadRequest)
}

func TestLogin_MissingCredentials(t *testing.T) {
	w := doRequest(http.MethodPost, "/token", strings.NewReader(""),
		map[string]string{"Content-Type": "application/x-www-form-urlencoded"})
	assertStatus(t, w, http.StatusBadRequest)
}

func TestTokenRefresh(t *testing.T) {
	// Valid token → new access_token and expiry returned
	w := doRequest(http.MethodPost, "/token/refresh", nil, authHeader())
	assertStatus(t, w, http.StatusOK)
	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	if _, ok := resp["access_token"].(string); !ok {
		t.Error("expected access_token string in refresh response")
	}
	if _, ok := resp["expiry"]; !ok {
		t.Error("expected expiry in refresh response")
	}

	// No token → 401
	w2 := doRequest(http.MethodPost, "/token/refresh", nil, nil)
	assertStatus(t, w2, http.StatusUnauthorized)
}

// ── Auth middleware ───────────────────────────────────────────────────────────

func TestProtectedRoute_NoToken(t *testing.T) {
	w := doRequest(http.MethodGet, "/users/me", nil, nil)
	assertStatus(t, w, http.StatusUnauthorized)
}

func TestProtectedRoute_InvalidToken(t *testing.T) {
	w := doRequest(http.MethodGet, "/users/me", nil, map[string]string{
		"Authorization": "Bearer notavalidtoken",
	})
	assertStatus(t, w, http.StatusUnauthorized)
}

// ── Users ─────────────────────────────────────────────────────────────────────

func TestGetCurrentUser(t *testing.T) {
	w := doRequest(http.MethodGet, "/users/me", nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	if resp["username"] != "testuser" {
		t.Errorf("expected username 'testuser', got %v", resp["username"])
	}
	// Password must never appear in the response.
	if _, ok := resp["password"]; ok {
		t.Error("password field must not be serialised in response")
	}
}

func TestListUsers(t *testing.T) {
	w := doRequest(http.MethodGet, "/users", nil, authHeader())
	assertStatus(t, w, http.StatusOK)
}

func TestGetUserByID(t *testing.T) {
	w := doRequest(http.MethodGet, "/users/1", nil, authHeader())
	assertStatus(t, w, http.StatusOK)
}

// ── Vendo Machines ────────────────────────────────────────────────────────────

func TestListVendoMachines(t *testing.T) {
	w := doRequest(http.MethodGet, "/vendo-machines", nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	items, ok := resp["data"].([]interface{})
	if !ok || len(items) == 0 {
		t.Error("expected at least one vendo machine in data")
	}
}

func TestListVendoMachines_FilterByName(t *testing.T) {
	w := doRequest(http.MethodGet, "/vendo-machines?q=Test", nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	items := resp["data"].([]interface{})
	if len(items) == 0 {
		t.Error("filter by name 'Test' should return results")
	}
}

func TestListVendoMachines_FilterByActive(t *testing.T) {
	w := doRequest(http.MethodGet, "/vendo-machines?is_active=true", nil, authHeader())
	assertStatus(t, w, http.StatusOK)
	assertStatus(t, w, http.StatusOK)
}

func TestGetVendoMachine(t *testing.T) {
	path := fmt.Sprintf("/vendo-machines/%d", testVendoID)
	w := doRequest(http.MethodGet, path, nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object")
	}
	if data["name"] != "Test Vendo" {
		t.Errorf("expected name 'Test Vendo', got %v", data["name"])
	}
	// api_key must not appear in the response.
	if _, ok := data["api_key"]; ok {
		t.Error("api_key must not be serialised in vendo response")
	}
}

func TestGetVendoMachine_NotFound(t *testing.T) {
	// Create a vendo to confirm the DB is live, then query a definitely absent ID.
	apiURL := "http://192.168.99.99:8081"
	v := &models.Vendo{Name: "NotFound Probe", APIURL: &apiURL, IsActive: 0}
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("setup: failed to create probe vendo: %v", err)
	}

	absentID := fmt.Sprintf("%d", v.ID+999999)
	w := doRequest(http.MethodGet, "/vendo-machines/"+absentID, nil, authHeader())
	assertStatus(t, w, http.StatusNotFound)
}

func TestCreateVendoMachine(t *testing.T) {
	payload := map[string]interface{}{
		"name":    "New Vendo",
		"api_url": "http://192.168.1.100:8081",
		"api_key": "newkey",
	}
	w := doRequest(http.MethodPost, "/vendo-machines", jsonBody(payload), authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "New Vendo" {
		t.Errorf("expected name 'New Vendo', got %v", data["name"])
	}
}

func TestCreateVendoMachine_MissingName(t *testing.T) {
	payload := map[string]interface{}{
		"api_url": "http://192.168.1.100:8081",
	}
	w := doRequest(http.MethodPost, "/vendo-machines", jsonBody(payload), authHeader())
	assertStatus(t, w, http.StatusBadRequest)
}

func TestSetVendoStatus(t *testing.T) {
	path := fmt.Sprintf("/vendo-machines/%d/set-status", testVendoID)
	payload := map[string]interface{}{"status": false}
	w := doRequest(http.MethodPost, path, jsonBody(payload), authHeader())
	assertStatus(t, w, http.StatusOK)

	// Restore active status.
	payload["status"] = true
	doRequest(http.MethodPost, path, jsonBody(payload), authHeader())
}

// TestGetVendoMachineStatus calls the live Juanfi API which is unavailable in
// tests, so we expect a 502 Bad Gateway (network unreachable).
func TestGetVendoMachineStatus_NoConnection(t *testing.T) {
	path := fmt.Sprintf("/vendo-machines/%d/status", testVendoID)
	w := doRequest(http.MethodGet, path, nil, authHeader())
	assertStatus(t, w, http.StatusBadGateway)
}

// TestWithdrawCurrentSales also calls the live Juanfi API → expect 502.
func TestWithdrawCurrentSales_NoConnection(t *testing.T) {
	path := fmt.Sprintf("/vendo-machines/%d/withdraw-current-sales", testVendoID)
	w := doRequest(http.MethodPost, path, nil, authHeader())
	assertStatus(t, w, http.StatusBadGateway)
}

func TestDeleteVendoMachine(t *testing.T) {
	// Create a vendo specifically for deletion and surface any DB errors immediately.
	apiURL := "http://10.0.0.1:8081"
	v := &models.Vendo{Name: "Delete Me", APIURL: &apiURL, IsActive: 1}
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("setup: failed to create vendo for deletion: %v", err)
	}
	if v.ID == 0 {
		t.Fatal("setup: created vendo has zero ID")
	}

	path := fmt.Sprintf("/vendo-machines/%d", v.ID)
	w := doRequest(http.MethodDelete, path, nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	// Verify it's gone.
	w2 := doRequest(http.MethodGet, path, nil, authHeader())
	assertStatus(t, w2, http.StatusNotFound)
}

// ── Logs ──────────────────────────────────────────────────────────────────────

func TestSearchLogs(t *testing.T) {
	w := doRequest(http.MethodGet, "/logs", nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	if _, ok := resp["items"]; !ok {
		t.Error("expected paginated response with 'items' key")
	}
	if _, ok := resp["total"]; !ok {
		t.Error("expected 'total' key in pagination envelope")
	}
}

func TestSearchLogs_FilterByDescription(t *testing.T) {
	w := doRequest(http.MethodGet, "/logs?q=System", nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	items := resp["items"].([]interface{})
	if len(items) == 0 {
		t.Error("filter q=System should return at least one log")
	}
}

func TestSearchLogs_FilterByVendoID(t *testing.T) {
	path := fmt.Sprintf("/logs?vendo_id=%d", testVendoID)
	w := doRequest(http.MethodGet, path, nil, authHeader())
	assertStatus(t, w, http.StatusOK)
}

func TestSearchLogs_FilterByDate(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	w := doRequest(http.MethodGet, "/logs?date="+today, nil, authHeader())
	assertStatus(t, w, http.StatusOK)
}

func TestSearchLogs_Pagination(t *testing.T) {
	w := doRequest(http.MethodGet, "/logs?page=1&size=1", nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	items := resp["items"].([]interface{})
	if len(items) > 1 {
		t.Errorf("size=1 should return at most 1 item, got %d", len(items))
	}
	if resp["size"].(float64) != 1 {
		t.Errorf("expected size=1, got %v", resp["size"])
	}
}

// TestRefreshLogs triggers a log pull; the vendo API is unreachable in tests
// so the handler should still return 200 (errors are logged, not surfaced).
func TestRefreshLogs(t *testing.T) {
	w := doRequest(http.MethodPost, "/log/refresh", nil, authHeader())
	assertStatus(t, w, http.StatusOK)
}

// ── Sales ─────────────────────────────────────────────────────────────────────

func TestSearchSales(t *testing.T) {
	w := doRequest(http.MethodGet, "/sales", nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	if _, ok := resp["items"]; !ok {
		t.Error("expected paginated 'items' key")
	}
}

func TestSearchSales_FilterByMacAddress(t *testing.T) {
	w := doRequest(http.MethodGet, "/sales?q=AA:BB", nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	items := resp["items"].([]interface{})
	if len(items) == 0 {
		t.Error("filter q=AA:BB should match seeded sale")
	}
}

func TestSearchSales_FilterByVoucher(t *testing.T) {
	w := doRequest(http.MethodGet, "/sales?q=VOUCHER1", nil, authHeader())
	assertStatus(t, w, http.StatusOK)
}

func TestSearchSales_FilterByVendoID(t *testing.T) {
	path := fmt.Sprintf("/sales?vendo_id=%d", testVendoID)
	w := doRequest(http.MethodGet, path, nil, authHeader())
	assertStatus(t, w, http.StatusOK)
}

func TestSearchSales_FilterByDate(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	w := doRequest(http.MethodGet, "/sales?date="+today, nil, authHeader())
	assertStatus(t, w, http.StatusOK)
}

func TestDailySales(t *testing.T) {
	from := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")
	w := doRequest(http.MethodGet, fmt.Sprintf("/daily-sales?from_date=%s&to_date=%s", from, to), nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	if _, ok := resp["data"]; !ok {
		t.Error("expected 'data' key in daily-sales response")
	}
}

func TestDailySales_DefaultDates(t *testing.T) {
	// Without date params – should default gracefully.
	w := doRequest(http.MethodGet, "/daily-sales", nil, authHeader())
	assertStatus(t, w, http.StatusOK)
}

func TestMonthlySales(t *testing.T) {
	from := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")
	w := doRequest(http.MethodGet, fmt.Sprintf("/monthly-sales?from_date=%s&to_date=%s", from, to), nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	if _, ok := resp["data"]; !ok {
		t.Error("expected 'data' key in monthly-sales response")
	}
}

// ── Vendo Status History ──────────────────────────────────────────────────────

func TestVendoStatusHistory(t *testing.T) {
	w := doRequest(http.MethodGet, "/vendo-status-history", nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	if _, ok := resp["data"]; !ok {
		t.Error("expected 'data' key in vendo-status-history response")
	}
}

func TestVendoStatusHistory_FilterByVendoID(t *testing.T) {
	path := fmt.Sprintf("/vendo-status-history?vendo_id=%d", testVendoID)
	w := doRequest(http.MethodGet, path, nil, authHeader())
	assertStatus(t, w, http.StatusOK)
}

func TestVendoStatusHistory_ActiveOnly(t *testing.T) {
	w := doRequest(http.MethodGet, "/vendo-status-history?active_only=true", nil, authHeader())
	assertStatus(t, w, http.StatusOK)
}

func TestVendoStatusHistory_DateRange(t *testing.T) {
	from := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")
	path := fmt.Sprintf("/vendo-status-history?from_date=%s&to_date=%s", from, to)
	w := doRequest(http.MethodGet, path, nil, authHeader())
	assertStatus(t, w, http.StatusOK)
}

// ── Withdrawals ───────────────────────────────────────────────────────────────

func TestListWithdrawals(t *testing.T) {
	w := doRequest(http.MethodGet, "/withdrawals", nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp map[string]interface{}
	decodeJSON(t, w.Body, &resp)
	items, ok := resp["data"].([]interface{})
	if !ok || len(items) == 0 {
		t.Error("expected at least one withdrawal record")
	}
}
