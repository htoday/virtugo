package llm

import "regexp"

// FilterEmoji 过滤字符串中的emoji表情符号
func FilterEmoji(s string) string {
	emojiRegex := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|` + // 表情符号
		`[\x{1F300}-\x{1F5FF}]|` + // 各类图形
		`[\x{1F680}-\x{1F6FF}]|` + // 交通工具等
		`[\x{1F700}-\x{1F77F}]|` + // 炼金术符号
		`[\x{1F780}-\x{1F7FF}]|` + // 几何图形
		`[\x{1F800}-\x{1F8FF}]|` + // 补充箭头符号
		`[\x{1F900}-\x{1F9FF}]|` + // 补充表情符号
		`[\x{1FA00}-\x{1FA6F}]|` + // 补充象形文字
		`[\x{1FA70}-\x{1FAFF}]|` + // 表情补充
		`[\x{2600}-\x{26FF}]|` + // 杂项符号
		`[\x{2700}-\x{27BF}]|` + // 杂项符号和箭头
		`[\x{FE00}-\x{FE0F}]|` + // 变体选择符
		`[\x{1F1E6}-\x{1F1FF}]`) // 旗帜

	return emojiRegex.ReplaceAllString(s, "")
}
