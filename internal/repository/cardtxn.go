package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"card_manager/api_service/ent"
	entbalance "card_manager/api_service/ent/carditembalance"
	entcard "card_manager/api_service/ent/membercard"
	domainledger "card_manager/api_service/internal/domain/ledger"
	domaincard "card_manager/api_service/internal/domain/membercard"
	"card_manager/api_service/internal/postgres"
)

// CardTxnRepository 卡交易（充值/扣除）与流水同事务写入。
type CardTxnRepository interface {
	Recharge(ctx context.Context, cardID int64, amount int, operatorID int64, remark *string) (*domaincard.Card, *domainledger.Entry, error)
	ConsumeValue(ctx context.Context, cardID int64, amount int, operatorID int64, remark *string) (*domaincard.Card, *domainledger.Entry, error)
	ConsumeCount(ctx context.Context, cardID int64, times int, operatorID int64, remark *string) (*domaincard.Card, *domainledger.Entry, error)
	ConsumePack(ctx context.Context, cardID int64, items []domaincard.PackDeductItem, operatorID int64, remark *string) (*domaincard.Card, *domainledger.Entry, error)
	OpenCard(ctx context.Context, cardIn domaincard.CreateInput, ledgerIn domainledger.CreateInput) (*domaincard.Card, *domainledger.Entry, error)
}

type cardTxnRepository struct {
	client *postgres.Client
}

func NewCardTxnRepository(client *postgres.Client) CardTxnRepository {
	return &cardTxnRepository{client: client}
}

var ErrInsufficient = errors.New("insufficient")

func IsInsufficient(err error) bool {
	return errors.Is(err, ErrInsufficient)
}

func (r *cardTxnRepository) loadCard(ctx context.Context, tx *ent.Tx, id int64) (*ent.MemberCard, error) {
	return tx.MemberCard.Query().
		Where(entcard.IDEQ(id)).
		WithItemBalances(func(iq *ent.CardItemBalanceQuery) {
			iq.Order(ent.Asc(entbalance.FieldID))
		}).
		Only(ctx)
}

func (r *cardTxnRepository) createLedger(ctx context.Context, tx *ent.Tx, in domainledger.CreateInput) (*ent.LedgerEntry, []*ent.LedgerEntryItem, error) {
	b := tx.LedgerEntry.Create().
		SetStoreID(in.StoreID).
		SetMemberID(in.MemberID).
		SetCardID(in.CardID).
		SetType(in.Type).
		SetCardType(in.CardType).
		SetOperatorID(in.OperatorID)
	if in.Amount != nil {
		b.SetAmount(*in.Amount)
	}
	if in.Times != nil {
		b.SetTimes(*in.Times)
	}
	if in.ItemName != nil {
		b.SetItemName(*in.ItemName)
	}
	if in.BalanceAfter != nil {
		b.SetBalanceAfter(*in.BalanceAfter)
	}
	if in.TimesAfter != nil {
		b.SetTimesAfter(*in.TimesAfter)
	}
	if in.Remark != nil {
		b.SetRemark(*in.Remark)
	}
	e, err := b.Save(ctx)
	if err != nil {
		return nil, nil, err
	}
	created := make([]*ent.LedgerEntryItem, 0, len(in.Items))
	for _, it := range in.Items {
		row, err := tx.LedgerEntryItem.Create().
			SetLedgerID(e.ID).
			SetItemBalanceID(it.ItemBalanceID).
			SetProductItemID(it.ProductItemID).
			SetNameSnapshot(it.NameSnapshot).
			SetTimes(it.Times).
			SetTimesAfter(it.TimesAfter).
			Save(ctx)
		if err != nil {
			return nil, nil, err
		}
		created = append(created, row)
	}
	return e, created, nil
}

func (r *cardTxnRepository) Recharge(ctx context.Context, cardID int64, amount int, operatorID int64, remark *string) (*domaincard.Card, *domainledger.Entry, error) {
	var card *domaincard.Card
	var entry *domainledger.Entry
	err := r.client.WithTx(ctx, func(tx *ent.Tx) error {
		n, err := tx.MemberCard.Update().
			Where(entcard.IDEQ(cardID), entcard.TypeEQ(domaincard.TypeValue)).
			AddBalance(amount).
			Save(ctx)
		if err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("card not found")
		}
		c, err := r.loadCard(ctx, tx, cardID)
		if err != nil {
			return err
		}
		bal := c.Balance
		e, items, err := r.createLedger(ctx, tx, domainledger.CreateInput{
			StoreID:      c.StoreID,
			MemberID:     c.MemberID,
			CardID:       c.ID,
			Type:         domainledger.TypeRecharge,
			CardType:     domaincard.TypeValue,
			Amount:       &amount,
			BalanceAfter: &bal,
			Remark:       remark,
			OperatorID:   operatorID,
		})
		if err != nil {
			return err
		}
		card = domaincard.FromEnt(c, nil)
		entry = domainledger.FromEnt(e, items)
		return nil
	})
	return card, entry, err
}

func (r *cardTxnRepository) ConsumeValue(ctx context.Context, cardID int64, amount int, operatorID int64, remark *string) (*domaincard.Card, *domainledger.Entry, error) {
	var card *domaincard.Card
	var entry *domainledger.Entry
	err := r.client.WithTx(ctx, func(tx *ent.Tx) error {
		n, err := tx.MemberCard.Update().
			Where(
				entcard.IDEQ(cardID),
				entcard.TypeEQ(domaincard.TypeValue),
				entcard.BalanceGTE(amount),
			).
			AddBalance(-amount).
			Save(ctx)
		if err != nil {
			return err
		}
		if n == 0 {
			return ErrInsufficient
		}
		c, err := r.loadCard(ctx, tx, cardID)
		if err != nil {
			return err
		}
		amt := -amount
		bal := c.Balance
		e, items, err := r.createLedger(ctx, tx, domainledger.CreateInput{
			StoreID:      c.StoreID,
			MemberID:     c.MemberID,
			CardID:       c.ID,
			Type:         domainledger.TypeConsumeValue,
			CardType:     domaincard.TypeValue,
			Amount:       &amt,
			BalanceAfter: &bal,
			Remark:       remark,
			OperatorID:   operatorID,
		})
		if err != nil {
			return err
		}
		card = domaincard.FromEnt(c, nil)
		entry = domainledger.FromEnt(e, items)
		return nil
	})
	return card, entry, err
}

func (r *cardTxnRepository) ConsumeCount(ctx context.Context, cardID int64, times int, operatorID int64, remark *string) (*domaincard.Card, *domainledger.Entry, error) {
	var card *domaincard.Card
	var entry *domainledger.Entry
	err := r.client.WithTx(ctx, func(tx *ent.Tx) error {
		n, err := tx.MemberCard.Update().
			Where(
				entcard.IDEQ(cardID),
				entcard.TypeEQ(domaincard.TypeCount),
				entcard.RemainTimesGTE(times),
			).
			AddRemainTimes(-times).
			Save(ctx)
		if err != nil {
			return err
		}
		if n == 0 {
			return ErrInsufficient
		}
		c, err := r.loadCard(ctx, tx, cardID)
		if err != nil {
			return err
		}
		neg := -times
		after := 0
		if c.RemainTimes != nil {
			after = *c.RemainTimes
		}
		e, items, err := r.createLedger(ctx, tx, domainledger.CreateInput{
			StoreID:    c.StoreID,
			MemberID:   c.MemberID,
			CardID:     c.ID,
			Type:       domainledger.TypeConsumeCount,
			CardType:   domaincard.TypeCount,
			Times:      &neg,
			TimesAfter: &after,
			Remark:     remark,
			OperatorID: operatorID,
		})
		if err != nil {
			return err
		}
		card = domaincard.FromEnt(c, nil)
		entry = domainledger.FromEnt(e, items)
		return nil
	})
	return card, entry, err
}

func (r *cardTxnRepository) ConsumePack(ctx context.Context, cardID int64, items []domaincard.PackDeductItem, operatorID int64, remark *string) (*domaincard.Card, *domainledger.Entry, error) {
	var card *domaincard.Card
	var entry *domainledger.Entry
	err := r.client.WithTx(ctx, func(tx *ent.Tx) error {
		c, err := r.loadCard(ctx, tx, cardID)
		if err != nil {
			if ent.IsNotFound(err) {
				return fmt.Errorf("card not found")
			}
			return err
		}
		if c.Type != domaincard.TypePack {
			return fmt.Errorf("not pack card")
		}
		byID := make(map[int64]*ent.CardItemBalance, len(c.Edges.ItemBalances))
		for _, b := range c.Edges.ItemBalances {
			byID[b.ID] = b
		}

		ledgerItems := make([]domainledger.ItemInput, 0, len(items))
		names := make([]string, 0, len(items))
		totalTimes := 0
		for _, it := range items {
			bal, ok := byID[it.ItemBalanceID]
			if !ok {
				return fmt.Errorf("item not found")
			}
			if bal.RemainTimes < it.Times {
				return ErrInsufficient
			}
			n, err := tx.CardItemBalance.Update().
				Where(
					entbalance.IDEQ(it.ItemBalanceID),
					entbalance.CardIDEQ(cardID),
					entbalance.RemainTimesGTE(it.Times),
				).
				AddRemainTimes(-it.Times).
				Save(ctx)
			if err != nil {
				return err
			}
			if n == 0 {
				return ErrInsufficient
			}
			updated, err := tx.CardItemBalance.Get(ctx, it.ItemBalanceID)
			if err != nil {
				return err
			}
			totalTimes += it.Times
			names = append(names, updated.NameSnapshot)
			ledgerItems = append(ledgerItems, domainledger.ItemInput{
				ItemBalanceID: updated.ID,
				ProductItemID: updated.ProductItemID,
				NameSnapshot:  updated.NameSnapshot,
				Times:         it.Times,
				TimesAfter:    updated.RemainTimes,
			})
		}

		neg := -totalTimes
		itemName := strings.Join(names, "、")
		if len([]rune(itemName)) > 64 {
			runes := []rune(itemName)
			itemName = string(runes[:61]) + "…"
		}
		e, created, err := r.createLedger(ctx, tx, domainledger.CreateInput{
			StoreID:    c.StoreID,
			MemberID:   c.MemberID,
			CardID:     c.ID,
			Type:       domainledger.TypeConsumePack,
			CardType:   domaincard.TypePack,
			Times:      &neg,
			ItemName:   &itemName,
			Remark:     remark,
			OperatorID: operatorID,
			Items:      ledgerItems,
		})
		if err != nil {
			return err
		}
		c2, err := r.loadCard(ctx, tx, cardID)
		if err != nil {
			return err
		}
		card = domaincard.FromEnt(c2, nil)
		entry = domainledger.FromEnt(e, created)
		return nil
	})
	return card, entry, err
}

func (r *cardTxnRepository) OpenCard(ctx context.Context, cardIn domaincard.CreateInput, ledgerIn domainledger.CreateInput) (*domaincard.Card, *domainledger.Entry, error) {
	var card *domaincard.Card
	var entry *domainledger.Entry
	err := r.client.WithTx(ctx, func(tx *ent.Tx) error {
		b := tx.MemberCard.Create().
			SetStoreID(cardIn.StoreID).
			SetMemberID(cardIn.MemberID).
			SetProductID(cardIn.ProductID).
			SetType(cardIn.Type).
			SetNameSnapshot(cardIn.NameSnapshot).
			SetBalance(cardIn.Balance).
			SetStatus(domaincard.StatusActive).
			SetOpenedBy(cardIn.OpenedBy)
		if cardIn.RemainTimes != nil {
			b.SetRemainTimes(*cardIn.RemainTimes)
		}
		if cardIn.ValidFrom != nil {
			b.SetValidFrom(*cardIn.ValidFrom)
		}
		if cardIn.ValidTo != nil {
			b.SetValidTo(*cardIn.ValidTo)
		}
		c, err := b.Save(ctx)
		if err != nil {
			return err
		}
		items := make([]*ent.CardItemBalance, 0, len(cardIn.Items))
		for _, it := range cardIn.Items {
			row, err := tx.CardItemBalance.Create().
				SetCardID(c.ID).
				SetProductItemID(it.ProductItemID).
				SetNameSnapshot(it.NameSnapshot).
				SetRemainTimes(it.RemainTimes).
				Save(ctx)
			if err != nil {
				return err
			}
			items = append(items, row)
		}
		ledgerIn.CardID = c.ID
		ledgerIn.MemberID = c.MemberID
		ledgerIn.StoreID = c.StoreID
		if ledgerIn.CardType == "" {
			ledgerIn.CardType = cardIn.Type
		}
		e, ledItems, err := r.createLedger(ctx, tx, ledgerIn)
		if err != nil {
			return err
		}
		card = domaincard.FromEnt(c, items)
		entry = domainledger.FromEnt(e, ledItems)
		return nil
	})
	return card, entry, err
}