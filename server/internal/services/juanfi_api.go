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

const salesLogTypeIndex = 14 // log type index for purchase transactions

// SystemStatus holds all fields returned by the Juanfi dashboard API.
type SystemStatus struct {
	SystemUptimeMs      int64   `json:"system_uptime_ms"`
	TotalCoinCount      int     `json:"total_coin_count"`
	CurrentCoinCount    int     `json:"current_coin_count"`
	CustomerCount       int     `json:"customer_count"`
	InternetStatus      bool    `json:"internet_status"`
	MikrotikStatus      bool    `json:"mikrotik_status"`
	MacAddress          string  `json:"mac_address"`
	IPAddress           string  `json:"ip_address"`
	HardwareType        string  `json:"hardware_type"`
	Version             float64 `json:"version"`
	InterfaceType       string  `json:"interface_type"`
	WirelessStrength    int     `json:"wireless_signal_strength"`
	FreeHeap            int     `json:"free_heap"`
	AuthType            string  `json:"auth_type"`
	NightLightStatus    bool    `json:"night_light_status"`
	ActiveUserCount     int     `json:"active_user_count"`
	SystemClock         string  `json:"system_clock"`
	ServerTime          float64 `json:"server_time"`
}

// RawLog is a parsed row from the Juanfi getSystemLogs response.
type RawLog struct {
	HasHeader    bool
	Time         int64    // milliseconds since device startup
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

// ResetCurrentSales calls the Juanfi API to reset the current sales counter.
func (j *JuanfiAPI) ResetCurrentSales() error {
	_, err := j.sendRequest("api/resetStatistic", map[string]string{"type": "coinCount"})
	return err
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

// sendRequest performs an authenticated GET to the vendo machine API.
// A Unix-ms timestamp is appended as the `query` parameter on every request.
func (j *JuanfiAPI) sendRequest(path string, extraQuery map[string]string) (string, error) {
	base := strings.TrimRight(j.baseURL, "/")
	endpoint := fmt.Sprintf("%s/admin/%s", base, path)

	params := url.Values{}
	params.Set("query", strconv.FormatInt(time.Now().UnixMilli(), 10))
	for k, v := range extraQuery {
		params.Set(k, v)
	}

	fullURL := endpoint + "?" + params.Encode()
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
