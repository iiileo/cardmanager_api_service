package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type SmsCode struct {
	ent.Schema
}

func (SmsCode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "sms_codes"},
	}
}

func (SmsCode) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("phone").MaxLen(20).NotEmpty(),
		field.String("scene").MaxLen(16).NotEmpty(),
		field.String("code_hash").MaxLen(64).NotEmpty(),
		field.Time("expires_at"),
		field.Time("used_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (SmsCode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("phone", "scene"),
		index.Fields("expires_at"),
	}
}
