package accesslog

import "time"

// Entry 单次 HTTP 访问记录。
type Entry struct {
	ID             string       `json:"id"`
	Time           time.Time    `json:"time"`
	Method         string       `json:"method"`
	Path           string       `json:"path"`
	Query          string       `json:"query,omitempty"`
	ClientIP       string       `json:"client_ip"`
	RequestBody    string       `json:"request_body,omitempty"`
	Status         int          `json:"status"`
	ResponseBody   string       `json:"response_body,omitempty"`
	DurationMs     float64      `json:"duration_ms"`
	SQL            []SQLRecord  `json:"sql,omitempty"`
	Error          string       `json:"error,omitempty"`
}
