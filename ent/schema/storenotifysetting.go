package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// StoreNotifySetting 门店通知配置（按事件类型）。
type StoreNotifySetting struct {
	ent.Schema
}

func (StoreNotifySetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "store_notify_settings"},
	}
}

func (StoreNotifySetting) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("store_id"),
		field.String("event").MaxLen(16).NotEmpty(), // open|recharge|consume|count
		field.Bool("enabled").Default(true),
		field.Bool("notify_boss").Default(true),
		field.Bool("wechat").Default(true),
		field.Bool("app").Default(false),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (StoreNotifySetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("store_id", "event").Unique(),
	}
}
