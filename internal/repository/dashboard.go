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

func (r *dashboardRepository) HomeStats(ctx context.Context, storeID int64, from, to time.Time) (*domaindash.HomeStats, error) {
	out := &domaindash.HomeStats{}

	// 今日流水：充值金额、储值消费金额、业务笔数（充值+消费）
	ledgerSQL := `
SELECT
  COALESCE(SUM(CASE WHEN type = $4 THEN amount ELSE 0 END), 0)::int,
  COALESCE(SUM(CASE WHEN type = $5 THEN ABS(amount) ELSE 0 END), 0)::int,
  COUNT(*) FILTER (WHERE type = ANY($6))::int
FROM ledger_entries
WHERE store_id = $1
  AND created_at >= $2
  AND created_at <= $3`

	txnTypes := []string{
		domainledger.TypeRecharge,
		domainledger.TypeConsumeValue,
		domainledger.TypeConsumeCount,
		domainledger.TypeConsumePack,
	}
	if err := r.client.DB().QueryRowContext(
		ctx, ledgerSQL,
		storeID, from, to,
		domainledger.TypeRecharge,
		domainledger.TypeConsumeValue,
		pq.Array(txnTypes),
	).Scan(&out.TodayRecharge, &out.TodayConsume, &out.TodayTxnCount); err != nil {
		return nil, fmt.Errorf("home stats ledger: %w", err)
	}

	memberSQL := `
SELECT COUNT(*)::int
FROM members
WHERE store_id = $1
  AND created_at >= $2
  AND created_at <= $3`
	if err := r.client.DB().QueryRowContext(ctx, memberSQL, storeID, from, to).
		Scan(&out.TodayNewMembers); err != nil {
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
