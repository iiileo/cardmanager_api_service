package accesslog

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Middleware(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		if store == nil || !store.Enabled() {
			c.Next()
			return
		}
		if shouldSkipPath(c.Request.URL.Path, store.SkipPaths()) {
			c.Next()
			return
		}

		start := time.Now()
		ctx, rec := WithRecorder(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		maxBody := store.MaxBodyBytes()
		reqBody := readRequestBody(c.Request, maxBody)
		capture := newRespCapture(c.Writer, maxBody)
		c.Writer = capture

		c.Next()

		status := c.Writer.Status()
		if status == 0 {
			status = http.StatusOK
		}

		var errMsg string
		if len(c.Errors) > 0 {
			errMsg = c.Errors.Last().Err.Error()
		}

		entry := Entry{
			ID:           uuid.NewString(),
			Time:         start,
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			Query:        c.Request.URL.RawQuery,
			ClientIP:     c.ClientIP(),
			RequestBody:  reqBody,
			Status:       status,
			ResponseBody: capture.snapshot(),
			DurationMs:   float64(time.Since(start).Microseconds()) / 1000,
			SQL:          rec.Snapshot(),
			Error:        errMsg,
		}
		_ = store.Append(entry)
	}
}
