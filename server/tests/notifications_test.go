package tests

import (
	"net/http"
	"testing"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/jariesdev/vendoreport/internal/services"
)

// ── GET /notifications ───────────────────────────────────────────────────────

func TestSearchNotifications(t *testing.T) {
	w := doRequest(http.MethodGet, "/notifications", nil, authHeader())
	assertStatus(t, w, http.StatusOK)

	var resp struct {
		Items []models.Notification `json:"items"`
		Total int64                 `json:"total"`
		Page  int                   `json:"page"`
		Size  int                   `json:"size"`
	}
	decodeJSON(t, w.Body, &resp)

	if resp.Total < 1 {
		t.Errorf("expected at least the seeded notification, got total=%d", resp.Total)
	}
	for i := 1; i < len(resp.Items); i++ {
		if resp.Items[i].CreatedAt.After(resp.Items[i-1].CreatedAt) {
			t.Errorf("notifications not ordered created_at DESC at index %d", i)
		}
	}
}

func TestSearchNotifications_UserScoping(t *testing.T) {
	otherUserID := uint(9999)
	db.Create(&models.Notification{Message: "private for someone else", UserID: &otherUserID})

	repo := repository.NewNotificationRepository(db)

	// Non-admin scope: only global + own rows.
	ownID := uint(1234)
	db.Create(&models.Notification{Message: "mine", UserID: &ownID})
	result, err := repo.Search(&ownID, 1, 100)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	for _, n := range result.Items {
		if n.UserID != nil && *n.UserID != ownID {
			t.Errorf("non-admin scope leaked notification for user %d", *n.UserID)
		}
	}

	// Admin scope (nil) sees the other user's row too.
	all, err := repo.Search(nil, 1, 100)
	if err != nil {
		t.Fatalf("admin search: %v", err)
	}
	found := false
	for _, n := range all.Items {
		if n.UserID != nil && *n.UserID == otherUserID {
			found = true
		}
	}
	if !found {
		t.Error("admin scope should include other users' notifications")
	}
}

// ── Voucher failure detection ────────────────────────────────────────────────

func testVendo(t *testing.T) *models.Vendo {
	t.Helper()
	var v models.Vendo
	if err := db.First(&v, testVendoID).Error; err != nil {
		t.Fatalf("load test vendo: %v", err)
	}
	return &v
}

func TestResolveVoucherFailures_FlagsExpiredGroupOnce(t *testing.T) {
	v := testVendo(t)
	mac := "F1:00:00:00:00:01"
	old := time.Now().Add(-30 * time.Minute)
	db.Create(&models.CoinInsert{VendoID: v.ID, MacAddress: mac, Amount: 5, InsertTime: old, Status: models.CoinInsertPending})
	db.Create(&models.CoinInsert{VendoID: v.ID, MacAddress: mac, Amount: 10, InsertTime: old.Add(30 * time.Second), Status: models.CoinInsertPending})

	if err := services.ResolveVoucherFailures(db, v); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	var failures int64
	db.Model(&models.VoucherFailure{}).Where("vendo_id = ? AND mac_address = ?", v.ID, mac).Count(&failures)
	if failures != 1 {
		t.Fatalf("expected 1 voucher failure, got %d", failures)
	}
	var failure models.VoucherFailure
	db.Where("vendo_id = ? AND mac_address = ?", v.ID, mac).First(&failure)
	if failure.CoinTotal != 15 {
		t.Errorf("expected coin total 15, got %v", failure.CoinTotal)
	}

	var notifs int64
	db.Model(&models.Notification{}).Where("message LIKE ?", "%"+mac+"%").Count(&notifs)
	if notifs < 1 {
		t.Error("expected a notification for the failure")
	}

	// Idempotency: a second resolve run creates nothing new.
	if err := services.ResolveVoucherFailures(db, v); err != nil {
		t.Fatalf("second resolve: %v", err)
	}
	var failures2, notifs2 int64
	db.Model(&models.VoucherFailure{}).Where("vendo_id = ? AND mac_address = ?", v.ID, mac).Count(&failures2)
	db.Model(&models.Notification{}).Where("message LIKE ?", "%"+mac+"%").Count(&notifs2)
	if failures2 != failures || notifs2 != notifs {
		t.Errorf("second resolve created duplicates: failures %d→%d, notifications %d→%d", failures, failures2, notifs, notifs2)
	}
}

func TestResolveVoucherFailures_PurchaseResolvesGroup(t *testing.T) {
	v := testVendo(t)
	mac := "F1:00:00:00:00:02"
	old := time.Now().Add(-30 * time.Minute)
	db.Create(&models.CoinInsert{VendoID: v.ID, MacAddress: mac, Amount: 5, InsertTime: old, Status: models.CoinInsertPending})
	db.Create(&models.CoinInsert{VendoID: v.ID, MacAddress: mac, Amount: 5, InsertTime: old.Add(20 * time.Second), Status: models.CoinInsertPending})
	db.Create(&models.VendoSale{VendoID: v.ID, SaleTime: old.Add(40 * time.Second), MacAddress: mac, Voucher: "VF-TEST", Amount: 10})

	if err := services.ResolveVoucherFailures(db, v); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	var purchased int64
	db.Model(&models.CoinInsert{}).Where("vendo_id = ? AND mac_address = ? AND status = ?", v.ID, mac, models.CoinInsertPurchased).Count(&purchased)
	if purchased != 2 {
		t.Errorf("expected both inserts purchased, got %d", purchased)
	}
	var failures int64
	db.Model(&models.VoucherFailure{}).Where("vendo_id = ? AND mac_address = ?", v.ID, mac).Count(&failures)
	if failures != 0 {
		t.Errorf("expected no failure, got %d", failures)
	}
}

func TestResolveVoucherFailures_InsideGraceWindowStaysPending(t *testing.T) {
	v := testVendo(t)
	mac := "F1:00:00:00:00:03"
	db.Create(&models.CoinInsert{VendoID: v.ID, MacAddress: mac, Amount: 5, InsertTime: time.Now().Add(-2 * time.Minute), Status: models.CoinInsertPending})

	if err := services.ResolveVoucherFailures(db, v); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	var pending int64
	db.Model(&models.CoinInsert{}).Where("vendo_id = ? AND mac_address = ? AND status = ?", v.ID, mac, models.CoinInsertPending).Count(&pending)
	if pending != 1 {
		t.Errorf("expected insert to stay pending inside grace window, got %d pending", pending)
	}
	var failures int64
	db.Model(&models.VoucherFailure{}).Where("vendo_id = ? AND mac_address = ?", v.ID, mac).Count(&failures)
	if failures != 0 {
		t.Errorf("expected no failure inside grace window, got %d", failures)
	}
}

func TestResolveVoucherFailures_SaleDriftTolerance(t *testing.T) {
	v := testVendo(t)
	mac := "F1:00:00:00:00:04"
	old := time.Now().Add(-30 * time.Minute)
	// Sale recorded 3s BEFORE the insert time (log_time drift between polls).
	db.Create(&models.CoinInsert{VendoID: v.ID, MacAddress: mac, Amount: 5, InsertTime: old, Status: models.CoinInsertPending})
	db.Create(&models.VendoSale{VendoID: v.ID, SaleTime: old.Add(-3 * time.Second), MacAddress: mac, Voucher: "VF-DRIFT", Amount: 5})

	if err := services.ResolveVoucherFailures(db, v); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	var purchased int64
	db.Model(&models.CoinInsert{}).Where("vendo_id = ? AND mac_address = ? AND status = ?", v.ID, mac, models.CoinInsertPurchased).Count(&purchased)
	if purchased != 1 {
		t.Errorf("expected insert purchased despite 3s drift, got %d", purchased)
	}
}
