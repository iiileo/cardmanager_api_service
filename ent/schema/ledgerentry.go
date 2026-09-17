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

// LedgerEntry 流水账（开卡 / 充值 / 扣除记录）。
type LedgerEntry struct {
	ent.Schema
}

func (LedgerEntry) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ledger_entries"},
	}
}

func (LedgerEntry) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("store_id"),
		field.Int64("member_id"),
		field.Int64("card_id"),
		field.String("type").MaxLen(32).NotEmpty(), // open / recharge / consume_*
		// card_type 快照：value / count / pack，便于列表筛选与统计（不随卡变更）
		field.String("card_type").MaxLen(16).Default(""),
		field.Int("amount").Optional().Nillable(),
		field.Int("times").Optional().Nillable(),
		field.String("item_name").MaxLen(64).Optional().Nillable(),
		field.Int("balance_after").Optional().Nillable(),
		field.Int("times_after").Optional().Nillable(),
		field.String("remark").MaxLen(128).Optional().Nillable(),
		field.Int64("operator_id"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (LedgerEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("member", Member.Type).
			Ref("ledgers").
			Unique().
			Required().
			Field("member_id"),
		edge.From("card", MemberCard.Type).
			Ref("ledgers").
			Unique().
			Required().
			Field("card_id"),
		edge.From("operator", User.Type).
			Ref("operated_ledgers").
			Unique().
			Required().
			Field("operator_id"),
		edge.To("items", LedgerEntryItem.Type),
	}
}

func (LedgerEntry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("store_id", "created_at"),
		index.Fields("member_id", "created_at"),
		index.Fields("operator_id", "created_at"),
		index.Fields("card_id", "created_at"),
		index.Fields("store_id", "type"),
		index.Fields("store_id", "card_type", "created_at"),
		index.Fields("store_id", "type", "created_at"),
	}
}
