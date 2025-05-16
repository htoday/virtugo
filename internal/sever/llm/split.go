package llm

import (
	"strings"
	"unicode"
)

// SplitSentence 从 text 中提取首句 sentence（含分隔符）和剩余 rest。
// 仅在已达到 minLen 后遇到句末符号才分割；超过 maxLen 强制截断。
func SplitSentence(text string, minLen, maxLen int) (sentence, rest string) {
	runes := []rune(text)
	isDelim := func(r rune) bool {
		switch r {
		case '。', '！', '!', '？', '?', '；', ';', '.', '…', '，', ',':
			return true
		}
		return false
	}

	for i, r := range runes {
		// 达到最小长度后，遇到合法分隔符
		if i+1 >= minLen && isDelim(r) {
			// 跳过数字小数点情况
			if r == '.' && i > 0 && i < len(runes)-1 &&
				unicode.IsDigit(runes[i-1]) && unicode.IsDigit(runes[i+1]) {
				continue
			}
			end := i + 1
			sentence = strings.TrimSpace(string(runes[:end]))
			rest = string(runes[end:])
			return
		}
		// 超过最大长度强制分段
		if i+1 >= maxLen {
			sentence = strings.TrimSpace(string(runes[:maxLen]))
			rest = string(runes[maxLen:])
			return
		}
	}
	// 不足 maxLen 且无分隔符，返回空 sentence 和原始 text
	return "", text
}
