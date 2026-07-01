package services

import (
	"fmt"
	"math"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
	"github.com/jariesdev/vendoreport/internal/repository"
)

// profitPHT is the Philippine Time zone (UTC+8). Month buckets are computed in
// PHT so they line up with the sales aggregation ('+8 hours' offset).
var profitPHT = time.FixedZone("PHT", 8*60*60)

// forecastAvgMonths is how many trailing complete months are averaged to
// estimate recurring monthly revenue for the payback forecast.
const forecastAvgMonths = 3

// forecastHorizonMonths caps how far the break-even projection runs forward.
const forecastHorizonMonths = 60

// ProfitService computes per-owner revenue-vs-expense reports and a
// payback/break-even forecast. Revenue comes from the owner's own vendo sales;
// expenses come from the owner's own expense entries. All business logic lives
// here so controllers stay thin.
type ProfitService struct {
	saleRepo    repository.SaleRepositoryInterface
	expenseRepo repository.ExpenseRepositoryInterface
}

func NewProfitService(saleRepo repository.SaleRepositoryInterface, expenseRepo repository.ExpenseRepositoryInterface) *ProfitService {
	return &ProfitService{saleRepo: saleRepo, expenseRepo: expenseRepo}
}

// MonthlyReportRow is a single month of the profit report.
type MonthlyReportRow struct {
	Month         string  `json:"month"` // YYYY-MM
	Revenue       float64 `json:"revenue"`
	OneTime       float64 `json:"one_time"`
	Recurring     float64 `json:"recurring"`
	Expenses      float64 `json:"expenses"`
	Net           float64 `json:"net"`
	CumulativeNet float64 `json:"cumulative_net"`
}

// YearlyReportRow rolls the monthly rows up by calendar year.
type YearlyReportRow struct {
	Year          string  `json:"year"`
	Revenue       float64 `json:"revenue"`
	Expenses      float64 `json:"expenses"`
	Net           float64 `json:"net"`
	CumulativeNet float64 `json:"cumulative_net"`
}

// ReportSummary is the window total and the resulting profit status.
type ReportSummary struct {
	TotalRevenue  float64 `json:"total_revenue"`
	TotalExpenses float64 `json:"total_expenses"`
	Net           float64 `json:"net"`
	Status        string  `json:"status"` // profit | break_even | loss
}

// ProfitReport is the full response for the report endpoint.
type ProfitReport struct {
	Monthly []MonthlyReportRow `json:"monthly"`
	Yearly  []YearlyReportRow  `json:"yearly"`
	Summary ReportSummary      `json:"summary"`
}

// ForecastPoint is one month of the forward break-even projection.
type ForecastPoint struct {
	Month                  string  `json:"month"`
	ProjectedCumulativeNet float64 `json:"projected_cumulative_net"`
}

// ProfitForecast is the payback/break-even analysis response.
type ProfitForecast struct {
	AvgMonthlyRevenue       float64         `json:"avg_monthly_revenue"`
	MonthlyRecurringExpense float64         `json:"monthly_recurring_expense"`
	NetMonthly              float64         `json:"net_monthly"`
	CumulativeNet           float64         `json:"cumulative_net"`
	Status                  string          `json:"status"` // recovered | on_track | not_recovering
	MonthsToBreakEven       *int            `json:"months_to_break_even"`
	ProjectedBreakEvenMonth *string         `json:"projected_break_even_month"`
	Projection              []ForecastPoint `json:"projection"`
}

// Report builds the monthly/yearly profit report for the owner over [from, to).
// ownVendoIDs are the owner's own vendo IDs; an empty slice means the owner has
// no vendos and revenue is zero (we never fall back to "all vendos").
func (s *ProfitService) Report(ownVendoIDs []uint, userID uint, from, to time.Time) (*ProfitReport, error) {
	// Revenue per month (map[YYYY-MM]total). Empty ownVendoIDs => no revenue;
	// we must not call GetMonthlySales, whose empty-slice branch means "all vendos".
	revenue := map[string]float64{}
	if len(ownVendoIDs) > 0 {
		rows, err := s.saleRepo.GetMonthlySales(from, to, ownVendoIDs)
		if err != nil {
			return nil, fmt.Errorf("monthly sales: %w", err)
		}
		for _, r := range rows {
			revenue[r.Month] += r.Total
		}
	}

	// One-time (non-recurring) expenses per month.
	oneTime := map[string]float64{}
	otRows, err := s.expenseRepo.MonthlyOneTime(userID, from, to)
	if err != nil {
		return nil, fmt.Errorf("one-time expenses: %w", err)
	}
	for _, r := range otRows {
		oneTime[r.Month] += r.Amount
	}

	// Recurring expenses: each contributes Amount to every month within its
	// [start, end] span (end open when EndDate is nil).
	recurring, err := s.expenseRepo.ActiveRecurring(userID)
	if err != nil {
		return nil, fmt.Errorf("recurring expenses: %w", err)
	}
	spans := toSpans(recurring)

	// Determine the month window: from the earliest month that has any data
	// (or the requested `from`, whichever is later) through `to`.
	earliest := earliestKey(revenue, oneTime)
	for _, sp := range spans {
		if earliest == "" || sp.start < earliest {
			earliest = sp.start
		}
	}
	startMonth := maxMonthKey(monthKey(from), earliest)
	endMonth := monthKey(to)
	months := monthRange(startMonth, endMonth)

	report := &ProfitReport{Monthly: []MonthlyReportRow{}, Yearly: []YearlyReportRow{}}
	yearly := map[string]*YearlyReportRow{}
	yearOrder := []string{}

	var cumulative float64
	for _, m := range months {
		rec := recurringForMonth(spans, m)
		row := MonthlyReportRow{
			Month:     m,
			Revenue:   round2(revenue[m]),
			OneTime:   round2(oneTime[m]),
			Recurring: round2(rec),
		}
		row.Expenses = round2(row.OneTime + row.Recurring)
		row.Net = round2(row.Revenue - row.Expenses)
		cumulative = round2(cumulative + row.Net)
		row.CumulativeNet = cumulative
		report.Monthly = append(report.Monthly, row)

		report.Summary.TotalRevenue = round2(report.Summary.TotalRevenue + row.Revenue)
		report.Summary.TotalExpenses = round2(report.Summary.TotalExpenses + row.Expenses)

		year := m[:4]
		yr, ok := yearly[year]
		if !ok {
			yr = &YearlyReportRow{Year: year}
			yearly[year] = yr
			yearOrder = append(yearOrder, year)
		}
		yr.Revenue = round2(yr.Revenue + row.Revenue)
		yr.Expenses = round2(yr.Expenses + row.Expenses)
		yr.Net = round2(yr.Net + row.Net)
	}

	var yc float64
	for _, y := range yearOrder {
		yr := yearly[y]
		yc = round2(yc + yr.Net)
		yr.CumulativeNet = yc
		report.Yearly = append(report.Yearly, *yr)
	}

	report.Summary.Net = round2(report.Summary.TotalRevenue - report.Summary.TotalExpenses)
	report.Summary.Status = profitStatus(report.Summary.Net)
	return report, nil
}

// Forecast projects when (or whether) the owner recovers their investment,
// using the standard payback period = unrecovered net ÷ net monthly cash flow.
func (s *ProfitService) Forecast(ownVendoIDs []uint, userID uint) (*ProfitForecast, error) {
	now := time.Now().In(profitPHT)
	from := time.Date(2000, 1, 1, 0, 0, 0, 0, profitPHT)

	report, err := s.Report(ownVendoIDs, userID, from, now)
	if err != nil {
		return nil, err
	}

	// Average revenue over the last few COMPLETE months (exclude the partial
	// current month so it doesn't drag the average down).
	current := monthKey(now)
	var revs []float64
	for i := len(report.Monthly) - 1; i >= 0 && len(revs) < forecastAvgMonths; i-- {
		if report.Monthly[i].Month >= current {
			continue
		}
		revs = append(revs, report.Monthly[i].Revenue)
	}
	var avgRevenue float64
	if len(revs) > 0 {
		var sum float64
		for _, v := range revs {
			sum += v
		}
		avgRevenue = sum / float64(len(revs))
	}

	// Current recurring monthly drain = recurring amounts active this month
	// (started on/before now and not yet ended).
	recurring, err := s.expenseRepo.ActiveRecurring(userID)
	if err != nil {
		return nil, fmt.Errorf("recurring expenses: %w", err)
	}
	monthlyRecurring := recurringForMonth(toSpans(recurring), current)

	netMonthly := round2(avgRevenue - monthlyRecurring)
	cumulative := report.Summary.Net

	f := &ProfitForecast{
		AvgMonthlyRevenue:       round2(avgRevenue),
		MonthlyRecurringExpense: round2(monthlyRecurring),
		NetMonthly:              netMonthly,
		CumulativeNet:           cumulative,
		Projection:              []ForecastPoint{},
	}

	switch {
	case cumulative >= 0:
		f.Status = "recovered"
	case netMonthly > 0:
		f.Status = "on_track"
		months := int(math.Ceil(-cumulative / netMonthly))
		f.MonthsToBreakEven = &months
		proj := cumulative
		m := current
		for i := 1; i <= forecastHorizonMonths; i++ {
			m = addMonths(m, 1)
			proj = round2(proj + netMonthly)
			f.Projection = append(f.Projection, ForecastPoint{Month: m, ProjectedCumulativeNet: proj})
			if proj >= 0 && f.ProjectedBreakEvenMonth == nil {
				bm := m
				f.ProjectedBreakEvenMonth = &bm
			}
		}
	default:
		f.Status = "not_recovering"
	}

	return f, nil
}

// ---- month helpers (YYYY-MM keys, PHT) ----

func monthKey(t time.Time) string { return t.In(profitPHT).Format("2006-01") }

// recurringSpan is a recurring expense's monthly amount over its active window.
// end == "" means open-ended (no end date).
type recurringSpan struct {
	start  string
	end    string
	amount float64
}

// toSpans converts recurring expenses into month-keyed active spans.
func toSpans(expenses []models.Expense) []recurringSpan {
	spans := make([]recurringSpan, 0, len(expenses))
	for _, e := range expenses {
		sp := recurringSpan{start: monthKey(e.ExpenseDate), amount: e.Amount}
		if e.EndDate != nil {
			sp.end = monthKey(*e.EndDate)
		}
		spans = append(spans, sp)
	}
	return spans
}

// recurringForMonth sums the monthly amounts of spans active during month m
// (started on/before m and not ended before m).
func recurringForMonth(spans []recurringSpan, m string) float64 {
	var sum float64
	for _, sp := range spans {
		if sp.start <= m && (sp.end == "" || m <= sp.end) {
			sum += sp.amount
		}
	}
	return sum
}

// earliestKey returns the smallest month key present across the given maps, or "".
func earliestKey(maps ...map[string]float64) string {
	earliest := ""
	for _, mp := range maps {
		for k := range mp {
			if earliest == "" || k < earliest {
				earliest = k
			}
		}
	}
	return earliest
}

func maxMonthKey(a, b string) string {
	if b == "" || a > b {
		return a
	}
	return b
}

// monthRange returns an inclusive continuous list of YYYY-MM keys from start to
// end. Returns just [start] if end < start.
func monthRange(start, end string) []string {
	if start == "" {
		return nil
	}
	if end < start {
		return []string{start}
	}
	var months []string
	m := start
	for m <= end {
		months = append(months, m)
		if len(months) > forecastHorizonMonths*12+24 { // safety bound
			break
		}
		m = addMonths(m, 1)
	}
	return months
}

// addMonths advances a YYYY-MM key by n months.
func addMonths(key string, n int) string {
	var y, mo int
	fmt.Sscanf(key, "%d-%d", &y, &mo)
	t := time.Date(y, time.Month(mo), 1, 0, 0, 0, 0, profitPHT).AddDate(0, n, 0)
	return t.Format("2006-01")
}

func profitStatus(net float64) string {
	switch {
	case net > 0.005:
		return "profit"
	case net < -0.005:
		return "loss"
	default:
		return "break_even"
	}
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
