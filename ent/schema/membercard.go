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

// MemberCard 会员持卡实例。
type MemberCard struct {
	ent.Schema
}

func (MemberCard) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "member_cards"},
	}
}

func (MemberCard) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("store_id"),
		field.Int64("member_id"),
		field.Int64("product_id"),
		field.String("type").MaxLen(16).NotEmpty(), // value / count / pack
		field.String("name_snapshot").MaxLen(64).NotEmpty(),
		field.Int("balance").Default(0),
		field.Int("remain_times").Optional().Nillable(),
		field.Time("valid_from").Optional().Nillable(),
		field.Time("valid_to").Optional().Nillable(),
		field.String("status").MaxLen(16).Default("active"), // active / expired / exhausted
		field.Int64("opened_by"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (MemberCard) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("member", Member.Type).
			Ref("cards").
			Unique().
			Required().
			Field("member_id"),
		edge.To("item_balances", CardItemBalance.Type),
		edge.To("ledgers", LedgerEntry.Type),
	}
}

func (MemberCard) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("member_id"),
		index.Fields("store_id", "type"),
		index.Fields("store_id", "status"),
	}
}
