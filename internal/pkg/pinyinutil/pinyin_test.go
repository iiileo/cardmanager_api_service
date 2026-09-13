package pinyinutil

import "testing"

func TestNameKeys(t *testing.T) {
	cases := []struct {
		name, full, initials string
	}{
		{"李明", "liming", "lm"},
		{" 张三丰 ", "zhangsanfeng", "zsf"},
		{"Anna", "anna", "anna"},
		{"王小2", "wangxiao2", "wx2"},
	}
	for _, c := range cases {
		full, initials := NameKeys(c.name)
		if full != c.full || initials != c.initials {
			t.Fatalf("%q: got %q/%q want %q/%q", c.name, full, initials, c.full, c.initials)
		}
	}
}

func TestNormalizeQuery(t *testing.T) {
	if got := NormalizeQuery(" Li Ming "); got != "liming" {
		t.Fatalf("got %q", got)
	}
}
