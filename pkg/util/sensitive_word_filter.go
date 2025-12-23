package util

import (
	"strings"
	"sync"
)

// SensitiveWordFilter 敏感词过滤器（使用DFA算法）
type SensitiveWordFilter struct {
	root *trieNode
	mu   sync.RWMutex
}

// trieNode 字典树节点
type trieNode struct {
	children    map[rune]*trieNode
	isEnd       bool
	word        string // 完整敏感词
	level       int    // 敏感等级：1-一般 2-严重 3-非常严重
	action      int    // 处理动作：1-替换 2-拦截 3-人工审核
	replacement string // 替换词
}

// NewSensitiveWordFilter 创建敏感词过滤器
func NewSensitiveWordFilter() *SensitiveWordFilter {
	return &SensitiveWordFilter{
		root: &trieNode{
			children: make(map[rune]*trieNode),
		},
	}
}

// AddWord 添加敏感词
func (f *SensitiveWordFilter) AddWord(word string, level, action int, replacement string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// 转换为小写，支持英文大小写不敏感
	word = strings.ToLower(word)
	runes := []rune(word)

	node := f.root
	for _, r := range runes {
		if node.children[r] == nil {
			node.children[r] = &trieNode{
				children: make(map[rune]*trieNode),
			}
		}
		node = node.children[r]
	}

	node.isEnd = true
	node.word = word
	node.level = level
	node.action = action
	node.replacement = replacement
}

// AddWords 批量添加敏感词
func (f *SensitiveWordFilter) AddWords(words []SensitiveWord) {
	for _, word := range words {
		f.AddWord(word.Word, word.Level, word.Action, word.Replacement)
	}
}

// SensitiveWord 敏感词结构
type SensitiveWord struct {
	Word        string
	Level       int
	Action      int
	Replacement string
}

// Contains 检查文本是否包含敏感词
func (f *SensitiveWordFilter) Contains(text string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	text = strings.ToLower(text)
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		node := f.root
		j := i

		for j < len(runes) {
			r := runes[j]
			if node.children[r] == nil {
				break
			}
			node = node.children[r]
			j++

			if node.isEnd {
				return true
			}
		}
	}

	return false
}

// FindAll 查找所有敏感词
func (f *SensitiveWordFilter) FindAll(text string) []MatchResult {
	f.mu.RLock()
	defer f.mu.RUnlock()

	text = strings.ToLower(text)
	runes := []rune(text)
	results := []MatchResult{}

	for i := 0; i < len(runes); i++ {
		node := f.root
		j := i
		maxMatchEnd := -1
		var matchedNode *trieNode

		for j < len(runes) {
			r := runes[j]
			if node.children[r] == nil {
				break
			}
			node = node.children[r]
			j++

			if node.isEnd {
				maxMatchEnd = j
				matchedNode = node
			}
		}

		if maxMatchEnd != -1 {
			results = append(results, MatchResult{
				Word:        matchedNode.word,
				StartPos:    i,
				EndPos:      maxMatchEnd,
				Level:       matchedNode.level,
				Action:      matchedNode.action,
				Replacement: matchedNode.replacement,
			})
			i = maxMatchEnd - 1 // 跳过已匹配的部分
		}
	}

	return results
}

// MatchResult 匹配结果
type MatchResult struct {
	Word        string
	StartPos    int
	EndPos      int
	Level       int
	Action      int
	Replacement string
}

// Replace 替换文本中的敏感词
func (f *SensitiveWordFilter) Replace(text string, replaceChar rune) string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	text = strings.ToLower(text)
	runes := []rune(text)
	replaced := make([]rune, len(runes))
	copy(replaced, runes)

	for i := 0; i < len(runes); i++ {
		node := f.root
		j := i
		maxMatchEnd := -1
		var matchedNode *trieNode

		for j < len(runes) {
			r := runes[j]
			if node.children[r] == nil {
				break
			}
			node = node.children[r]
			j++

			if node.isEnd {
				maxMatchEnd = j
				matchedNode = node
			}
		}

		if maxMatchEnd != -1 {
			// 根据配置的action决定如何处理
			if matchedNode.action == 1 { // 替换
				if matchedNode.replacement != "" {
					// 使用指定的替换词
					replacementRunes := []rune(matchedNode.replacement)
					for k := 0; k < maxMatchEnd-i && k < len(replacementRunes); k++ {
						replaced[i+k] = replacementRunes[k]
					}
				} else {
					// 使用默认替换字符
					for k := i; k < maxMatchEnd; k++ {
						replaced[k] = replaceChar
					}
				}
			}
			i = maxMatchEnd - 1
		}
	}

	return string(replaced)
}

// ReplaceWithCustom 使用自定义替换规则替换敏感词
func (f *SensitiveWordFilter) ReplaceWithCustom(text string) string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	originalRunes := []rune(text)
	lowerText := strings.ToLower(text)
	lowerRunes := []rune(lowerText)
	replaced := make([]rune, len(originalRunes))
	copy(replaced, originalRunes)

	for i := 0; i < len(lowerRunes); i++ {
		node := f.root
		j := i
		maxMatchEnd := -1
		var matchedNode *trieNode

		for j < len(lowerRunes) {
			r := lowerRunes[j]
			if node.children[r] == nil {
				break
			}
			node = node.children[r]
			j++

			if node.isEnd {
				maxMatchEnd = j
				matchedNode = node
			}
		}

		if maxMatchEnd != -1 && matchedNode.action == 1 {
			// 只处理action=1（替换）的敏感词
			if matchedNode.replacement != "" {
				// 使用自定义替换词
				replacementRunes := []rune(matchedNode.replacement)
				replLen := len(replacementRunes)
				wordLen := maxMatchEnd - i

				// 用替换词覆盖原词
				for k := 0; k < wordLen; k++ {
					if k < replLen {
						replaced[i+k] = replacementRunes[k]
					} else {
						replaced[i+k] = '*'
					}
				}
			} else {
				// 使用星号替换
				for k := i; k < maxMatchEnd; k++ {
					replaced[k] = '*'
				}
			}
			i = maxMatchEnd - 1
		}
	}

	return string(replaced)
}

// Validate 验证文本是否包含需要拦截的敏感词
func (f *SensitiveWordFilter) Validate(text string) (bool, []string) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	text = strings.ToLower(text)
	runes := []rune(text)
	blockedWords := []string{}

	for i := 0; i < len(runes); i++ {
		node := f.root
		j := i
		maxMatchEnd := -1
		var matchedNode *trieNode

		for j < len(runes) {
			r := runes[j]
			if node.children[r] == nil {
				break
			}
			node = node.children[r]
			j++

			if node.isEnd {
				maxMatchEnd = j
				matchedNode = node
			}
		}

		if maxMatchEnd != -1 {
			// action=2表示拦截
			if matchedNode.action == 2 {
				blockedWords = append(blockedWords, matchedNode.word)
			}
			i = maxMatchEnd - 1
		}
	}

	return len(blockedWords) == 0, blockedWords
}

// GetHighestRiskLevel 获取文本中敏感词的最高风险等级
func (f *SensitiveWordFilter) GetHighestRiskLevel(text string) int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	text = strings.ToLower(text)
	runes := []rune(text)
	maxLevel := 0

	for i := 0; i < len(runes); i++ {
		node := f.root
		j := i
		maxMatchEnd := -1
		var matchedNode *trieNode

		for j < len(runes) {
			r := runes[j]
			if node.children[r] == nil {
				break
			}
			node = node.children[r]
			j++

			if node.isEnd {
				maxMatchEnd = j
				matchedNode = node
			}
		}

		if maxMatchEnd != -1 {
			if matchedNode.level > maxLevel {
				maxLevel = matchedNode.level
			}
			i = maxMatchEnd - 1
		}
	}

	return maxLevel
}

// ==================== 全局敏感词过滤器 ====================

var (
	globalFilter     *SensitiveWordFilter
	globalFilterOnce sync.Once
)

// GetGlobalFilter 获取全局敏感词过滤器（单例）
func GetGlobalFilter() *SensitiveWordFilter {
	globalFilterOnce.Do(func() {
		globalFilter = NewSensitiveWordFilter()
		// 初始化一些常见敏感词
		initDefaultSensitiveWords()
	})
	return globalFilter
}

// initDefaultSensitiveWords 初始化默认敏感词
func initDefaultSensitiveWords() {
	defaultWords := []SensitiveWord{
		// 脏话类（等级3-非常严重，动作1-替换）
		{Word: "傻逼", Level: 3, Action: 1, Replacement: "**"},
		{Word: "傻b", Level: 3, Action: 1, Replacement: "**"},
		{Word: "fuck", Level: 3, Action: 1, Replacement: "****"},
		{Word: "shit", Level: 3, Action: 1, Replacement: "****"},
		{Word: "垃圾", Level: 2, Action: 1, Replacement: "**"},
		{Word: "白痴", Level: 2, Action: 1, Replacement: "**"},
		{Word: "智障", Level: 2, Action: 1, Replacement: "**"},

		// 广告类（等级2-严重，动作2-拦截）
		{Word: "加微信", Level: 2, Action: 2, Replacement: ""},
		{Word: "加qq", Level: 2, Action: 2, Replacement: ""},
		{Word: "刷单", Level: 2, Action: 2, Replacement: ""},
		{Word: "代理", Level: 2, Action: 3, Replacement: ""}, // 人工审核
		{Word: "兼职", Level: 1, Action: 3, Replacement: ""}, // 人工审核

		// 政治敏感类（等级3，动作2-拦截）
		// 注意：这里只是示例，实际生产环境需要更完善的词库
	}

	for _, word := range defaultWords {
		globalFilter.AddWord(word.Word, word.Level, word.Action, word.Replacement)
	}
}

// LoadSensitiveWordsFromDB 从数据库加载敏感词
func LoadSensitiveWordsFromDB(words []SensitiveWord) {
	filter := GetGlobalFilter()
	filter.AddWords(words)
}

// ==================== 便捷方法 ====================

// ContainsSensitiveWord 检查文本是否包含敏感词
func ContainsSensitiveWord(text string) bool {
	return GetGlobalFilter().Contains(text)
}

// ReplaceSensitiveWord 替换文本中的敏感词
func ReplaceSensitiveWord(text string) string {
	return GetGlobalFilter().ReplaceWithCustom(text)
}

// ValidateSensitiveWord 验证文本（返回是否通过和被拦截的词）
func ValidateSensitiveWord(text string) (bool, []string) {
	return GetGlobalFilter().Validate(text)
}

// FindSensitiveWords 查找所有敏感词
func FindSensitiveWords(text string) []MatchResult {
	return GetGlobalFilter().FindAll(text)
}

// GetSensitiveWordRiskLevel 获取文本风险等级
func GetSensitiveWordRiskLevel(text string) int {
	return GetGlobalFilter().GetHighestRiskLevel(text)
}
