package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type OAuthIdentity struct {
	ent.Schema
}

func (OAuthIdentity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth_identities"},
	}
}

func (OAuthIdentity) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("provider").MaxLen(32).NotEmpty(),
		field.String("provider_uid").MaxLen(128).NotEmpty(),
		field.String("union_id").MaxLen(128).Optional().Nillable(),
		field.Int64("user_id").Default(0),
		field.String("nickname").MaxLen(64).Optional().Nillable(),
		field.String("avatar_url").MaxLen(512).Optional().Nillable(),
		field.JSON("extra", map[string]any{}).Optional(),
		field.Time("authorized_at").Default(time.Now),
		field.Time("bound_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (OAuthIdentity) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider", "provider_uid").Unique(),
		index.Fields("user_id"),
		index.Fields("provider", "union_id"),
	}
}
