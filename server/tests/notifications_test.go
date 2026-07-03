package tests

import (
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
	"github.com/jariesdev/vendoreport/internal/services"
)

var assignTestVendoUserOnce sync.Once

// ensureTestVendoHasAssignedUser guarantees testVendoID has at least one
// assigned user, independent of whatever other tests in the suite may have
// already assigned (or run order) — notifyVendoUsers only creates a
// notification for vendos with assigned users, so tests asserting a
// notification was created need this to hold regardless of run order.
func ensureTestVendoHasAssignedUser(t *testing.T) {
	t.Helper()
	assignTestVendoUserOnce.Do(func() {
		u := &models.User{Username: "voucher-failure-notify-user", Password: "x", IsActive: true}
		if err := db.Create(u).Error; err != nil {
			t.Fatalf("create notify-test user: %v", err)
		}
		var v models.Vendo
		if err := db.First(&v, testVendoID).Error; err != nil {
			t.Fatalf("load test vendo: %v", err)
		}
		if err := db.Model(u).Association("Vendos").Append([]models.Vendo{v}); err != nil {
			t.Fatalf("assign test vendo: %v", err)
		}
	})
}

// ── DeleteOlderThan (notifications-clear command) ───────────────────────────

func TestNotificationRepository_DeleteOlderThan(t *testing.T) {
	repo := repository.NewNotificationRepository(db)

	cutoff := time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)
	old := &models.Notification{Message: "delete-me-old", CreatedAt: cutoff.Add(-24 * time.Hour)}
	onCutoff := &models.Notification{Message: "keep-on-cutoff", CreatedAt: cutoff}
	recent := &models.Notification{Message: "keep-me-recent", CreatedAt: cutoff.Add(24 * time.Hour)}
	for _, n := range []*models.Notification{old, onCutoff, recent} {
		if err := db.Create(n).Error; err != nil {
			t.Fatalf("seed notification: %v", err)
		}
	}

	deleted, err := repo.DeleteOlderThan(cutoff)
	if err != nil {
		t.Fatalf("DeleteOlderThan: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 notification deleted, got %d", deleted)
	}

	var remainingIDs []uint
	db.Model(&models.Notification{}).Where("id IN ?", []uint{old.ID, onCutoff.ID, recent.ID}).Pluck("id", &remainingIDs)
	if len(remainingIDs) != 2 {
		t.Fatalf("expected 2 rows remaining out of the 3 seeded, got %d", len(remainingIDs))
	}
	for _, id := range remainingIDs {
		if id == old.ID {
			t.Error("notification older than the cutoff was not deleted")
		}
	}

	// Re-running with the same cutoff deletes nothing further.
	deleted2, err := repo.DeleteOlderThan(cutoff)
	if err != nil {
		t.Fatalf("second DeleteOlderThan: %v", err)
	}
	if deleted2 != 0 {
		t.Errorf("expected 0 deleted on second run, got %d", deleted2)
	}
}

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
	ownID := uint(1234)
	otherID := uint(9999)
	db.Create(&models.Notification{Message: "mine", UserID: &ownID})
	db.Create(&models.Notification{Message: "private for someone else", UserID: &otherID})
	db.Create(&models.Notification{Message: "system-wide notice"}) // global, NULL user_id

	repo := repository.NewNotificationRepository(db)

	// Non-admin (includeGlobal=false): strictly the caller's own rows —
	// never another user's, and not even global/system notifications.
	own, err := repo.Search(ownID, false, nil, 1, 100)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	for _, n := range own.Items {
		if n.UserID == nil || *n.UserID != ownID {
			t.Errorf("non-admin scope leaked a row not owned by %d: user_id=%v", ownID, n.UserID)
		}
	}
	found := false
	for _, n := range own.Items {
		if n.UserID != nil && *n.UserID == ownID {
			found = true
		}
	}
	if !found {
		t.Error("expected the caller's own notification to be present")
	}

	// Admin (includeGlobal=true): own rows plus global/system notifications,
	// but still never another specific user's notifications.
	admin, err := repo.Search(ownID, true, nil, 1, 100)
	if err != nil {
		t.Fatalf("admin search: %v", err)
	}
	sawOwn, sawGlobal, sawOther := false, false, false
	for _, n := range admin.Items {
		switch {
		case n.UserID != nil && *n.UserID == ownID:
			sawOwn = true
		case n.UserID == nil:
			sawGlobal = true
		case n.UserID != nil && *n.UserID == otherID:
			sawOther = true
		}
	}
	if !sawOwn {
		t.Error("expected admin scope to include the caller's own notification")
	}
	if !sawGlobal {
		t.Error("expected admin scope to include global/system notifications")
	}
	if sawOther {
		t.Error("admin scope must not include another specific user's notifications")
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

func TestResolveVoucherFailures_NotifiesOnlyUsersAssignedToVendo(t *testing.T) {
	// Dedicated vendo so this test's user_vendos assignment doesn't affect
	// the shared testVendoID used elsewhere in the suite.
	v := &models.Vendo{Name: "Notify Scoping Vendo", IsActive: 1}
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("create vendo: %v", err)
	}

	assignedUser := &models.User{Username: "notif-assigned-user", Password: "x", IsActive: true}
	unassignedUser := &models.User{Username: "notif-unassigned-user", Password: "x", IsActive: true}
	if err := db.Create(assignedUser).Error; err != nil {
		t.Fatalf("create assigned user: %v", err)
	}
	if err := db.Create(unassignedUser).Error; err != nil {
		t.Fatalf("create unassigned user: %v", err)
	}
	// Only assignedUser is granted access to v; unassignedUser is a user of
	// the system but has no relationship to this vendo.
	if err := db.Model(assignedUser).Association("Vendos").Replace([]models.Vendo{*v}); err != nil {
		t.Fatalf("assign vendo: %v", err)
	}

	mac := "F1:00:00:00:00:09"
	old := time.Now().Add(-30 * time.Minute)
	db.Create(&models.CoinInsert{VendoID: v.ID, MacAddress: mac, Amount: 5, InsertTime: old, Status: models.CoinInsertPending})

	if err := services.ResolveVoucherFailures(db, v); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	var notifications []models.Notification
	if err := db.Where("message LIKE ?", "%"+mac+"%").Find(&notifications).Error; err != nil {
		t.Fatalf("load notifications: %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("expected exactly 1 notification (one per assigned user), got %d", len(notifications))
	}
	if notifications[0].UserID == nil || *notifications[0].UserID != assignedUser.ID {
		t.Errorf("expected notification for assigned user %d, got user_id=%v", assignedUser.ID, notifications[0].UserID)
	}
	if notifications[0].UserID != nil && *notifications[0].UserID == unassignedUser.ID {
		t.Error("notification was sent to a user not assigned to the vendo")
	}
}

func TestResolveVoucherFailures_NoAssignedUsers_NoNotification(t *testing.T) {
	// A vendo with zero assigned users must produce zero notifications — no
	// falling back to a "global" (nil user_id) notification, since that would
	// be visible to every non-admin user (Search treats NULL as global).
	v := &models.Vendo{Name: "Unassigned Vendo", IsActive: 1}
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("create vendo: %v", err)
	}

	mac := "F1:00:00:00:00:10"
	old := time.Now().Add(-30 * time.Minute)
	db.Create(&models.CoinInsert{VendoID: v.ID, MacAddress: mac, Amount: 5, InsertTime: old, Status: models.CoinInsertPending})

	if err := services.ResolveVoucherFailures(db, v); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	var count int64
	db.Model(&models.Notification{}).Where("message LIKE ?", "%"+mac+"%").Count(&count)
	if count != 0 {
		t.Errorf("expected no notification for a vendo with no assigned users, got %d", count)
	}

	// The failure itself is still recorded for audit purposes even without a notification.
	var failures int64
	db.Model(&models.VoucherFailure{}).Where("vendo_id = ? AND mac_address = ?", v.ID, mac).Count(&failures)
	if failures != 1 {
		t.Errorf("expected the voucher failure to still be recorded, got %d", failures)
	}
}

func TestResolveVoucherFailures_FlagsExpiredGroupOnce(t *testing.T) {
	ensureTestVendoHasAssignedUser(t)
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

func TestResolveVoucherFailures_CancelledTopupReasonInMessage(t *testing.T) {
	ensureTestVendoHasAssignedUser(t)
	v := testVendo(t)
	mac := "F1:00:00:00:00:05"
	old := time.Now().Add(-30 * time.Minute)
	db.Create(&models.CoinInsert{VendoID: v.ID, MacAddress: mac, Amount: 5, InsertTime: old, Status: models.CoinInsertPending, Cancelled: true})

	if err := services.ResolveVoucherFailures(db, v); err != nil {
		t.Fatalf("resolve: %v", err)
	}

	var failure models.VoucherFailure
	if err := db.Where("vendo_id = ? AND mac_address = ?", v.ID, mac).First(&failure).Error; err != nil {
		t.Fatalf("load voucher failure: %v", err)
	}
	if !failure.Cancelled {
		t.Error("expected VoucherFailure.Cancelled to be true")
	}

	var notification models.Notification
	if err := db.Where("message LIKE ?", "%"+mac+"%").First(&notification).Error; err != nil {
		t.Fatalf("load notification: %v", err)
	}
	if !strings.Contains(notification.Message, "cancelled the top-up") {
		t.Errorf("expected notification message to mention the cancel reason, got: %q", notification.Message)
	}
}

func TestParseCancelTopupLog(t *testing.T) {
	mac, ok := services.ParseCancelTopupLog("AA:BB:CC:DD:EE:FF Cancel Topup")
	if !ok || mac != "AA:BB:CC:DD:EE:FF" {
		t.Errorf("expected mac AA:BB:CC:DD:EE:FF, ok=true, got mac=%q ok=%v", mac, ok)
	}
	if _, ok := services.ParseCancelTopupLog("AA:BB:CC:DD:EE:FF Inserted coin 5"); ok {
		t.Error("expected no match for an unrelated log line")
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
