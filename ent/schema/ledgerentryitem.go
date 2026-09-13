package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// LedgerEntryItem 一次扣除中的套餐项目明细（支持一次扣多项）。
type LedgerEntryItem struct {
	ent.Schema
}

func (LedgerEntryItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ledger_entry_items"},
	}
}

func (LedgerEntryItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("ledger_id"),
		field.Int64("item_balance_id"),
		field.Int64("product_item_id"),
		field.String("name_snapshot").MaxLen(32).NotEmpty(),
		field.Int("times"),       // 扣减次数（正数）
		field.Int("times_after"), // 扣后剩余
	}
}

func (LedgerEntryItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("ledger", LedgerEntry.Type).
			Ref("items").
			Unique().
			Required().
			Field("ledger_id"),
	}
}

func (LedgerEntryItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ledger_id"),
		// 套餐项目消耗统计
		index.Fields("product_item_id"),
		index.Fields("name_snapshot"),
	}
}
