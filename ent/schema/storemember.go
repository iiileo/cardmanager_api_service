package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type StoreMember struct {
	ent.Schema
}

func (StoreMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "store_members"},
	}
}

func (StoreMember) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("store_id"),
		field.Int64("user_id"),
		field.String("role").MaxLen(16).NotEmpty(),
		field.String("status").MaxLen(16).NotEmpty(),
		field.String("display_name").MaxLen(32).Optional().Nillable(),
		field.Time("joined_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (StoreMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("store_id", "user_id").Unique(),
		index.Fields("store_id", "status"),
		index.Fields("user_id"),
	}
}
