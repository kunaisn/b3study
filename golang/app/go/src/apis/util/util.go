package util

import (
	"fmt"
	"strconv"
	"strings"
)

// JoinStringByTabは文字列の配列をタブ区切りで結合して戻す
func JoinStringByTab(ary []string) string {
	if len(ary) == 0 {
		return ""
	}
	var bld strings.Builder
	for _, s := range ary {
		bld.WriteByte('\t')
		bld.WriteString(s)
	}
	return bld.String()[1:]
}

// JoinIntByTabは整数の配列をタブ区切りで結合して、文字列にして戻す
func JoinIntByTab(ary []int) string {
	if len(ary) == 0 {
		return ""
	}
	var bld strings.Builder
	for _, s := range ary {
		bld.WriteByte('\t')
		bld.WriteString(fmt.Sprint(s))
	}
	return bld.String()[1:]
}

// SplitIntByTabはタブで区切られた数字をint型の配列にして戻す
func SplitIntByTab(str string) ([]int, error) {
	aryStr := strings.Split(str, "\t")
	if len(aryStr) != 0 && aryStr[0] == "" {
		return []int{}, nil
	}
	var aryInt []int
	for _, v := range aryStr {
		c, err := strconv.Atoi(v)
		if err != nil {
			return nil, err
		}
		aryInt = append(aryInt, c)
	}
	return aryInt, nil
}

func CopyStrAry(str []string) []string {
	ans := make([]string, len(str))
	copy(ans, str)
	return ans
}

// searchIdは与えられた文字列に対応する整数の配列を戻す
func SearchId(m map[string]int, strAry []string) []int {
	var ary []int
	for _, str := range strAry {
		if v, found := m[str]; found {
			ary = append(ary, v)
		}
	}
	return ary
}

// truncTailBracketsTextは最後の（...）に囲まれた部分を切り捨てる
func TruncTailBracketsText(text string) string {
	endBracket := 0
	cutIdx := len(text)
	for i := len(text) - 1; i >= 0; i-- {
		switch text[i] {
		case ')':
			endBracket++
		case '(':
			endBracket--
		}
		if text[i] == '.' && endBracket == 0 {
			cutIdx = i
			break
		}
	}
	return text[:cutIdx]
}

// cutRemoveWordsは指定の記号を除いた文字列を戻す
func CutRemoveWords(str string) string {
	const removeWords = `!@#$%^&*()[]"';:\|/<>-_=+~,.`
	var got string
	for i := 0; i < len(str); i++ {
		hit := false
		for _, v := range removeWords {
			if byte(v) == str[i] {
				hit = true
				break
			}
		}
		if hit {
			continue
		}
		got += str[i : i+1]
	}
	return got
}
