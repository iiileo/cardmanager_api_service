package dto

import "testing"

func TestNewListPageHasMore(t *testing.T) {
	pg := NewListPage(1, 20, 45)
	if !pg.HasMore {
		t.Fatal("expected has_more true")
	}
	pg = NewListPage(3, 20, 45)
	if pg.HasMore {
		t.Fatal("expected has_more false")
	}
}
