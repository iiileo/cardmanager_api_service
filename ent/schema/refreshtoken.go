package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type RefreshToken struct {
	ent.Schema
}

func (RefreshToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "refresh_tokens"},
	}
}

func (RefreshToken) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("user_id"),
		field.String("token_hash").MaxLen(64).Unique().NotEmpty(),
		field.Time("expires_at"),
		field.Time("revoked_at").Optional().Nillable(),
		field.String("device_id").MaxLen(64).Optional().Nillable(),
		field.String("user_agent").MaxLen(256).Optional().Nillable(),
		field.String("ip").MaxLen(64).Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (RefreshToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token_hash").Unique(),
		index.Fields("user_id"),
		index.Fields("expires_at"),
	}
}
