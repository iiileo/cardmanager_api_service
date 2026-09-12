package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Store struct {
	ent.Schema
}

func (Store) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "stores"},
	}
}

func (Store) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("name").MaxLen(64).NotEmpty(),
		field.String("city").MaxLen(32).NotEmpty(),
		field.String("address").MaxLen(128).Optional().Nillable(),
		field.String("open_time").MaxLen(5).NotEmpty(),
		field.String("close_time").MaxLen(5).NotEmpty(),
		field.String("biz_type").MaxLen(16).Optional().Nillable(),
		field.String("invite_code").MaxLen(6).Unique().NotEmpty(),
		field.Int64("owner_user_id"),
		field.Int8("status").Default(1),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Store) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("invite_code").Unique(),
		index.Fields("owner_user_id"),
	}
}
