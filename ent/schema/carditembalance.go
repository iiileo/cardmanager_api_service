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

// CardItemBalance 套餐卡项目剩余次数。
type CardItemBalance struct {
	ent.Schema
}

func (CardItemBalance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "card_item_balances"},
	}
}

func (CardItemBalance) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("card_id"),
		field.Int64("product_item_id"),
		field.String("name_snapshot").MaxLen(32).NotEmpty(),
		field.Int("remain_times"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (CardItemBalance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("card", MemberCard.Type).
			Ref("item_balances").
			Unique().
			Required().
			Field("card_id"),
	}
}

func (CardItemBalance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("card_id", "product_item_id").Unique(),
		index.Fields("card_id"),
	}
}
