// Package services contains domain services that interact with external systems.
package services

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jariesdev/vendoreport/internal/models"
)

const salesLogTypeIndex = 14      // log type index for purchase transactions
const coinInsertLogTypeIndex = 18 // log type index for "Inserted coin" events
const cancelTopupLogTypeIndex = 6 // log type index for "Cancel Topup" events

// SystemStatus holds all fields returned by the Juanfi dashboard API.
type SystemStatus struct {
	SystemUptimeMs   int64   `json:"system_uptime_ms"`
	TotalCoinCount   int     `json:"total_coin_count"`
	CurrentCoinCount int     `json:"current_coin_count"`
	CustomerCount    int     `json:"customer_count"`
	InternetStatus   bool    `json:"internet_status"`
	MikrotikStatus   bool    `json:"mikrotik_status"`
	MacAddress       string  `json:"mac_address"`
	IPAddress        string  `json:"ip_address"`
	HardwareType     string  `json:"hardware_type"`
	Version          float64 `json:"version"`
	InterfaceType    string  `json:"interface_type"`
	WirelessStrength int     `json:"wireless_signal_strength"`
	FreeHeap         int     `json:"free_heap"`
	AuthType         string  `json:"auth_type"`
	NightLightStatus bool    `json:"night_light_status"`
	ActiveUserCount  int     `json:"active_user_count"`
	SystemClock      string  `json:"system_clock"`
	ServerTime       float64 `json:"server_time"`
}

// SystemConfig holds the device's persistent configuration, as returned by
// api/getSystemConfig and written back by api/saveSystemConfig (same
// pipe-delimited positional format for both — saving is a direct round-trip
// of reading).
//
// Field order for indices 0-29 is confirmed from an older Juanfi admin
// console's populateSystemConfigFields() JS, which still matches this build's
// response byte-for-byte. Indices 30+ were added by newer firmware after
// that JS was captured; OperatorUsername/OperatorPassword/APIKey/
// CoinMultiplier/VoucherLength/BillAcceptorMultiplier/IncludeVendoName are
// confirmed by exact-value matches against a live device's admin panel. The
// rest of the new fields (marked "best-effort" below) are positional guesses
// based on the admin panel's field list and have not been confirmed against
// firmware source — verify against a live device before trusting them for
// anything beyond display.
type SystemConfig struct {
	VendoName           string `json:"vendo_name"`
	WiFiSSID            string `json:"wifi_ssid"`
	WiFiPassword        string `json:"wifi_password"`
	MikrotikIP          string `json:"mikrotik_ip"`
	MikrotikUsername    string `json:"mikrotik_username"`
	MikrotikPassword    string `json:"mikrotik_password"`
	CoinSlotWaitTimeSec int    `json:"coin_slot_wait_time_sec"`
	AdminUsername       string `json:"admin_username"`
	AdminPassword       string `json:"admin_password"`
	CoinSlotAbuseCount  int    `json:"coin_slot_abuse_count"`
	CoinSlotBanMinutes  int    `json:"coin_slot_ban_minutes"`
	CoinSlotPin         int    `json:"coin_slot_pin"`
	CoinSlotSetPin      int    `json:"coin_slot_set_pin"`
	SystemReadyLEDPin   int    `json:"system_ready_led_pin"`
	InsertCoinLEDPin    int    `json:"insert_coin_led_pin"`
	LCDScreen           int    `json:"lcd_screen"`
	InsertCoinButtonPin int    `json:"insert_coin_button_pin"`
	CheckInternetStatus bool   `json:"check_internet_status"`
	VoucherPrefix       string `json:"voucher_prefix"`
	WelcomeLCDMarquee   string `json:"welcome_lcd_marquee"`
	SetupDoneFlag       bool   `json:"setup_done_flag"`
	VoucherLoginOption  int    `json:"voucher_login_option"`
	VoucherProfile      string `json:"voucher_profile"`
	VoucherValidity     int    `json:"voucher_validity"`
	LEDTriggerType      int    `json:"led_trigger_type"`
	IPAddressMode       int    `json:"ip_address_mode"`
	LocalIPAddress      string `json:"local_ip_address"`
	GatewayIP           string `json:"gateway_ip"`
	SubnetMask          string `json:"subnet_mask"`
	DNSServer           string `json:"dns_server"`

	// --- Fields below this point were added by newer firmware. ---

	ConnectionMode         int    `json:"connection_mode"`          // best-effort
	CoinSlotType           int    `json:"coin_slot_type"`           // best-effort
	ButtonFunction         int    `json:"button_function"`          // best-effort
	OperatorUsername       string `json:"operator_username"`        // confirmed
	OperatorPassword       string `json:"operator_password"`        // confirmed
	APIKey                 string `json:"api_key"`                  // confirmed
	BillAcceptorPin        int    `json:"bill_acceptor_pin"`        // best-effort
	CoinMultiplier         int    `json:"coin_multiplier"`          // confirmed
	VoucherLength          int    `json:"voucher_length"`           // confirmed
	NightLightPin          int    `json:"night_light_pin"`          // best-effort
	LCDSDAPin              int    `json:"lcd_sda_pin"`              // best-effort
	LCDSCLPin              int    `json:"lcd_scl_pin"`              // best-effort
	LANCSPin               int    `json:"lan_cs_pin"`               // best-effort
	PrinterPin             int    `json:"printer_pin"`              // best-effort
	BillAcceptorMultiplier int    `json:"bill_acceptor_multiplier"` // confirmed
	PrintOption            int    `json:"print_option"`             // best-effort
	IncludeVendoName       bool   `json:"include_vendo_name"`       // confirmed

	// ExtraFields preserves any trailing fields this struct doesn't model
	// (observed as blank/zero padding past index 46), so SaveSystemConfig can
	// write them back unchanged instead of silently dropping unknown device
	// settings.
	ExtraFields []string `json:"extra_fields,omitempty"`
}

// systemConfigFieldCount is the number of leading fields this struct names
// explicitly (indices 0-46). Anything beyond that is kept in ExtraFields.
const systemConfigFieldCount = 47

// RawLog is a parsed row from the Juanfi getSystemLogs response.
type RawLog struct {
	HasHeader    bool
	Time         int64 // milliseconds since device startup
	LogTypeIndex int
	LogParams    []string
}

// FormattedLog is a human-readable log entry with a computed absolute timestamp.
type FormattedLog struct {
	LogTime     time.Time
	Description string
}

// JuanfiAPI is an HTTP client for a single Juanfi vendo machine.
// All requests carry an X-TOKEN header and append a Unix-ms timestamp query param.
type JuanfiAPI struct {
	baseURL      string
	apiKey       string
	systemUptime int64 // milliseconds; set after the first dashboard call
	httpClient   *http.Client
}

// NewJuanfiAPI constructs a JuanfiAPI for the given vendo machine.
func NewJuanfiAPI(vendo *models.Vendo) *JuanfiAPI {
	base := ""
	key := ""
	if vendo.APIURL != nil {
		base = *vendo.APIURL
	}
	if vendo.APIKey != nil {
		key = *vendo.APIKey
	}
	return &JuanfiAPI{
		baseURL:    base,
		apiKey:     key,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetSystemStatus fetches and parses the dashboard endpoint.
func (j *JuanfiAPI) GetSystemStatus() (*SystemStatus, error) {
	return j.loadSystemStatus()
}

// LoadSystemLogs fetches raw system logs from the device and returns them in
// chronological order (oldest first), ready for further processing.
func (j *JuanfiAPI) LoadSystemLogs() ([]RawLog, error) {
	// Dashboard must be called first to populate system uptime for time calculations.
	if _, err := j.loadSystemStatus(); err != nil {
		return nil, err
	}

	body, err := j.sendRequest("api/getSystemLogs", nil)
	if err != nil {
		return nil, err
	}
	if body == "" {
		return nil, nil
	}

	var rows []RawLog
	for _, chunk := range strings.Split(body, "|") {
		if chunk == "" {
			continue
		}

		hasHeader := false
		content := chunk
		if parts := strings.SplitN(chunk, "~", 2); len(parts) == 2 {
			content = parts[1]
			hasHeader = true
		}

		fields := strings.Split(content, "#")
		if len(fields) < 2 {
			continue
		}

		t, _ := strconv.ParseInt(fields[0], 10, 64)
		lt, _ := strconv.Atoi(fields[1])
		params := []string{}
		if len(fields) >= 3 {
			params = fields[2:]
		}

		rows = append(rows, RawLog{
			HasHeader:    hasHeader,
			Time:         t,
			LogTypeIndex: lt,
			LogParams:    params,
		})
	}

	// Reverse so index 0 is the oldest entry (chronological order).
	for i, k := 0, len(rows)-1; i < k; i, k = i+1, k-1 {
		rows[i], rows[k] = rows[k], rows[i]
	}
	return rows, nil
}

// GetFormattedLogs returns logs with human-readable timestamps and descriptions.
func (j *JuanfiAPI) GetFormattedLogs(rawLogs []RawLog) []FormattedLog {
	formatted := make([]FormattedLog, 0, len(rawLogs))
	for _, r := range rawLogs {
		formatted = append(formatted, FormattedLog{
			LogTime:     j.ComputeLogTime(r.Time),
			Description: j.FormatLogMessage(r.LogTypeIndex, r.LogParams),
		})
	}
	return formatted
}

// ComputeLogTime converts a device-relative millisecond timestamp into a wall
// clock time using the device's reported system uptime.
func (j *JuanfiAPI) ComputeLogTime(timeSinceStartup int64) time.Time {
	nowMs := time.Now().UnixMilli()
	absMs := (nowMs - j.systemUptime) + timeSinceStartup
	return time.UnixMilli(absMs)
}

// ActiveUser represents a single connected user returned by api/getActiveUsers.
type ActiveUser struct {
	NodeID      string `json:"node_id"`
	User        string `json:"user"`
	MacAddress  string `json:"mac_address"`
	SessionLeft string `json:"session_left"`
}

// GetActiveUsers fetches the list of currently connected users from the device.
func (j *JuanfiAPI) GetActiveUsers() ([]ActiveUser, error) {
	body, err := j.sendRequest("api/getActiveUsers", nil)
	if err != nil {
		return nil, err
	}

	var users []ActiveUser
	for _, row := range strings.Split(body, "|") {
		fields := strings.Split(row, "#")
		if len(fields) < 4 {
			continue
		}
		users = append(users, ActiveUser{
			NodeID:      fields[0],
			User:        fields[1],
			MacAddress:  fields[2],
			SessionLeft: fields[3],
		})
	}
	return users, nil
}

// ResetCurrentSales calls the Juanfi API to reset the current sales counter.
func (j *JuanfiAPI) ResetCurrentSales() error {
	_, err := j.sendRequest("api/resetStatistic", map[string]string{"type": "coinCount"})
	return err
}

// Rate is a single pricing tier as configured on the device.
type Rate struct {
	Name         string
	Price        float64
	Minutes      int
	ValidityMins int
	DataLimitMB  *int   // nil when the device leaves the field blank
	UserProfile  string // "default" (Mikrotik's default hotspot profile) when the device leaves the field blank
}

// GeneratedVoucher is a single voucher code returned by api/generateVouchers.
// VendoName, Amount, and Duration are shared across the whole generated batch.
type GeneratedVoucher struct {
	VendoName string
	Amount    float64
	Duration  int // minutes
	Code      string
}

// GetRates fetches and parses the rate plans configured on the device.
func (j *JuanfiAPI) GetRates() ([]Rate, error) {
	body, err := j.sendRequest("api/getRates", nil)
	if err != nil {
		return nil, err
	}

	var rates []Rate
	for _, chunk := range strings.Split(body, "|") {
		if chunk == "" {
			continue
		}

		fields := strings.Split(chunk, "#")
		if fields[0] == "" {
			continue
		}

		// The device may emit a final entry with trailing/missing fields
		// (e.g. "Name#"), so read each field defensively and default to zero
		// rather than skipping the whole entry.
		var price float64
		if len(fields) > 1 {
			price, _ = strconv.ParseFloat(fields[1], 64)
		}
		var minutes int
		if len(fields) > 2 {
			minutes, _ = strconv.Atoi(fields[2])
		}
		var validity int
		if len(fields) > 3 {
			validity, _ = strconv.Atoi(fields[3])
		}

		var dataLimit *int
		if len(fields) > 4 && fields[4] != "" {
			if v, err := strconv.Atoi(fields[4]); err == nil {
				dataLimit = &v
			}
		}

		// "default" is Mikrotik's default hotspot user profile, used by the
		// device whenever a rate doesn't specify an override profile.
		userProfile := "default"
		if len(fields) > 5 && fields[5] != "" {
			userProfile = fields[5]
		}

		rates = append(rates, Rate{
			Name:         fields[0],
			Price:        price,
			Minutes:      minutes,
			ValidityMins: validity,
			DataLimitMB:  dataLimit,
			UserProfile:  userProfile,
		})
	}
	return rates, nil
}

// FormatLogMessage renders a log type template with the provided parameters.
// Templates are indexed by log_type_index from the Juanfi firmware.
func (j *JuanfiAPI) FormatLogMessage(logType int, params []string) string {
	templates := logTypeTemplates()
	if logType < 0 || logType >= len(templates) {
		return fmt.Sprintf("unknown log type %d", logType)
	}
	template := templates[logType]
	// Replace positional placeholders {0}, {1}, {2} with actual params.
	for i, p := range params {
		template = strings.ReplaceAll(template, fmt.Sprintf("{%d}", i), p)
	}
	return template
}

// loadSystemStatus calls the dashboard endpoint and stores the uptime.
func (j *JuanfiAPI) loadSystemStatus() (*SystemStatus, error) {
	body, err := j.sendRequest("api/dashboard", nil)
	if err != nil {
		return nil, err
	}

	data := strings.Split(body, "|")
	if len(data) < 17 {
		return nil, fmt.Errorf("dashboard response too short: got %d fields", len(data))
	}

	uptime, _ := strconv.ParseInt(data[0], 10, 64)
	j.systemUptime = uptime

	totalCoins, _ := strconv.Atoi(data[1])
	currentCoins, _ := strconv.Atoi(data[2])
	customers, _ := strconv.Atoi(data[3])
	version, _ := strconv.ParseFloat(data[9], 64)
	wirelessStrength, _ := strconv.Atoi(data[11])
	freeHeap, _ := strconv.Atoi(data[12])
	activeUsers, _ := strconv.Atoi(data[15])

	return &SystemStatus{
		SystemUptimeMs:   uptime,
		TotalCoinCount:   totalCoins,
		CurrentCoinCount: currentCoins,
		CustomerCount:    customers,
		InternetStatus:   data[4] == "1",
		MikrotikStatus:   data[5] == "1",
		MacAddress:       data[6],
		IPAddress:        data[7],
		HardwareType:     data[8],
		Version:          version,
		InterfaceType:    data[10],
		WirelessStrength: wirelessStrength,
		FreeHeap:         freeHeap,
		AuthType:         data[13],
		NightLightStatus: data[14] == "1",
		ActiveUserCount:  activeUsers,
		SystemClock:      data[16],
		ServerTime:       float64(time.Now().UnixMilli()),
	}, nil
}

// buildURL constructs the full request URL for a device endpoint path.
// A Unix-ms timestamp is always appended as the `query` parameter, plus any
// extra query params the caller supplies.
func (j *JuanfiAPI) buildURL(path string, extraQuery map[string]string) string {
	base := strings.TrimRight(j.baseURL, "/")
	endpoint := fmt.Sprintf("%s/admin/%s", base, path)

	params := url.Values{}
	params.Set("query", strconv.FormatInt(time.Now().UnixMilli(), 10))
	for k, v := range extraQuery {
		params.Set(k, v)
	}
	return endpoint + "?" + params.Encode()
}

// sendRequest performs an authenticated GET to the vendo machine API.
func (j *JuanfiAPI) sendRequest(path string, extraQuery map[string]string) (string, error) {
	fullURL := j.buildURL(path, extraQuery)
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return "", fmt.Errorf("juanfi: build request: %w", err)
	}
	req.Header.Set("X-TOKEN", j.apiKey)

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("juanfi: request to %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("juanfi: unexpected status %d from %s", resp.StatusCode, path)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("juanfi: read response body: %w", err)
	}
	return string(bodyBytes), nil
}

// SaveRates pushes the given rate plans to the device, replacing whatever
// rate plan it currently has configured under rateType 1 (the WiFi/hotspot
// rate plan — the only rate table vendo machines use).
func (j *JuanfiAPI) SaveRates(rates []Rate) error {
	fullURL := j.buildURL("api/saveRates", map[string]string{"rateType": "1"})

	form := url.Values{}
	form.Set("data", encodeRates(rates))

	req, err := http.NewRequest(http.MethodPost, fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("juanfi: build request: %w", err)
	}
	req.Header.Set("X-TOKEN", j.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("juanfi: request to %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("juanfi: unexpected status %d from api/saveRates", resp.StatusCode)
	}
	return nil
}

// GetSystemConfig fetches and parses the device's persistent configuration.
func (j *JuanfiAPI) GetSystemConfig() (*SystemConfig, error) {
	body, err := j.sendRequest("api/getSystemConfig", nil)
	if err != nil {
		return nil, err
	}

	data := strings.Split(body, "|")
	if len(data) < systemConfigFieldCount {
		return nil, fmt.Errorf("getSystemConfig response too short: got %d fields", len(data))
	}

	atoi := func(s string) int {
		v, _ := strconv.Atoi(s)
		return v
	}

	cfg := &SystemConfig{
		VendoName:           data[0],
		WiFiSSID:            data[1],
		WiFiPassword:        data[2],
		MikrotikIP:          data[3],
		MikrotikUsername:    data[4],
		MikrotikPassword:    data[5],
		CoinSlotWaitTimeSec: atoi(data[6]),
		AdminUsername:       data[7],
		AdminPassword:       data[8],
		CoinSlotAbuseCount:  atoi(data[9]),
		CoinSlotBanMinutes:  atoi(data[10]),
		CoinSlotPin:         atoi(data[11]),
		CoinSlotSetPin:      atoi(data[12]),
		SystemReadyLEDPin:   atoi(data[13]),
		InsertCoinLEDPin:    atoi(data[14]),
		LCDScreen:           atoi(data[15]),
		InsertCoinButtonPin: atoi(data[16]),
		CheckInternetStatus: data[17] == "1",
		VoucherPrefix:       data[18],
		WelcomeLCDMarquee:   data[19],
		SetupDoneFlag:       data[20] == "1",
		VoucherLoginOption:  atoi(data[21]),
		VoucherProfile:      data[22],
		VoucherValidity:     atoi(data[23]),
		LEDTriggerType:      atoi(data[24]),
		IPAddressMode:       atoi(data[25]),
		LocalIPAddress:      data[26],
		GatewayIP:           data[27],
		SubnetMask:          data[28],
		DNSServer:           data[29],

		ConnectionMode:         atoi(data[30]),
		CoinSlotType:           atoi(data[31]),
		ButtonFunction:         atoi(data[32]),
		OperatorUsername:       data[33],
		OperatorPassword:       data[34],
		APIKey:                 data[35],
		BillAcceptorPin:        atoi(data[36]),
		CoinMultiplier:         atoi(data[37]),
		VoucherLength:          atoi(data[38]),
		NightLightPin:          atoi(data[39]),
		LCDSDAPin:              atoi(data[40]),
		LCDSCLPin:              atoi(data[41]),
		LANCSPin:               atoi(data[42]),
		PrinterPin:             atoi(data[43]),
		BillAcceptorMultiplier: atoi(data[44]),
		PrintOption:            atoi(data[45]),
		IncludeVendoName:       data[46] == "1",
	}
	if len(data) > systemConfigFieldCount {
		cfg.ExtraFields = data[systemConfigFieldCount:]
	}
	return cfg, nil
}

// SaveSystemConfig pushes the given configuration to the device, replacing
// whatever configuration it currently has. Fields this struct doesn't model
// (ExtraFields) are written back unchanged to avoid corrupting device
// settings this client doesn't understand. The device restarts after a
// successful save to apply the new configuration.
func (j *JuanfiAPI) SaveSystemConfig(cfg *SystemConfig) error {
	fullURL := j.buildURL("api/saveSystemConfig", nil)

	form := url.Values{}
	form.Set("data", encodeSystemConfig(cfg))

	req, err := http.NewRequest(http.MethodPost, fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("juanfi: build request: %w", err)
	}
	req.Header.Set("X-TOKEN", j.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("juanfi: request to %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("juanfi: unexpected status %d from api/saveSystemConfig", resp.StatusCode)
	}
	return nil
}

// encodeSystemConfig serialises a SystemConfig back into the device's
// pipe-delimited positional format — the inverse of GetSystemConfig's
// parsing, including any unmodeled ExtraFields tacked back on at the end.
func encodeSystemConfig(cfg *SystemConfig) string {
	boolStr := func(b bool) string {
		if b {
			return "1"
		}
		return "0"
	}

	fields := []string{
		cfg.VendoName,
		cfg.WiFiSSID,
		cfg.WiFiPassword,
		cfg.MikrotikIP,
		cfg.MikrotikUsername,
		cfg.MikrotikPassword,
		strconv.Itoa(cfg.CoinSlotWaitTimeSec),
		cfg.AdminUsername,
		cfg.AdminPassword,
		strconv.Itoa(cfg.CoinSlotAbuseCount),
		strconv.Itoa(cfg.CoinSlotBanMinutes),
		strconv.Itoa(cfg.CoinSlotPin),
		strconv.Itoa(cfg.CoinSlotSetPin),
		strconv.Itoa(cfg.SystemReadyLEDPin),
		strconv.Itoa(cfg.InsertCoinLEDPin),
		strconv.Itoa(cfg.LCDScreen),
		strconv.Itoa(cfg.InsertCoinButtonPin),
		boolStr(cfg.CheckInternetStatus),
		cfg.VoucherPrefix,
		cfg.WelcomeLCDMarquee,
		boolStr(cfg.SetupDoneFlag),
		strconv.Itoa(cfg.VoucherLoginOption),
		cfg.VoucherProfile,
		strconv.Itoa(cfg.VoucherValidity),
		strconv.Itoa(cfg.LEDTriggerType),
		strconv.Itoa(cfg.IPAddressMode),
		cfg.LocalIPAddress,
		cfg.GatewayIP,
		cfg.SubnetMask,
		cfg.DNSServer,

		strconv.Itoa(cfg.ConnectionMode),
		strconv.Itoa(cfg.CoinSlotType),
		strconv.Itoa(cfg.ButtonFunction),
		cfg.OperatorUsername,
		cfg.OperatorPassword,
		cfg.APIKey,
		strconv.Itoa(cfg.BillAcceptorPin),
		strconv.Itoa(cfg.CoinMultiplier),
		strconv.Itoa(cfg.VoucherLength),
		strconv.Itoa(cfg.NightLightPin),
		strconv.Itoa(cfg.LCDSDAPin),
		strconv.Itoa(cfg.LCDSCLPin),
		strconv.Itoa(cfg.LANCSPin),
		strconv.Itoa(cfg.PrinterPin),
		strconv.Itoa(cfg.BillAcceptorMultiplier),
		strconv.Itoa(cfg.PrintOption),
		boolStr(cfg.IncludeVendoName),
	}
	fields = append(fields, cfg.ExtraFields...)
	return strings.Join(fields, "|")
}

// GenerateVouchers calls api/generateVouchers to create qty new prepaid vouchers
// at the given price. The device matches amount against its own configured rate
// plan to determine duration; amount/duration are not chosen by the caller.
func (j *JuanfiAPI) GenerateVouchers(prefix string, amount float64, qty int, addToSales bool, printThermal bool) ([]GeneratedVoucher, error) {
	fullURL := j.buildURL("api/generateVouchers", nil)

	form := url.Values{}
	form.Set("amt", strconv.FormatFloat(amount, 'f', -1, 64))
	form.Set("pfx", prefix)
	form.Set("qty", strconv.Itoa(qty))
	form.Set("sales", boolFlag(addToSales))
	form.Set("print", boolFlag(printThermal))

	req, err := http.NewRequest(http.MethodPost, fullURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("juanfi: build request: %w", err)
	}
	req.Header.Set("X-TOKEN", j.apiKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	resp, err := j.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("juanfi: request to %s: %w", fullURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("juanfi: unexpected status %d from api/generateVouchers", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("juanfi: read response body: %w", err)
	}
	return parseGeneratedVouchers(string(body))
}

func boolFlag(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// encodeRates serialises rates back into the device's pipe/hash-delimited
// format — the inverse of GetRates' parsing.
func encodeRates(rates []Rate) string {
	entries := make([]string, len(rates))
	for i, r := range rates {
		dataLimit := ""
		if r.DataLimitMB != nil {
			dataLimit = strconv.Itoa(*r.DataLimitMB)
		}
		entries[i] = strings.Join([]string{
			r.Name,
			strconv.FormatFloat(r.Price, 'f', -1, 64),
			strconv.Itoa(r.Minutes),
			strconv.Itoa(r.ValidityMins),
			dataLimit,
			r.UserProfile,
		}, "#")
	}
	return strings.Join(entries, "|")
}

// parseGeneratedVouchers parses the api/generateVouchers response, e.g.
// "Your WiFi|2|2400|VC5416#VC2879" — vendo name, amount, duration (minutes),
// then a "#"-delimited list of voucher codes sharing that amount/duration.
func parseGeneratedVouchers(body string) ([]GeneratedVoucher, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return nil, nil
	}
	fields := strings.SplitN(body, "|", 4)
	if len(fields) < 4 {
		return nil, fmt.Errorf("juanfi: unexpected generateVouchers response format: %q", body)
	}

	name := fields[0]
	amount, _ := strconv.ParseFloat(fields[1], 64)
	duration, _ := strconv.Atoi(fields[2])

	codes := strings.Split(fields[3], "#")
	vouchers := make([]GeneratedVoucher, 0, len(codes))
	for _, code := range codes {
		if code == "" {
			continue
		}
		vouchers = append(vouchers, GeneratedVoucher{
			VendoName: name,
			Amount:    amount,
			Duration:  duration,
			Code:      code,
		})
	}
	return vouchers, nil
}

// logTypeTemplates returns the ordered slice of message templates indexed by
// log_type_index from the Juanfi firmware (33 entries total).
func logTypeTemplates() []string {
	return []string{
		"No log message",                                         // 0
		"System Starting...",                                     // 1
		"Network Connected Succesfully",                          // 2
		"Mikrotik Connected Succesfully",                         // 3
		"Juanfi Initial Setup",                                   // 4
		"Failed to login to mikrotik",                            // 5
		"{0} Cancel Topup",                                       // 6
		"Rates modified",                                         // 7
		"System Configuration Modified",                          // 8
		"Reset Total Sales",                                      // 9
		"Reset Current sales",                                    // 10
		"Reset Customer Count",                                   // 11
		"{0} Login Sucessfully",                                  // 12
		"{0} Login Failed",                                       // 13
		"{0} Purchase {1}, amount: {2}",                          // 14 ← sale log
		"{0} Attempted to insert coin",                           // 15
		"{0} was banned from using coinslot",                     // 16
		"Create voucher failed {0}, retrying...",                 // 17
		"{0} Inserted coin {1}",                                  // 18
		"Manual voucher purchase activated",                      // 19
		"Generated {0} voucher(s)",                               // 20
		"NightLight Turn On",                                     // 21
		"NightLight Turn Off",                                    // 22
		"Kick active user {0}",                                   // 23
		"{0} tried to insert coin but no internet available",     // 24
		"Reset Daily Sales",                                      // 25
		"Reset Monthly Sales",                                    // 26
		"Update charging settings",                               // 27
		"{0} Purchase Eload success {1}",                         // 28
		"Update Eload settings",                                  // 29
		"Update Eload rates",                                     // 30
		"Clear Eload transactions",                               // 31
		"{0} Purchase Eload failed {1}",                          // 32
		"Unusual coinslot pulse detected, please check coinslot", // 33
	}
}
