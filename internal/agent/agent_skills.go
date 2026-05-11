package agent

import (
	"fmt"
	"log/slog"

	"github.com/yockii/wangshu/internal/types"
	"github.com/yockii/wangshu/pkg/skills"
)

// skillManagementTools 技能管理相关工具函数

// CreateSkill 创建新技能
// 参数：name(技能名称), description(技能描述), triggers(触发关键词数组), content(技能内容), category(分类可选)
// 返回：成功/失败消息
func (a *Agent) CreateSkill(name, description string, triggers []string, content, category string) (string, error) {
	// 检查权限（可选：可以添加权限控制）
	// 目前允许所有用户创建技能

	// 创建技能创建器
	creator := skills.NewSkillCreator(a.workspaceDir, skills.GetDefaultLoader())

	// 构建请求
	req := skills.CreateSkillRequest{
		Name:        name,
		Description: description,
		Triggers:    triggers,
		Content:     content,
		Category:    category,
	}

	// 创建技能
	if err := creator.CreateSkill(req); err != nil {
		return "", fmt.Errorf("failed to create skill: %w", err)
	}

	// 重新加载技能列表
	skills.GetDefaultLoader().LoadSkills()

	return fmt.Sprintf("✅ 技能 '%s' 创建成功！\n位置: %s/%s/SKILL.md", name, a.workspaceDir, name), nil
}

// UpdateSkill 更新现有技能
// 参数：name(技能名称), new_name(新名称可选), description(描述), triggers(触发词), content(内容), category(分类)
// 返回：成功/失败消息
func (a *Agent) UpdateSkill(name, newName, description string, triggers []string, content, category string) (string, error) {
	creator := skills.NewSkillCreator(a.workspaceDir, skills.GetDefaultLoader())

	req := skills.CreateSkillRequest{
		Name:        newName,
		Description: description,
		Triggers:    triggers,
		Content:     content,
		Category:    category,
	}

	if err := creator.UpdateSkill(name, req); err != nil {
		return "", fmt.Errorf("failed to update skill: %w", err)
	}

	// 重新加载技能列表
	skills.GetDefaultLoader().LoadSkills()

	return fmt.Sprintf("✅ 技能更新成功！\n名称: %s → %s", name, newName), nil
}

// DeleteSkill 删除技能
// 参数：name(技能名称)
// 返回：成功/失败消息
func (a *Agent) DeleteSkill(name string) (string, error) {
	creator := skills.NewSkillCreator(a.workspaceDir, skills.GetDefaultLoader())

	if err := creator.DeleteSkill(name); err != nil {
		return "", fmt.Errorf("failed to delete skill: %w", err)
	}

	// 重新加载技能列表
	skills.GetDefaultLoader().LoadSkills()

	return fmt.Sprintf("✅ 技能 '%s' 已删除", name), nil
}

// ListSkills 列出所有技能
// 返回：技能列表
func (a *Agent) ListSkills() (string, error) {
	creator := skills.NewSkillCreator(a.workspaceDir, skills.GetDefaultLoader())

	skillList, err := creator.ListSkills()
	if err != nil {
		return "", fmt.Errorf("failed to list skills: %w", err)
	}

	if len(skillList) == 0 {
		return "当前没有可用的技能", nil
	}

	result := fmt.Sprintf("📋 可用技能列表 (%d个):\n\n", len(skillList))
	for i, skill := range skillList {
		result += fmt.Sprintf("%d. **%s**\n", i+1, skill.Name)
		result += fmt.Sprintf("   描述: %s\n", skill.Description)
		if len(skill.Triggers) > 0 {
			result += fmt.Sprintf("   触发词: %v\n", skill.Triggers)
		}
		if skill.Category != "" {
			result += fmt.Sprintf("   分类: %s\n", skill.Category)
		}
		result += "\n"
	}

	return result, nil
}

// GetSkill 获取技能详细信息
// 参数：name(技能名称)
// 返回：技能详情
func (a *Agent) GetSkill(name string) (string, error) {
	creator := skills.NewSkillCreator(a.workspaceDir, skills.GetDefaultLoader())

	skill, content, err := creator.GetSkill(name)
	if err != nil {
		return "", fmt.Errorf("failed to get skill: %w", err)
	}

	result := fmt.Sprintf("📖 技能详情: %s\n\n", skill.Name)
	result += fmt.Sprintf("**描述**: %s\n\n", skill.Description)
	if len(skill.Triggers) > 0 {
		result += fmt.Sprintf("**触发词**: %v\n\n", skill.Triggers)
	}
	if skill.Category != "" {
		result += fmt.Sprintf("**分类**: %s\n\n", skill.Category)
	}
	result += "**内容**:\n\n"
	result += content

	return result, nil
}

// SearchSkills 搜索技能
// 参数：query(搜索关键词)
// 返回：匹配的技能列表
func (a *Agent) SearchSkills(query string) (string, error) {
	creator := skills.NewSkillCreator(a.workspaceDir, skills.GetDefaultLoader())

	results, err := creator.SearchSkills(query)
	if err != nil {
		return "", fmt.Errorf("failed to search skills: %w", err)
	}

	if len(results) == 0 {
		return fmt.Sprintf("未找到匹配 '%s' 的技能", query), nil
	}

	result := fmt.Sprintf("🔍 搜索结果: '%s' (%d个匹配):\n\n", query, len(results))
	for i, skill := range results {
		result += fmt.Sprintf("%d. **%s**\n", i+1, skill.Name)
		result += fmt.Sprintf("   描述: %s\n", skill.Description)
		if len(skill.Triggers) > 0 {
			result += fmt.Sprintf("   触发词: %v\n", skill.Triggers)
		}
		result += "\n"
	}

	return result, nil
}

// ValidateSkill 验证技能
// 参数：name(技能名称)
// 返回：验证结果
func (a *Agent) ValidateSkill(name string) (string, error) {
	creator := skills.NewSkillCreator(a.workspaceDir, skills.GetDefaultLoader())

	if err := creator.ValidateSkill(name); err != nil {
		return fmt.Sprintf("❌ 技能验证失败: %v", err), nil
	}

	return fmt.Sprintf("✅ 技能 '%s' 验证通过", name), nil
}

// GetSkillStats 获取技能统计信息
// 返回：统计信息
func (a *Agent) GetSkillStats() (string, error) {
	creator := skills.NewSkillCreator(a.workspaceDir, skills.GetDefaultLoader())

	stats := creator.GetSkillStats()

	result := "📊 技能统计信息:\n\n"
	result += fmt.Sprintf("- 总技能数: %v\n", stats["total_skills"])
	if categories, ok := stats["categories"].(map[string]int); ok {
		result += "\n分类统计:\n"
		for category, count := range categories {
			result += fmt.Sprintf("  - %s: %d\n", category, count)
		}
	}
	result += fmt.Sprintf("\n工作空间: %v\n", stats["workspace"])
	result += fmt.Sprintf("最后更新: %v\n", stats["last_updated"])

	slog.Info("Skill stats requested", "stats", stats)

	return result, nil
}

// CreateSkillFromConversation 从对话创建技能
// 参数：conversation_text(对话文本), skill_name(技能名称), description(描述)
// 返回：创建的技能预览
func (a *Agent) CreateSkillFromConversation(conversationText, skillName, description string) (string, error) {
	creator := skills.NewSkillCreator(a.workspaceDir, skills.GetDefaultLoader())

	// 调用技能创建器
	req, err := creator.CreateSkillFromConversation(conversationText)
	if err != nil {
		return "", fmt.Errorf("failed to create skill from conversation: %w", err)
	}

	// 使用提供的名称和描述（如果有）
	if skillName != "" {
		req.Name = skillName
	}
	if description != "" {
		req.Description = description
	}

	// 创建技能
	if err := creator.CreateSkill(*req); err != nil {
		return "", fmt.Errorf("failed to save skill: %w", err)
	}

	// 重新加载技能列表
	skills.GetDefaultLoader().LoadSkills()

	result := fmt.Sprintf("✅ 从对话创建技能成功！\n\n")
	result += fmt.Sprintf("**技能名称**: %s\n", req.Name)
	result += fmt.Sprintf("**描述**: %s\n", req.Description)
	if len(req.Triggers) > 0 {
		result += fmt.Sprintf("**触发词**: %v\n", req.Triggers)
	}
	result += fmt.Sprintf("\n位置: %s/%s/SKILL.md", a.workspaceDir, req.Name)

	return result, nil
}

// RecordSkillUsage 记录技能使用情况（用于生命周期管理）
// 参数：skillName(技能名称), success(是否成功)
func (a *Agent) RecordSkillUsage(skillName string, success bool) {
	lifecycle := skills.NewSkillLifecycleManager(a.workspaceDir)

	if err := lifecycle.LoadStats(); err != nil {
		slog.Warn("Failed to load skill stats", "error", err)
		return
	}

	if err := lifecycle.RecordUsage(skillName, success); err != nil {
		slog.Warn("Failed to record skill usage", "skill", skillName, "error", err)
	}
}
