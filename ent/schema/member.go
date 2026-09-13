package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Member 门店会员（顾客）。
type Member struct {
	ent.Schema
}

func (Member) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "members"},
	}
}

func (Member) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("store_id"),
		field.String("name").MaxLen(32).NotEmpty(),
		field.String("name_pinyin").MaxLen(128).Default(""),  // 全拼，如 liming
		field.String("name_initials").MaxLen(32).Default(""), // 首拼，如 lm
		field.String("phone").MaxLen(20).NotEmpty(),
		field.String("source").MaxLen(32).Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (Member) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("cards", MemberCard.Type),
		edge.To("ledgers", LedgerEntry.Type),
	}
}

func (Member) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("store_id", "phone").Unique(),
		index.Fields("store_id", "name"),
		index.Fields("store_id", "name_pinyin"),
		index.Fields("store_id", "name_initials"),
	}
}
