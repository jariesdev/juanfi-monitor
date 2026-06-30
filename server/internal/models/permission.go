package models

const (
	PermDashboard   = "dashboard"
	PermAccount     = "account"
	PermVendos      = "vendos"
	PermSales       = "sales"
	PermLogs        = "logs"
	PermWithdrawals = "withdrawals"
	PermRates       = "rates"
	PermVouchers    = "vouchers"
	PermSettings    = "settings"
	PermUsers       = "users"
)

// AllPermissions returns all defined permission constants in display order.
func AllPermissions() []string {
	return []string{
		PermDashboard,
		PermAccount,
		PermVendos,
		PermSales,
		PermLogs,
		PermWithdrawals,
		PermRates,
		PermVouchers,
		PermSettings,
		PermUsers,
	}
}

// defaultPermissions are granted to every user regardless of role.
var defaultPermissions = map[string]bool{
	PermDashboard: true,
	PermAccount:   true,
}
