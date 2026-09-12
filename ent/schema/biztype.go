package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// BizType 门店业态字典。
type BizType struct {
	ent.Schema
}

func (BizType) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "biz_types"},
	}
}

func (BizType) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("code").MaxLen(16).Unique().NotEmpty(),
		field.String("name").MaxLen(32).NotEmpty(),
		field.Int("sort").Default(0),
		field.Int8("status").Default(1), // 1=启用 0=停用
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (BizType) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("status", "sort"),
	}
}
