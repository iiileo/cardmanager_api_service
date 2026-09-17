package accesslog

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func readRequestBody(r *http.Request, max int) string {
	if r.Body == nil {
		return ""
	}
	data, _ := io.ReadAll(io.LimitReader(r.Body, int64(max)+1))
	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(data))
	if len(data) == 0 {
		return ""
	}
	return redactBody(truncate(string(data), max))
}

type respCapture struct {
	gin.ResponseWriter
	buf bytes.Buffer
	max int
}

func newRespCapture(w gin.ResponseWriter, max int) *respCapture {
	return &respCapture{ResponseWriter: w, max: max}
}

func (w *respCapture) Write(b []byte) (int, error) {
	if w.buf.Len() < w.max {
		remain := w.max - w.buf.Len()
		if len(b) > remain {
			w.buf.Write(b[:remain])
		} else {
			w.buf.Write(b)
		}
	}
	return w.ResponseWriter.Write(b)
}

func (w *respCapture) snapshot() string {
	if w.buf.Len() == 0 {
		return ""
	}
	s := w.buf.String()
	if w.buf.Len() >= w.max {
		s += "...(truncated)"
	}
	return redactBody(s)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

func redactBody(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if json.Valid([]byte(raw)) {
		var v any
		if err := json.Unmarshal([]byte(raw), &v); err == nil {
			redactValue(v)
			b, err := json.Marshal(v)
			if err == nil {
				return string(b)
			}
		}
	}
	return raw
}

func redactValue(v any) {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if isSensitiveKey(k) {
				x[k] = "***"
			} else {
				redactValue(val)
			}
		}
	case []any:
		for i := range x {
			redactValue(x[i])
		}
	}
}

func isSensitiveKey(key string) bool {
	switch strings.ToLower(key) {
	case "password", "token", "refresh_token", "access_token",
		"authorization", "jwt", "secret", "jwt_secret", "sms_code":
		return true
	default:
		return false
	}
}

func shouldSkipPath(path string, skip []string) bool {
	for _, p := range skip {
		if p != "" && path == p {
			return true
		}
	}
	return false
}
