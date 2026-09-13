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

// CardProduct 门店卡种。
type CardProduct struct {
	ent.Schema
}

func (CardProduct) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "card_products"},
	}
}

func (CardProduct) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.Int64("store_id"),
		field.String("type").MaxLen(16).NotEmpty(), // value / count / pack
		field.String("name").MaxLen(64).NotEmpty(),
		field.Int("price").Default(0),
		field.Int("times").Optional().Nillable(),
		field.Int("valid_months").Optional().Nillable(),
		field.Int8("status").Default(1), // 1 上架 0 停用
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (CardProduct) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("items", CardProductItem.Type),
	}
}

func (CardProduct) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("store_id", "type", "status"),
		index.Fields("store_id"),
	}
}
