package repository

import (
	"context"
	"fmt"
	"time"

	domaindash "card_manager/api_service/internal/domain/dashboard"
	domainledger "card_manager/api_service/internal/domain/ledger"
	domaincard "card_manager/api_service/internal/domain/membercard"
	"card_manager/api_service/internal/postgres"
	"github.com/lib/pq"
)

type dashboardRepository struct {
	client *postgres.Client
}

func NewDashboardRepository(client *postgres.Client) domaindash.Repository {
	return &dashboardRepository{client: client}
}

func (r *dashboardRepository) HomeStats(
	ctx context.Context,
	storeID int64,
	todayFrom, todayTo, monthFrom, monthTo time.Time,
) (*domaindash.HomeStats, error) {
	out := &domaindash.HomeStats{}

	ledgerSQL := `
SELECT
  COALESCE(SUM(CASE WHEN type = $4 AND created_at >= $2 AND created_at <= $3 THEN amount ELSE 0 END), 0)::int,
  COALESCE(SUM(CASE WHEN type = $5 AND created_at >= $2 AND created_at <= $3 THEN ABS(amount) ELSE 0 END), 0)::int,
  COUNT(*) FILTER (WHERE type = ANY($6) AND created_at >= $2 AND created_at <= $3)::int,
  COALESCE(SUM(CASE WHEN type = $4 AND created_at >= $7 AND created_at <= $8 THEN amount ELSE 0 END), 0)::int,
  COALESCE(SUM(CASE WHEN type = $5 AND created_at >= $7 AND created_at <= $8 THEN ABS(amount) ELSE 0 END), 0)::int,
  COUNT(*) FILTER (WHERE type = ANY($6) AND created_at >= $7 AND created_at <= $8)::int
FROM ledger_entries
WHERE store_id = $1
  AND (
    (created_at >= $2 AND created_at <= $3)
    OR (created_at >= $7 AND created_at <= $8)
  )`

	txnTypes := []string{
		domainledger.TypeRecharge,
		domainledger.TypeConsumeValue,
		domainledger.TypeConsumeCount,
		domainledger.TypeConsumePack,
	}
	if err := r.client.DB().QueryRowContext(
		ctx, ledgerSQL,
		storeID, todayFrom, todayTo,
		domainledger.TypeRecharge,
		domainledger.TypeConsumeValue,
		pq.Array(txnTypes),
		monthFrom, monthTo,
	).Scan(
		&out.Today.Recharge, &out.Today.Consume, &out.Today.TxnCount,
		&out.Month.Recharge, &out.Month.Consume, &out.Month.TxnCount,
	); err != nil {
		return nil, fmt.Errorf("home stats ledger: %w", err)
	}

	memberSQL := `
SELECT
  COUNT(*) FILTER (WHERE created_at >= $2 AND created_at <= $3)::int,
  COUNT(*) FILTER (WHERE created_at >= $4 AND created_at <= $5)::int
FROM members
WHERE store_id = $1
  AND (
    (created_at >= $2 AND created_at <= $3)
    OR (created_at >= $4 AND created_at <= $5)
  )`
	if err := r.client.DB().QueryRowContext(
		ctx, memberSQL,
		storeID, todayFrom, todayTo, monthFrom, monthTo,
	).Scan(&out.Today.NewMembers, &out.Month.NewMembers); err != nil {
		return nil, fmt.Errorf("home stats members: %w", err)
	}

	balanceSQL := `
SELECT COALESCE(SUM(balance), 0)::int
FROM member_cards
WHERE store_id = $1
  AND type = $2`
	if err := r.client.DB().QueryRowContext(ctx, balanceSQL, storeID, domaincard.TypeValue).
		Scan(&out.StoreBalance); err != nil {
		return nil, fmt.Errorf("home stats balance: %w", err)
	}

	return out, nil
}
