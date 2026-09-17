package service

import (
	"context"

	domainledger "card_manager/api_service/internal/domain/ledger"
	domaincard "card_manager/api_service/internal/domain/membercard"
	domainmember "card_manager/api_service/internal/domain/member"
	domainuser "card_manager/api_service/internal/domain/user"
)

type ledgerRelated struct {
	members map[int64]*domainmember.Member
	cards   map[int64]*domaincard.Card
	users   map[int64]*domainuser.User
}

func (s *ledgerService) loadLedgerRelated(ctx context.Context, list []*domainledger.Entry) (*ledgerRelated, error) {
	memberIDs := make([]int64, 0)
	cardIDs := make([]int64, 0)
	userIDs := make([]int64, 0)
	seenM := make(map[int64]struct{})
	seenC := make(map[int64]struct{})
	seenU := make(map[int64]struct{})

	for _, e := range list {
		if e == nil {
			continue
		}
		if e.MemberID != 0 {
			if _, ok := seenM[e.MemberID]; !ok {
				seenM[e.MemberID] = struct{}{}
				memberIDs = append(memberIDs, e.MemberID)
			}
		}
		if e.CardID != 0 {
			if _, ok := seenC[e.CardID]; !ok {
				seenC[e.CardID] = struct{}{}
				cardIDs = append(cardIDs, e.CardID)
			}
		}
		if e.OperatorID != 0 {
			if _, ok := seenU[e.OperatorID]; !ok {
				seenU[e.OperatorID] = struct{}{}
				userIDs = append(userIDs, e.OperatorID)
			}
		}
	}

	members, err := s.members.ListByIDs(ctx, memberIDs)
	if err != nil {
		return nil, err
	}
	cards, err := s.cards.ListByIDs(ctx, cardIDs)
	if err != nil {
		return nil, err
	}
	users, err := s.users.ListByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	return &ledgerRelated{members: members, cards: cards, users: users}, nil
}
