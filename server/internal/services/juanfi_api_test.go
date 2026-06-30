package services

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/jariesdev/vendoreport/internal/models"
)

func TestGetRates_ParsesDeviceResponse(t *testing.T) {
	const sample = "Basic: 1 PHP / 20 Mins#1#20#131400##basic|" +
		"Standard: 5 PHP / 1 Hour 40 Mins#5#100#131400##standard|" +
		"Standard: 10 PHP / 3 Hours#10#180#131400##standard|" +
		"Advance: 20 PHP / 5 Hours#20#300#131400##advance|" +
		"Premium: 50 PHP / 1 Day#50#1440#131400##premium|" +
		"Ultimate: 100 PHP / 2 Days 6 hrs#100#3240#131400##ultimate"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/admin/api/getRates" {
			t.Errorf("unexpected request path: %s", r.URL.Path)
		}
		if r.Header.Get("X-TOKEN") != "testkey" {
			t.Errorf("expected X-TOKEN header 'testkey', got %q", r.Header.Get("X-TOKEN"))
		}
		w.Write([]byte(sample))
	}))
	defer server.Close()

	apiURL := server.URL
	apiKey := "testkey"
	vendo := &models.Vendo{APIURL: &apiURL, APIKey: &apiKey}

	rates, err := NewJuanfiAPI(vendo).GetRates()
	if err != nil {
		t.Fatalf("GetRates returned error: %v", err)
	}
	if len(rates) != 6 {
		t.Fatalf("expected 6 rates, got %d", len(rates))
	}

	first := rates[0]
	if first.Name != "Basic: 1 PHP / 20 Mins" {
		t.Errorf("unexpected name: %q", first.Name)
	}
	if first.Price != 1 {
		t.Errorf("expected price 1, got %v", first.Price)
	}
	if first.Minutes != 20 {
		t.Errorf("expected minutes 20, got %v", first.Minutes)
	}
	if first.ValidityMins != 131400 {
		t.Errorf("expected validity 131400, got %v", first.ValidityMins)
	}
	if first.DataLimitMB != nil {
		t.Errorf("expected nil data limit, got %v", *first.DataLimitMB)
	}
	if first.UserProfile != "basic" {
		t.Errorf("expected user profile 'basic', got %q", first.UserProfile)
	}

	last := rates[5]
	if last.Name != "Ultimate: 100 PHP / 2 Days 6 hrs" {
		t.Errorf("unexpected last name: %q", last.Name)
	}
	if last.Price != 100 {
		t.Errorf("expected price 100, got %v", last.Price)
	}
	if last.Minutes != 3240 {
		t.Errorf("expected minutes 3240, got %v", last.Minutes)
	}
	if last.UserProfile != "ultimate" {
		t.Errorf("expected user profile 'ultimate', got %q", last.UserProfile)
	}
}

// TestGetRates_BlankUserProfileDefaults verifies that a rate with no override
// profile falls back to "default" — Mikrotik's default hotspot user profile —
// rather than an empty string.
func TestGetRates_BlankUserProfileDefaults(t *testing.T) {
	const sample = "Basic: 1 PHP / 20 Mins#1#20#131400##"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(sample))
	}))
	defer server.Close()

	apiURL := server.URL
	apiKey := "testkey"
	vendo := &models.Vendo{APIURL: &apiURL, APIKey: &apiKey}

	rates, err := NewJuanfiAPI(vendo).GetRates()
	if err != nil {
		t.Fatalf("GetRates returned error: %v", err)
	}
	if len(rates) != 1 {
		t.Fatalf("expected 1 rate, got %d", len(rates))
	}
	if rates[0].UserProfile != "default" {
		t.Errorf("expected user profile 'default', got %q", rates[0].UserProfile)
	}
}

func TestSaveRates_PostsEncodedRates(t *testing.T) {
	const expectedData = "Basic: 1 PHP / 20 Mins#1#20#131400##basic|" +
		"Standard: 5 PHP / 1 Hour 40 Mins#5#100#131400##standard|" +
		"Standard: 10 PHP / 3 Hours#10#180#131400##standard|" +
		"Advance: 20 PHP / 5 Hours#20#300#131400##advance|" +
		"Premium: 50 PHP / 1 Day#50#1440#131400##premium|" +
		"Ultimate: 100 PHP / 2 Days 6 hrs#100#3240#131400##ultimate"

	var gotMethod, gotPath, gotContentType, gotRateType, gotToken, gotData string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		gotRateType = r.URL.Query().Get("rateType")
		gotToken = r.Header.Get("X-TOKEN")

		bodyBytes, _ := io.ReadAll(r.Body)
		form, _ := url.ParseQuery(string(bodyBytes))
		gotData = form.Get("data")

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	apiURL := server.URL
	apiKey := "testkey"
	vendo := &models.Vendo{APIURL: &apiURL, APIKey: &apiKey}

	rates := []Rate{
		{Name: "Basic: 1 PHP / 20 Mins", Price: 1, Minutes: 20, ValidityMins: 131400, UserProfile: "basic"},
		{Name: "Standard: 5 PHP / 1 Hour 40 Mins", Price: 5, Minutes: 100, ValidityMins: 131400, UserProfile: "standard"},
		{Name: "Standard: 10 PHP / 3 Hours", Price: 10, Minutes: 180, ValidityMins: 131400, UserProfile: "standard"},
		{Name: "Advance: 20 PHP / 5 Hours", Price: 20, Minutes: 300, ValidityMins: 131400, UserProfile: "advance"},
		{Name: "Premium: 50 PHP / 1 Day", Price: 50, Minutes: 1440, ValidityMins: 131400, UserProfile: "premium"},
		{Name: "Ultimate: 100 PHP / 2 Days 6 hrs", Price: 100, Minutes: 3240, ValidityMins: 131400, UserProfile: "ultimate"},
	}

	if err := NewJuanfiAPI(vendo).SaveRates(rates); err != nil {
		t.Fatalf("SaveRates returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/admin/api/saveRates" {
		t.Errorf("unexpected request path: %s", gotPath)
	}
	if gotContentType != "application/x-www-form-urlencoded; charset=UTF-8" {
		t.Errorf("unexpected Content-Type: %q", gotContentType)
	}
	if gotRateType != "1" {
		t.Errorf("expected rateType=1, got %q", gotRateType)
	}
	if gotToken != "testkey" {
		t.Errorf("expected X-TOKEN 'testkey', got %q", gotToken)
	}
	if gotData != expectedData {
		t.Errorf("encoded data mismatch:\ngot:  %q\nwant: %q", gotData, expectedData)
	}
}
