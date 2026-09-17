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

func (r *dashboardRepository) StatsOverview(
	ctx context.Context,
	in domaindash.StatsOverviewQuery,
) (*domaindash.StatsOverview, error) {
	out := &domaindash.StatsOverview{}

	cur, err := r.overviewPeriod(ctx, in.StoreID, in.CurFrom, in.CurTo)
	if err != nil {
		return nil, err
	}
	prev, err := r.overviewPeriod(ctx, in.StoreID, in.PrevFrom, in.PrevTo)
	if err != nil {
		return nil, err
	}
	out.Cur = cur
	out.Prev = prev

	if err := r.overviewMemberActivity(
		ctx, in.StoreID, in.CurFrom, in.CurTo, &out.RepeatCustomers, &out.ActiveMembers,
	); err != nil {
		return nil, err
	}

	if err := r.client.DB().QueryRowContext(ctx, `
SELECT COUNT(DISTINCT member_id)::int
FROM member_cards
WHERE store_id = $1 AND status = $2`, in.StoreID, "active").
		Scan(&out.TotalCardHolders); err != nil {
		return nil, fmt.Errorf("overview card holders: %w", err)
	}

	daily, err := r.overviewDailyRecharge(ctx, in.StoreID, in.CurFrom, in.CurTo)
	if err != nil {
		return nil, err
	}
	out.DailyRecharge = daily

	mix, err := r.overviewCardTypeMix(ctx, in.StoreID)
	if err != nil {
		return nil, err
	}
	out.CardTypeMix = mix

	rank, err := r.overviewRechargeRank(ctx, in.StoreID, in.CurFrom, in.CurTo, 10)
	if err != nil {
		return nil, err
	}
	out.RechargeRank = rank

	return out, nil
}

func (r *dashboardRepository) overviewPeriod(
	ctx context.Context,
	storeID int64,
	from, to time.Time,
) (domaindash.StatsOverviewPeriod, error) {
	var p domaindash.StatsOverviewPeriod
	consumeTypes := []string{
		domainledger.TypeConsumeValue,
		domainledger.TypeConsumeCount,
		domainledger.TypeConsumePack,
	}
	sql := `
SELECT
  COALESCE(SUM(CASE WHEN type = $4 THEN amount ELSE 0 END), 0)::int,
  COALESCE(SUM(CASE WHEN type = ANY($5) THEN ABS(COALESCE(amount, 0)) ELSE 0 END), 0)::int,
  COUNT(*) FILTER (WHERE type = $6)::int
FROM ledger_entries
WHERE store_id = $1
  AND created_at >= $2
  AND created_at <= $3`
	if err := r.client.DB().QueryRowContext(
		ctx, sql, storeID, from, to,
		domainledger.TypeRecharge,
		pq.Array(consumeTypes),
		domainledger.TypeOpen,
	).Scan(&p.RechargeAmount, &p.ConsumeAmount, &p.NewOpens); err != nil {
		return p, fmt.Errorf("overview period: %w", err)
	}
	return p, nil
}

func (r *dashboardRepository) overviewMemberActivity(
	ctx context.Context,
	storeID int64,
	from, to time.Time,
	repeat, active *int,
) error {
	txnTypes := []string{
		domainledger.TypeRecharge,
		domainledger.TypeConsumeValue,
		domainledger.TypeConsumeCount,
		domainledger.TypeConsumePack,
	}
	sql := `
SELECT
  COUNT(*) FILTER (WHERE cnt >= 2)::int,
  COUNT(*)::int
FROM (
  SELECT member_id, COUNT(*)::int AS cnt
  FROM ledger_entries
  WHERE store_id = $1
    AND created_at >= $2
    AND created_at <= $3
    AND type = ANY($4)
  GROUP BY member_id
) t`
	return r.client.DB().QueryRowContext(
		ctx, sql, storeID, from, to, pq.Array(txnTypes),
	).Scan(repeat, active)
}

func (r *dashboardRepository) overviewDailyRecharge(
	ctx context.Context,
	storeID int64,
	from, to time.Time,
) ([]domaindash.DailyRechargePoint, error) {
	sql := `
SELECT created_at, amount
FROM ledger_entries
WHERE store_id = $1
  AND type = $2
  AND created_at >= $3
  AND created_at <= $4
ORDER BY created_at`
	rows, err := r.client.DB().QueryContext(
		ctx, sql, storeID, domainledger.TypeRecharge, from, to,
	)
	if err != nil {
		return nil, fmt.Errorf("overview daily recharge: %w", err)
	}
	defer rows.Close()

	byDay := make(map[string]int)
	for rows.Next() {
		var createdAt time.Time
		var amt int
		if err := rows.Scan(&createdAt, &amt); err != nil {
			return nil, err
		}
		local := createdAt.In(time.Local)
		key := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.Local).
			Format("2006-01-02")
		byDay[key] += amt
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	out := make([]domaindash.DailyRechargePoint, 0)
	start := time.Date(from.In(time.Local).Year(), from.In(time.Local).Month(), from.In(time.Local).Day(), 0, 0, 0, 0, time.Local)
	endDay := time.Date(to.In(time.Local).Year(), to.In(time.Local).Month(), to.In(time.Local).Day(), 0, 0, 0, 0, time.Local)
	for d := start; !d.After(endDay); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		out = append(out, domaindash.DailyRechargePoint{
			Date:   d,
			Amount: byDay[key],
		})
	}
	return out, nil
}

func (r *dashboardRepository) overviewCardTypeMix(
	ctx context.Context,
	storeID int64,
) ([]domaindash.CardTypeMixRow, error) {
	sql := `
SELECT type, COUNT(*)::int
FROM member_cards
WHERE store_id = $1 AND status = $2
GROUP BY type
ORDER BY type`
	rows, err := r.client.DB().QueryContext(ctx, sql, storeID, "active")
	if err != nil {
		return nil, fmt.Errorf("overview card mix: %w", err)
	}
	defer rows.Close()

	out := make([]domaindash.CardTypeMixRow, 0)
	for rows.Next() {
		var row domaindash.CardTypeMixRow
		if err := rows.Scan(&row.CardType, &row.Count); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *dashboardRepository) overviewRechargeRank(
	ctx context.Context,
	storeID int64,
	from, to time.Time,
	limit int,
) ([]domaindash.RechargeRankRow, error) {
	if limit <= 0 {
		limit = 10
	}
	sql := `
SELECT le.member_id, m.name, COALESCE(SUM(le.amount), 0)::int
FROM ledger_entries le
INNER JOIN members m ON m.id = le.member_id
WHERE le.store_id = $1
  AND le.type = $2
  AND le.created_at >= $3
  AND le.created_at <= $4
GROUP BY le.member_id, m.name
ORDER BY SUM(le.amount) DESC, le.member_id
LIMIT $5`
	rows, err := r.client.DB().QueryContext(
		ctx, sql, storeID, domainledger.TypeRecharge, from, to, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("overview recharge rank: %w", err)
	}
	defer rows.Close()

	out := make([]domaindash.RechargeRankRow, 0)
	for rows.Next() {
		var row domaindash.RechargeRankRow
		if err := rows.Scan(&row.MemberID, &row.Name, &row.Amount); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}
