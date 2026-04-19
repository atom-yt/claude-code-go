package tui

import (
	"strings"
)

// AutocompleteState 管理自动补全状态
type AutocompleteState struct {
	visible      bool     // 是否显示补全菜单
	query        string   // 当前搜索查询（不包含 /）
	suggestions  []string // 过滤后的命令列表
	selectedIndex int     // 当前选中的建议索引
}

// showAutocomplete 显示斜杠命令自动补全
func (m *Model) showAutocomplete(query string) {
	m.autocomplete = &AutocompleteState{
		visible:      true,
		query:        strings.TrimPrefix(query, "/"),
		suggestions:  m.filterCommands(strings.TrimPrefix(query, "/")),
		selectedIndex: 0,
	}
}

// hideAutocomplete 关闭补全菜单
func (m *Model) hideAutocomplete() {
	m.autocomplete = nil
}

// isAutocompleteActive 返回补全菜单是否可见
func (m *Model) isAutocompleteActive() bool {
	return m.autocomplete != nil && m.autocomplete.visible
}

// updateAutocompleteQuery 更新搜索查询并过滤建议
func (m *Model) updateAutocompleteQuery(newQuery string) {
	m.autocomplete.query = strings.TrimPrefix(newQuery, "/")
	m.autocomplete.suggestions = m.filterCommands(m.autocomplete.query)

	// 重置选择索引如果超出范围
	if m.autocomplete.selectedIndex >= len(m.autocomplete.suggestions) {
		if len(m.autocomplete.suggestions) > 0 {
			m.autocomplete.selectedIndex = len(m.autocomplete.suggestions) - 1
		} else {
			m.autocomplete.selectedIndex = 0
		}
	}
}

// selectNextAutocomplete 选择下一个建议
func (m *Model) selectNextAutocomplete() {
	if !m.isAutocompleteActive() || len(m.autocomplete.suggestions) == 0 {
		return
	}
	m.autocomplete.selectedIndex = (m.autocomplete.selectedIndex + 1) % len(m.autocomplete.suggestions)
}

// selectPrevAutocomplete 选择上一个建议
func (m *Model) selectPrevAutocomplete() {
	if !m.isAutocompleteActive() || len(m.autocomplete.suggestions) == 0 {
		return
	}
	if m.autocomplete.selectedIndex == 0 {
		m.autocomplete.selectedIndex = len(m.autocomplete.suggestions) - 1
	} else {
		m.autocomplete.selectedIndex--
	}
}

// acceptAutocomplete 用选中的建议替换输入
func (m *Model) acceptAutocomplete() {
	if !m.isAutocompleteActive() || len(m.autocomplete.suggestions) == 0 {
		return
	}
	selected := m.autocomplete.suggestions[m.autocomplete.selectedIndex]
	m.input = "/" + selected
	m.hideAutocomplete()
}

// cycleAutocomplete 循环选择建议（Tab 键行为）
func (m *Model) cycleAutocomplete() {
	if !m.isAutocompleteActive() || len(m.autocomplete.suggestions) == 0 {
		return
	}
	selected := m.autocomplete.suggestions[m.autocomplete.selectedIndex]
	m.input = "/" + selected
	m.hideAutocomplete()
}

// filterCommands 返回匹配查询的命令（不区分大小写）
func (m *Model) filterCommands(query string) []string {
	var matches []string
	queryLower := strings.ToLower(query)

	// 收集所有命令名和别名
	for _, cmd := range m.cmdRegistry.List() {
		name := cmd.Name()
		if queryLower == "" || strings.HasPrefix(strings.ToLower(name), queryLower) {
			matches = append(matches, name)
		}

		// 检查别名
		for _, alias := range cmd.Aliases() {
			if queryLower == "" || strings.HasPrefix(strings.ToLower(alias), queryLower) {
				// 仅当主命令不在匹配列表时才添加别名
				alreadyAdded := false
				for _, m := range matches {
					if m == alias {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					matches = append(matches, alias)
				}
			}
		}
	}

	return matches
}
