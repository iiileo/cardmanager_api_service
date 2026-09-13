package pinyinutil

import (
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

var pyArgs = func() pinyin.Args {
	a := pinyin.NewArgs()
	a.Style = pinyin.Normal
	a.Heteronym = false
	return a
}()

// NameKeys 根据姓名生成全拼与首拼（小写、无空格），用于检索。
// 例：「李明」→ liming / lm；「Anna李」→ annali / annal。
func NameKeys(name string) (full, initials string) {
	var fullB, initB strings.Builder
	for _, r := range strings.TrimSpace(name) {
		switch {
		case unicode.Is(unicode.Han, r):
			py := pinyin.SinglePinyin(r, pyArgs)
			if len(py) == 0 || py[0] == "" {
				continue
			}
			s := strings.ToLower(py[0])
			fullB.WriteString(s)
			initB.WriteByte(s[0])
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			c := unicode.ToLower(r)
			fullB.WriteRune(c)
			initB.WriteRune(c)
		}
	}
	return fullB.String(), initB.String()
}

// NormalizeQuery 规范化检索词：去空白并转小写。
func NormalizeQuery(q string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(q)), ""))
}
