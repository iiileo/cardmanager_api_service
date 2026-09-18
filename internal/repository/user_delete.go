package repository

import (
	"context"
	"fmt"

	"card_manager/api_service/ent"
	"card_manager/api_service/ent/carditembalance"
	"card_manager/api_service/ent/cardproduct"
	"card_manager/api_service/ent/cardproductitem"
	"card_manager/api_service/ent/ledgerentry"
	"card_manager/api_service/ent/ledgerentryitem"
	"card_manager/api_service/ent/member"
	"card_manager/api_service/ent/membercard"
	"card_manager/api_service/ent/oauthidentity"
	"card_manager/api_service/ent/refreshtoken"
	"card_manager/api_service/ent/smscode"
	"card_manager/api_service/ent/store"
	"card_manager/api_service/ent/storemember"
	"card_manager/api_service/ent/storenotifysetting"
)

// DeleteAccount 注销账号：删除名下门店全量业务数据、员工身份、凭证与用户本身。
// 在他人门店留下的流水操作人会改挂到该店老板，避免破坏对方门店数据。
func (r *userRepository) DeleteAccount(ctx context.Context, userID int64) error {
	return r.client.WithTx(ctx, func(tx *ent.Tx) error {
		u, err := tx.User.Get(ctx, userID)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return err
		}

		ownedIDs, err := tx.Store.Query().
			Where(store.OwnerUserIDEQ(userID)).
			IDs(ctx)
		if err != nil {
			return fmt.Errorf("list owned stores: %w", err)
		}
		for _, storeID := range ownedIDs {
			if err := deleteStoreData(ctx, tx, storeID); err != nil {
				return fmt.Errorf("delete store %d: %w", storeID, err)
			}
		}

		// 仍引用本用户为操作人的流水（他人门店）→ 改挂该店老板
		leftover, err := tx.LedgerEntry.Query().
			Where(ledgerentry.OperatorIDEQ(userID)).
			All(ctx)
		if err != nil {
			return fmt.Errorf("list leftover ledgers: %w", err)
		}
		for _, e := range leftover {
			st, err := tx.Store.Get(ctx, e.StoreID)
			if err != nil {
				return fmt.Errorf("load store %d for ledger reassign: %w", e.StoreID, err)
			}
			if st.OwnerUserID == userID {
				return fmt.Errorf("unexpected leftover ledger on owned store %d", e.StoreID)
			}
			if _, err := tx.LedgerEntry.UpdateOneID(e.ID).
				SetOperatorID(st.OwnerUserID).
				Save(ctx); err != nil {
				return fmt.Errorf("reassign ledger %d operator: %w", e.ID, err)
			}
		}

		if _, err := tx.StoreMember.Delete().
			Where(storemember.UserIDEQ(userID)).
			Exec(ctx); err != nil {
			return fmt.Errorf("delete store members: %w", err)
		}
		if _, err := tx.RefreshToken.Delete().
			Where(refreshtoken.UserIDEQ(userID)).
			Exec(ctx); err != nil {
			return fmt.Errorf("delete refresh tokens: %w", err)
		}
		if _, err := tx.OAuthIdentity.Delete().
			Where(oauthidentity.UserIDEQ(userID)).
			Exec(ctx); err != nil {
			return fmt.Errorf("delete oauth identities: %w", err)
		}
		if _, err := tx.SmsCode.Delete().
			Where(smscode.PhoneEQ(u.Phone)).
			Exec(ctx); err != nil {
			return fmt.Errorf("delete sms codes: %w", err)
		}
		if err := tx.User.DeleteOneID(userID).Exec(ctx); err != nil {
			if !ent.IsNotFound(err) {
				return fmt.Errorf("delete user: %w", err)
			}
		}
		return nil
	})
}

func deleteStoreData(ctx context.Context, tx *ent.Tx, storeID int64) error {
	ledgerIDs, err := tx.LedgerEntry.Query().
		Where(ledgerentry.StoreIDEQ(storeID)).
		IDs(ctx)
	if err != nil {
		return err
	}
	if len(ledgerIDs) > 0 {
		if _, err := tx.LedgerEntryItem.Delete().
			Where(ledgerentryitem.LedgerIDIn(ledgerIDs...)).
			Exec(ctx); err != nil {
			return err
		}
	}
	if _, err := tx.LedgerEntry.Delete().
		Where(ledgerentry.StoreIDEQ(storeID)).
		Exec(ctx); err != nil {
		return err
	}

	cardIDs, err := tx.MemberCard.Query().
		Where(membercard.StoreIDEQ(storeID)).
		IDs(ctx)
	if err != nil {
		return err
	}
	if len(cardIDs) > 0 {
		if _, err := tx.CardItemBalance.Delete().
			Where(carditembalance.CardIDIn(cardIDs...)).
			Exec(ctx); err != nil {
			return err
		}
	}
	if _, err := tx.MemberCard.Delete().
		Where(membercard.StoreIDEQ(storeID)).
		Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.Member.Delete().
		Where(member.StoreIDEQ(storeID)).
		Exec(ctx); err != nil {
		return err
	}

	productIDs, err := tx.CardProduct.Query().
		Where(cardproduct.StoreIDEQ(storeID)).
		IDs(ctx)
	if err != nil {
		return err
	}
	if len(productIDs) > 0 {
		if _, err := tx.CardProductItem.Delete().
			Where(cardproductitem.ProductIDIn(productIDs...)).
			Exec(ctx); err != nil {
			return err
		}
	}
	if _, err := tx.CardProduct.Delete().
		Where(cardproduct.StoreIDEQ(storeID)).
		Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.StoreNotifySetting.Delete().
		Where(storenotifysetting.StoreIDEQ(storeID)).
		Exec(ctx); err != nil {
		return err
	}
	if _, err := tx.StoreMember.Delete().
		Where(storemember.StoreIDEQ(storeID)).
		Exec(ctx); err != nil {
		return err
	}
	if err := tx.Store.DeleteOneID(storeID).Exec(ctx); err != nil {
		return err
	}
	return nil
}
