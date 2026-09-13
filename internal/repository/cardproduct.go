package repository

import (
	"context"

	"card_manager/api_service/ent"
	entitem "card_manager/api_service/ent/cardproductitem"
	entproduct "card_manager/api_service/ent/cardproduct"
	domainproduct "card_manager/api_service/internal/domain/cardproduct"
	"card_manager/api_service/internal/postgres"
)

type cardProductRepository struct {
	client *postgres.Client
}

func NewCardProductRepository(client *postgres.Client) domainproduct.Repository {
	return &cardProductRepository{client: client}
}

func (r *cardProductRepository) ListByStore(ctx context.Context, storeID int64, typ string) ([]*domainproduct.Product, error) {
	q := r.client.Ent().CardProduct.Query().
		Where(
			entproduct.StoreIDEQ(storeID),
			entproduct.StatusEQ(domainproduct.StatusActive),
		).
		WithItems(func(iq *ent.CardProductItemQuery) {
			iq.Order(ent.Asc(entitem.FieldSort), ent.Asc(entitem.FieldID))
		}).
		Order(ent.Asc(entproduct.FieldType), ent.Desc(entproduct.FieldID))
	if typ != "" {
		q = q.Where(entproduct.TypeEQ(typ))
	}
	list, err := q.All(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domainproduct.Product, 0, len(list))
	for _, p := range list {
		out = append(out, domainproduct.FromEnt(p, p.Edges.Items))
	}
	return out, nil
}

func (r *cardProductRepository) GetByID(ctx context.Context, id int64) (*domainproduct.Product, error) {
	p, err := r.client.Ent().CardProduct.Query().
		Where(entproduct.IDEQ(id)).
		WithItems(func(iq *ent.CardProductItemQuery) {
			iq.Order(ent.Asc(entitem.FieldSort), ent.Asc(entitem.FieldID))
		}).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return domainproduct.FromEnt(p, p.Edges.Items), nil
}

func (r *cardProductRepository) CountActiveByStoreType(ctx context.Context, storeID int64, typ string) (int, error) {
	return r.client.Ent().CardProduct.Query().
		Where(
			entproduct.StoreIDEQ(storeID),
			entproduct.TypeEQ(typ),
			entproduct.StatusEQ(domainproduct.StatusActive),
		).
		Count(ctx)
}

func (r *cardProductRepository) Create(ctx context.Context, in domainproduct.CreateInput) (*domainproduct.Product, error) {
	var out *domainproduct.Product
	err := r.client.WithTx(ctx, func(tx *ent.Tx) error {
		b := tx.CardProduct.Create().
			SetStoreID(in.StoreID).
			SetType(in.Type).
			SetName(in.Name).
			SetPrice(in.Price).
			SetStatus(domainproduct.StatusActive)
		if in.Times != nil {
			b.SetTimes(*in.Times)
		}
		if in.ValidMonths != nil {
			b.SetValidMonths(*in.ValidMonths)
		}
		p, err := b.Save(ctx)
		if err != nil {
			return err
		}
		createdItems := make([]*ent.CardProductItem, 0, len(in.Items))
		for i, it := range in.Items {
			sort := it.Sort
			if sort == 0 {
				sort = i + 1
			}
			row, err := tx.CardProductItem.Create().
				SetProductID(p.ID).
				SetName(it.Name).
				SetTimes(it.Times).
				SetSort(sort).
				Save(ctx)
			if err != nil {
				return err
			}
			createdItems = append(createdItems, row)
		}
		out = domainproduct.FromEnt(p, createdItems)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *cardProductRepository) Delete(ctx context.Context, id int64) error {
	return r.client.WithTx(ctx, func(tx *ent.Tx) error {
		if _, err := tx.CardProductItem.Delete().Where(entitem.ProductIDEQ(id)).Exec(ctx); err != nil {
			return err
		}
		return tx.CardProduct.DeleteOneID(id).Exec(ctx)
	})
}
