package tests

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// pht is UTC+8, matching the server's aggregation zone.
var pht = time.FixedZone("PHT", 8*60*60)

// ── decode targets ────────────────────────────────────────────────────────────

type tMonthly struct {
	Month         string  `json:"month"`
	Revenue       float64 `json:"revenue"`
	OneTime       float64 `json:"one_time"`
	Recurring     float64 `json:"recurring"`
	Expenses      float64 `json:"expenses"`
	Net           float64 `json:"net"`
	CumulativeNet float64 `json:"cumulative_net"`
}

type tReport struct {
	Monthly []tMonthly `json:"monthly"`
	Yearly  []struct {
		Year          string  `json:"year"`
		Net           float64 `json:"net"`
		CumulativeNet float64 `json:"cumulative_net"`
	} `json:"yearly"`
	Summary struct {
		TotalRevenue     float64 `json:"total_revenue"`
		TotalAdjustments float64 `json:"total_adjustments"`
		TotalExpenses    float64 `json:"total_expenses"`
		Net              float64 `json:"net"`
		Status           string  `json:"status"`
	} `json:"summary"`
}

type tForecast struct {
	AvgMonthlyRevenue       float64 `json:"avg_monthly_revenue"`
	MonthlyRecurringExpense float64 `json:"monthly_recurring_expense"`
	NetMonthly              float64 `json:"net_monthly"`
	CumulativeNet           float64 `json:"cumulative_net"`
	Status                  string  `json:"status"`
	MonthsToBreakEven       *int    `json:"months_to_break_even"`
}

// ── seeding helpers ───────────────────────────────────────────────────────────

func seedRole(t *testing.T, name string, perms []string) *models.Role {
	t.Helper()
	r := &models.Role{Name: name}
	r.SetPermissions(perms)
	if err := db.Create(r).Error; err != nil {
		t.Fatalf("seed role: %v", err)
	}
	return r
}

// seedUser creates a user with the given role and assigned vendos, returning
// its ID and a bearer token.
func seedUser(t *testing.T, username string, role *models.Role, vendoIDs []uint) (uint, string) {
	t.Helper()
	hash, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	u := &models.User{Username: username, Password: string(hash), IsActive: true}
	if err := db.Create(u).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
	db.Model(u).Association("Roles").Replace([]models.Role{*role})
	if len(vendoIDs) > 0 {
		var vendos []models.Vendo
		db.Find(&vendos, vendoIDs)
		db.Model(u).Association("Vendos").Replace(vendos)
	}
	return u.ID, mustLogin(username, "pw")
}

func seedVendo(t *testing.T, name string) uint {
	t.Helper()
	v := &models.Vendo{Name: name, IsActive: 1, IsOnline: true}
	if err := db.Create(v).Error; err != nil {
		t.Fatalf("seed vendo: %v", err)
	}
	return v.ID
}

func addSale(t *testing.T, vendoID uint, when time.Time, amount float64) {
	t.Helper()
	s := &models.VendoSale{
		VendoID:    vendoID,
		SaleTime:   when,
		MacAddress: "AA:BB:CC:DD:EE:FF",
		Voucher:    fmt.Sprintf("v%d-%d", vendoID, when.UnixNano()),
		Amount:     amount,
	}
	if err := db.Create(s).Error; err != nil {
		t.Fatalf("add sale: %v", err)
	}
}

func addExpense(t *testing.T, userID uint, category string, amount float64, recurring bool, date time.Time) {
	t.Helper()
	e := &models.Expense{
		UserID:      userID,
		Category:    category,
		Amount:      amount,
		IsRecurring: recurring,
		ExpenseDate: date,
	}
	if err := db.Create(e).Error; err != nil {
		t.Fatalf("add expense: %v", err)
	}
}

func addRecurringExpense(t *testing.T, userID uint, category string, amount float64, start time.Time, end *time.Time) {
	t.Helper()
	e := &models.Expense{
		UserID:      userID,
		Category:    category,
		Amount:      amount,
		IsRecurring: true,
		ExpenseDate: start,
		EndDate:     end,
	}
	if err := db.Create(e).Error; err != nil {
		t.Fatalf("add recurring expense: %v", err)
	}
}

func addAdjustment(t *testing.T, userID uint, month string, amount float64, desc string) {
	t.Helper()
	a := &models.Adjustment{UserID: userID, Month: month, Amount: amount, Description: desc}
	if err := db.Create(a).Error; err != nil {
		t.Fatalf("add adjustment: %v", err)
	}
}

func monthlyRow(rows []tMonthly, month string) (tMonthly, bool) {
	for _, r := range rows {
		if r.Month == month {
			return r, true
		}
	}
	return tMonthly{}, false
}

// monthAnchor returns noon on the 15th of the month `offset` months from now
// (offset -1 = last month), safely inside that PHT month.
func monthAnchor(offset int) time.Time {
	now := time.Now().In(pht)
	return time.Date(now.Year(), now.Month(), 15, 12, 0, 0, 0, pht).AddDate(0, offset, 0)
}

// ── tests ─────────────────────────────────────────────────────────────────────

// Expense CRUD must be strictly scoped to the owning user.
func TestExpenseCRUD_UserScoped(t *testing.T) {
	role := seedRole(t, "ProfitRoleCRUD", []string{models.PermProfit})
	_, tokenA := seedUser(t, "owner_a", role, nil)
	_, tokenB := seedUser(t, "owner_b", role, nil)

	authA := map[string]string{"Authorization": tokenA, "Content-Type": "application/json"}
	authB := map[string]string{"Authorization": tokenB, "Content-Type": "application/json"}

	// A creates an expense.
	body := jsonBody(map[string]interface{}{
		"category": "materials", "description": "router", "amount": 1500.0,
		"is_recurring": false, "expense_date": "2026-01-10",
	})
	w := doRequest(http.MethodPost, "/expenses", body, authA)
	assertStatus(t, w, http.StatusCreated)
	var created models.Expense
	decodeJSON(t, w.Body, &created)
	if created.ID == 0 {
		t.Fatal("expected created expense to have an ID")
	}

	// A sees it; B does not.
	var listA struct {
		Data []models.Expense `json:"data"`
	}
	decodeJSON(t, doRequest(http.MethodGet, "/expenses", nil, authA).Body, &listA)
	if len(listA.Data) != 1 {
		t.Fatalf("owner A expected 1 expense, got %d", len(listA.Data))
	}
	var listB struct {
		Data []models.Expense `json:"data"`
	}
	decodeJSON(t, doRequest(http.MethodGet, "/expenses", nil, authB).Body, &listB)
	if len(listB.Data) != 0 {
		t.Fatalf("owner B should see 0 expenses, got %d", len(listB.Data))
	}

	// B cannot update or delete A's expense (scoped => 404).
	upd := jsonBody(map[string]interface{}{
		"category": "labor", "amount": 1.0, "is_recurring": false, "expense_date": "2026-01-10",
	})
	assertStatus(t, doRequest(http.MethodPut, fmt.Sprintf("/expenses/%d", created.ID), upd, authB), http.StatusNotFound)
	assertStatus(t, doRequest(http.MethodDelete, fmt.Sprintf("/expenses/%d", created.ID), nil, authB), http.StatusNotFound)

	// A can delete its own.
	assertStatus(t, doRequest(http.MethodDelete, fmt.Sprintf("/expenses/%d", created.ID), nil, authA), http.StatusOK)
}

// The profit report math: revenue from own sales, one-time expenses counted
// once, correct net/status, cumulative net matching the summary.
func TestProfitReport_Math(t *testing.T) {
	role := seedRole(t, "ProfitRoleMath", []string{models.PermProfit})
	vendoID := seedVendo(t, "MathVendo")
	userID, token := seedUser(t, "owner_math", role, []uint{vendoID})
	auth := map[string]string{"Authorization": token, "Content-Type": "application/json"}

	// Use last month so the seeded times are safely before "now" (monthAnchor's
	// day-15 could otherwise fall after now for the current month).
	m := monthAnchor(-1)
	addSale(t, vendoID, m, 100.0)
	addSale(t, vendoID, m.Add(time.Hour), 200.0) // revenue = 300
	addExpense(t, userID, "materials", 500.0, false, m)
	addExpense(t, userID, "labor", 100.0, false, m) // one-time expenses = 600

	var report tReport
	decodeJSON(t, doRequest(http.MethodGet, "/profit-report", nil, auth).Body, &report)

	if report.Summary.TotalRevenue != 300 {
		t.Errorf("total_revenue: want 300, got %v", report.Summary.TotalRevenue)
	}
	if report.Summary.TotalExpenses != 600 {
		t.Errorf("total_expenses: want 600, got %v", report.Summary.TotalExpenses)
	}
	if report.Summary.Net != -300 {
		t.Errorf("net: want -300, got %v", report.Summary.Net)
	}
	if report.Summary.Status != "loss" {
		t.Errorf("status: want loss, got %q", report.Summary.Status)
	}
	if len(report.Monthly) == 0 {
		t.Fatal("expected at least one monthly row")
	}
	last := report.Monthly[len(report.Monthly)-1]
	if last.CumulativeNet != report.Summary.Net {
		t.Errorf("final cumulative_net %v should equal summary net %v", last.CumulativeNet, report.Summary.Net)
	}
}

// Multi-tenant isolation: an admin with no assigned vendos must see 0 revenue,
// NOT every tenant's sales. The seeded testuser is admin with no vendos.
func TestProfitReport_AdminWithoutVendosSeesZeroRevenue(t *testing.T) {
	// Another owner's vendo with sales that must NOT leak into testuser's report.
	otherVendo := seedVendo(t, "OtherOwnerVendo")
	addSale(t, otherVendo, monthAnchor(0), 9999.0)

	var report tReport
	decodeJSON(t, doRequest(http.MethodGet, "/profit-report", nil, authHeader()).Body, &report)
	if report.Summary.TotalRevenue != 0 {
		t.Errorf("admin without assigned vendos should see 0 revenue, got %v", report.Summary.TotalRevenue)
	}
}

// Forecast: positive net monthly with unrecovered capital => on_track with a
// finite months_to_break_even.
func TestProfitForecast_OnTrack(t *testing.T) {
	role := seedRole(t, "ProfitRoleOnTrack", []string{models.PermProfit})
	vendoID := seedVendo(t, "OnTrackVendo")
	userID, token := seedUser(t, "owner_ontrack", role, []uint{vendoID})
	auth := map[string]string{"Authorization": token, "Content-Type": "application/json"}

	// 1000 revenue in each of the last 3 complete months.
	for _, off := range []int{-1, -2, -3} {
		addSale(t, vendoID, monthAnchor(off), 1000.0)
	}
	addExpense(t, userID, "electricity", 200.0, true, monthAnchor(-3)) // recurring drain
	addExpense(t, userID, "materials", 50000.0, false, monthAnchor(-3)) // large sunk capital

	var f tForecast
	decodeJSON(t, doRequest(http.MethodGet, "/profit-forecast", nil, auth).Body, &f)

	if f.MonthlyRecurringExpense != 200 {
		t.Errorf("monthly_recurring_expense: want 200, got %v", f.MonthlyRecurringExpense)
	}
	if f.AvgMonthlyRevenue != 1000 {
		t.Errorf("avg_monthly_revenue: want 1000, got %v", f.AvgMonthlyRevenue)
	}
	if f.NetMonthly != 800 {
		t.Errorf("net_monthly: want 800, got %v", f.NetMonthly)
	}
	if f.Status != "on_track" {
		t.Fatalf("status: want on_track, got %q (cumulative %v)", f.Status, f.CumulativeNet)
	}
	if f.MonthsToBreakEven == nil || *f.MonthsToBreakEven <= 0 {
		t.Errorf("expected positive months_to_break_even, got %v", f.MonthsToBreakEven)
	}
}

// Forecast: recurring costs exceed revenue => not_recovering, no break-even month.
func TestProfitForecast_NotRecovering(t *testing.T) {
	role := seedRole(t, "ProfitRoleLosing", []string{models.PermProfit})
	vendoID := seedVendo(t, "LosingVendo")
	userID, token := seedUser(t, "owner_losing", role, []uint{vendoID})
	auth := map[string]string{"Authorization": token, "Content-Type": "application/json"}

	for _, off := range []int{-1, -2, -3} {
		addSale(t, vendoID, monthAnchor(off), 100.0) // avg 100/mo
	}
	addExpense(t, userID, "subscription", 500.0, true, monthAnchor(-3)) // 500/mo drain > revenue

	var f tForecast
	decodeJSON(t, doRequest(http.MethodGet, "/profit-forecast", nil, auth).Body, &f)

	if f.NetMonthly >= 0 {
		t.Errorf("expected negative net_monthly, got %v", f.NetMonthly)
	}
	if f.Status != "not_recovering" {
		t.Errorf("status: want not_recovering, got %q", f.Status)
	}
	if f.MonthsToBreakEven != nil {
		t.Errorf("expected no months_to_break_even, got %v", *f.MonthsToBreakEven)
	}
}

// Forecast: revenue already exceeds all costs => recovered.
func TestProfitForecast_Recovered(t *testing.T) {
	role := seedRole(t, "ProfitRoleRecovered", []string{models.PermProfit})
	vendoID := seedVendo(t, "RecoveredVendo")
	userID, token := seedUser(t, "owner_recovered", role, []uint{vendoID})
	auth := map[string]string{"Authorization": token, "Content-Type": "application/json"}

	addSale(t, vendoID, monthAnchor(-1), 10000.0)
	addExpense(t, userID, "materials", 100.0, false, monthAnchor(-1)) // tiny cost

	var f tForecast
	decodeJSON(t, doRequest(http.MethodGet, "/profit-forecast", nil, auth).Body, &f)

	if f.CumulativeNet <= 0 {
		t.Errorf("expected positive cumulative_net, got %v", f.CumulativeNet)
	}
	if f.Status != "recovered" {
		t.Errorf("status: want recovered, got %q", f.Status)
	}
}

// A recurring expense with an end date only contributes within its [start, end]
// span — months after the end date carry no recurring cost.
func TestProfitReport_RecurringWithEndDate(t *testing.T) {
	role := seedRole(t, "ProfitRoleEndDate", []string{models.PermProfit})
	vendoID := seedVendo(t, "EndDateVendo")
	userID, token := seedUser(t, "owner_enddate", role, []uint{vendoID})
	auth := map[string]string{"Authorization": token, "Content-Type": "application/json"}

	start := monthAnchor(-3)
	end := monthAnchor(-2)
	addRecurringExpense(t, userID, "electricity", 300.0, start, &end) // active months -3 and -2 only

	var report tReport
	decodeJSON(t, doRequest(http.MethodGet, "/profit-report", nil, auth).Body, &report)

	// 300 for each of the two active months.
	if report.Summary.TotalExpenses != 600 {
		t.Errorf("total_expenses: want 600 (300 x 2 months), got %v", report.Summary.TotalExpenses)
	}
	if r, ok := monthlyRow(report.Monthly, end.Format("2006-01")); ok {
		if r.Recurring != 300 {
			t.Errorf("end month recurring: want 300, got %v", r.Recurring)
		}
	} else {
		t.Errorf("missing monthly row for %s", end.Format("2006-01"))
	}
	if r, ok := monthlyRow(report.Monthly, monthAnchor(-1).Format("2006-01")); ok {
		if r.Recurring != 0 {
			t.Errorf("month after end recurring: want 0, got %v", r.Recurring)
		}
	}
}

// A recurring expense that ended before the current month must not count toward
// the forecast's forward monthly recurring drain.
func TestProfitForecast_EndedRecurringExcluded(t *testing.T) {
	role := seedRole(t, "ProfitRoleEnded", []string{models.PermProfit})
	vendoID := seedVendo(t, "EndedVendo")
	userID, token := seedUser(t, "owner_ended", role, []uint{vendoID})
	auth := map[string]string{"Authorization": token, "Content-Type": "application/json"}

	addSale(t, vendoID, monthAnchor(-1), 1000.0)
	ended := monthAnchor(-1)
	addRecurringExpense(t, userID, "subscription", 500.0, monthAnchor(-3), &ended) // ended last month

	var f tForecast
	decodeJSON(t, doRequest(http.MethodGet, "/profit-forecast", nil, auth).Body, &f)

	if f.MonthlyRecurringExpense != 0 {
		t.Errorf("ended recurring should be excluded from forward drain, got %v", f.MonthlyRecurringExpense)
	}
}

// end_date before expense_date is rejected.
func TestExpense_EndDateBeforeStart_Rejected(t *testing.T) {
	role := seedRole(t, "ProfitRoleBadEnd", []string{models.PermProfit})
	_, token := seedUser(t, "owner_badend", role, nil)
	auth := map[string]string{"Authorization": token, "Content-Type": "application/json"}

	body := jsonBody(map[string]interface{}{
		"category": "subscription", "amount": 100.0, "is_recurring": true,
		"expense_date": "2026-03-01", "end_date": "2026-01-01",
	})
	assertStatus(t, doRequest(http.MethodPost, "/expenses", body, auth), http.StatusBadRequest)
}

// Adjustment CRUD must be strictly scoped to the owning user.
func TestAdjustmentCRUD_UserScoped(t *testing.T) {
	role := seedRole(t, "ProfitRoleAdjCRUD", []string{models.PermProfit})
	_, tokenA := seedUser(t, "adj_owner_a", role, nil)
	_, tokenB := seedUser(t, "adj_owner_b", role, nil)
	authA := map[string]string{"Authorization": tokenA, "Content-Type": "application/json"}
	authB := map[string]string{"Authorization": tokenB, "Content-Type": "application/json"}

	body := jsonBody(map[string]interface{}{
		"month": "2026-02", "amount": -25.0, "description": "device over-counted",
	})
	w := doRequest(http.MethodPost, "/adjustments", body, authA)
	assertStatus(t, w, http.StatusCreated)
	var created models.Adjustment
	decodeJSON(t, w.Body, &created)
	if created.ID == 0 || created.Amount != -25.0 {
		t.Fatalf("unexpected created adjustment: %+v", created)
	}

	var listB struct {
		Data []models.Adjustment `json:"data"`
	}
	decodeJSON(t, doRequest(http.MethodGet, "/adjustments", nil, authB).Body, &listB)
	if len(listB.Data) != 0 {
		t.Fatalf("owner B should see 0 adjustments, got %d", len(listB.Data))
	}

	// B cannot modify A's adjustment.
	upd := jsonBody(map[string]interface{}{"month": "2026-02", "amount": 5.0})
	assertStatus(t, doRequest(http.MethodPut, fmt.Sprintf("/adjustments/%d", created.ID), upd, authB), http.StatusNotFound)
	assertStatus(t, doRequest(http.MethodDelete, fmt.Sprintf("/adjustments/%d", created.ID), nil, authB), http.StatusNotFound)
	// A can delete its own.
	assertStatus(t, doRequest(http.MethodDelete, fmt.Sprintf("/adjustments/%d", created.ID), nil, authA), http.StatusOK)
}

// A zero-amount adjustment is rejected.
func TestAdjustment_ZeroAmount_Rejected(t *testing.T) {
	role := seedRole(t, "ProfitRoleAdjZero", []string{models.PermProfit})
	_, token := seedUser(t, "adj_zero", role, nil)
	auth := map[string]string{"Authorization": token, "Content-Type": "application/json"}
	body := jsonBody(map[string]interface{}{"month": "2026-02", "amount": 0})
	assertStatus(t, doRequest(http.MethodPost, "/adjustments", body, auth), http.StatusBadRequest)
}

// Adjustments fold into monthly revenue and the summary net.
func TestProfitReport_IncludesAdjustments(t *testing.T) {
	role := seedRole(t, "ProfitRoleAdj", []string{models.PermProfit})
	vendoID := seedVendo(t, "AdjVendo")
	userID, token := seedUser(t, "owner_adj", role, []uint{vendoID})
	auth := map[string]string{"Authorization": token, "Content-Type": "application/json"}

	m := monthAnchor(-1)
	key := m.Format("2006-01")
	addSale(t, vendoID, m, 100.0)
	addAdjustment(t, userID, key, 50.0, "under-counted")  // +50
	addAdjustment(t, userID, key, -20.0, "over-counted")  // -20  => net +30

	var report tReport
	decodeJSON(t, doRequest(http.MethodGet, "/profit-report", nil, auth).Body, &report)

	if report.Summary.TotalRevenue != 100 {
		t.Errorf("total_revenue: want 100, got %v", report.Summary.TotalRevenue)
	}
	if report.Summary.TotalAdjustments != 30 {
		t.Errorf("total_adjustments: want 30, got %v", report.Summary.TotalAdjustments)
	}
	if report.Summary.Net != 130 {
		t.Errorf("net: want 130 (100 + 30), got %v", report.Summary.Net)
	}
	if r, ok := monthlyRow(report.Monthly, key); ok {
		if r.Net != 130 {
			t.Errorf("month net: want 130, got %v", r.Net)
		}
	} else {
		t.Errorf("missing monthly row for %s", key)
	}
}

// Profit endpoints require the profit permission.
func TestProfit_ACL_Forbidden(t *testing.T) {
	role := seedRole(t, "NoProfitRole", []string{models.PermDashboard})
	_, token := seedUser(t, "noprofit_user", role, nil)
	auth := map[string]string{"Authorization": token, "Content-Type": "application/json"}

	assertStatus(t, doRequest(http.MethodGet, "/expenses", nil, auth), http.StatusForbidden)
	assertStatus(t, doRequest(http.MethodGet, "/profit-report", nil, auth), http.StatusForbidden)
	assertStatus(t, doRequest(http.MethodGet, "/profit-forecast", nil, auth), http.StatusForbidden)
	assertStatus(t, doRequest(http.MethodGet, "/adjustments", nil, auth), http.StatusForbidden)
}
