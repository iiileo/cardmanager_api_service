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

// CardProductItem 套餐卡项目定义。
type CardProductItem struct {
	ent.Schema
}

func (CardProductItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "card_product_items"},
	}
}

func (CardProductItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("product_id"),
		field.String("name").MaxLen(32).NotEmpty(),
		field.Int("times"),
		field.Int("sort").Default(0),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (CardProductItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("product", CardProduct.Type).
			Ref("items").
			Unique().
			Required().
			Field("product_id"),
	}
}

func (CardProductItem) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id"),
	}
}
