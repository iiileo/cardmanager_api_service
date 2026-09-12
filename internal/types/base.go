package types

import "time"

type Status string

const (
	StatusPublished Status = "published"
	StatusDeleted   Status = "deleted"
)

type BaseModel struct {
	TenantID  string    `json:"tenant_id"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy string    `json:"created_by,omitempty"`
	UpdatedBy string    `json:"updated_by,omitempty"`
}
