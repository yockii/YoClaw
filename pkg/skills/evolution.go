package skills

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// SkillEvolutionDecision 技能进化决策
type SkillEvolutionDecision struct {
	Action         string            `json:"action"`          // "create", "update", "none"
	Reasoning      string            `json:"reasoning"`       // 决策原因
	Confidence     float64           `json:"confidence"`      // 置信度 (0-1)
	SuggestedSkill *CreateSkillRequest `json:"suggested_skill,omitempty"` // 建议的技能
	ExistingSkill  string            `json:"existing_skill,omitempty"`  // 现有技能名称
	RelatedSkills  []string          `json:"related_skills,omitempty"`  // 相关技能
}

// SkillEvolutionEngine 技能进化引擎
type SkillEvolutionEngine struct {
	creator     *SkillCreator
	lifecycle   *SkillLifecycleManager
	workspace   string
	loader      *Loader
}

// NewSkillEvolutionEngine 创建技能进化引擎
func NewSkillEvolutionEngine(workspace string) *SkillEvolutionEngine {
	loader := GetDefaultLoader()
	if loader == nil {
		loader = NewLoader(workspace)
	}

	return &SkillEvolutionEngine{
		creator:   NewSkillCreator(workspace, loader),
		lifecycle: NewSkillLifecycleManager(workspace),
		workspace: workspace,
		loader:    loader,
	}
}

// AnalyzeConversationForEvolution 分析对话内容，判断是否需要技能进化
func (e *SkillEvolutionEngine) AnalyzeConversationForEvolution(conversation string) (*SkillEvolutionDecision, error) {
	// 1. 提取对话中的关键模式和模式
	patterns := e.extractPatterns(conversation)

	// 2. 检查是否存在相关技能
	matchingSkills := e.findMatchingSkills(patterns)

	// 3. 做出进化决策
	decision := e.makeEvolutionDecision(conversation, patterns, matchingSkills)

	return decision, nil
}

// extractPatterns 从对话中提取关键模式
func (e *SkillEvolutionEngine) extractPatterns(conversation string) map[string]int {
	patterns := make(map[string]int)

	// 常见操作模式
	actionPatterns := []string{
		"文件操作", "数据处理", "网络请求", "数据库", "API调用",
		"图片处理", "文本分析", "代码生成", "数据分析", "自动化",
		"格式转换", "批量处理", "定时任务", "通知提醒", "搜索",
		"验证", "计算", "翻译", "摘要", "分类",
	}

	for _, pattern := range actionPatterns {
		count := strings.Count(conversation, pattern)
		if count > 0 {
			patterns[pattern] = count
		}
	}

	// 检测重复步骤（可能是可自动化的流程）
	stepIndicators := []string{
		"第一步", "然后", "接着", "之后", "最后", "首先",
		"第1步", "第2步", "第3步", "step1", "step2", "step3",
	}

	stepCount := 0
	for _, indicator := range stepIndicators {
		stepCount += strings.Count(conversation, indicator)
	}

	if stepCount >= 2 {
		patterns["多步骤流程"] = stepCount
	}

	return patterns
}

// findMatchingSkills 查找匹配的现有技能
func (e *SkillEvolutionEngine) findMatchingSkills(patterns map[string]int) []string {
	skills, err := e.loader.LoadSkills()
	if err != nil {
		slog.Error("Failed to load skills", "error", err)
		return nil
	}

	var matchingSkills []string

	for _, skill := range skills {
		// 检查技能描述和触发词
		matched := false

		// 检查描述
		for pattern := range patterns {
			if strings.Contains(strings.ToLower(skill.Description), strings.ToLower(pattern)) {
				matched = true
				break
			}
		}

		// 检查触发词
		if !matched {
			for _, trigger := range skill.Triggers {
				for pattern := range patterns {
					if strings.Contains(strings.ToLower(trigger), strings.ToLower(pattern)) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
		}

		if matched {
			matchingSkills = append(matchingSkills, skill.Name)
		}
	}

	return matchingSkills
}

// makeEvolutionDecision 做出进化决策
func (e *SkillEvolutionEngine) makeEvolutionDecision(
	conversation string,
	patterns map[string]int,
	matchingSkills []string,
) *SkillEvolutionDecision {

	decision := &SkillEvolutionDecision{
		RelatedSkills: matchingSkills,
	}

	// 决策逻辑
	if len(matchingSkills) > 0 {
		// 有匹配技能，考虑更新
		if e.shouldUpdateExistingSkill(conversation, patterns, matchingSkills) {
			decision.Action = "update"
			decision.Reasoning = fmt.Sprintf(
				"发现 %d 个相关技能 (%s)，对话中包含新的模式或改进机会",
				len(matchingSkills), strings.Join(matchingSkills, ", "))
			decision.Confidence = 0.7
			decision.ExistingSkill = matchingSkills[0] // 选择第一个匹配的技能
		} else {
			decision.Action = "none"
			decision.Reasoning = "现有技能已足够，无需创建或更新"
			decision.Confidence = 0.6
		}
	} else if len(patterns) > 0 {
		// 没有匹配技能但有模式，考虑创建新技能
		if e.shouldCreateNewSkill(patterns) {
			decision.Action = "create"
			decision.Reasoning = fmt.Sprintf(
				"检测到 %d 个可重复模式，但无匹配技能",
				len(patterns))
			decision.Confidence = 0.8

			// 生成技能建议
			decision.SuggestedSkill = e.suggestSkillFromPatterns(patterns, conversation)
		} else {
			decision.Action = "none"
			decision.Reasoning = "检测到模式但不够明显或不够重复"
			decision.Confidence = 0.5
		}
	} else {
		decision.Action = "none"
		decision.Reasoning = "未检测到明显的可重复模式"
		decision.Confidence = 0.9
	}

	return decision
}

// shouldUpdateExistingSkill 判断是否应该更新现有技能
func (e *SkillEvolutionEngine) shouldUpdateExistingSkill(
	conversation string,
	patterns map[string]int,
	matchingSkills []string,
) bool {
	// 1. 检查现有技能的使用统计
	for _, skillName := range matchingSkills {
		if stats, exists := e.lifecycle.GetStats(skillName); exists {
			// 如果技能频繁使用且成功率高，可能需要优化
			if stats.TotalUsage >= 10 && stats.SuccessRate > 0.7 {
				// 检查对话中是否有新的模式
				newPatterns := e.findNewPatterns(conversation, skillName)
				if len(newPatterns) > 0 {
					slog.Info("Skill evolution suggested",
						"skill", skillName,
						"new_patterns", len(newPatterns))
					return true
				}
			}
		}
	}

	// 2. 检查对话长度（长对话可能包含新需求）
	if len(conversation) > 1000 {
		return true
	}

	return false
}

// shouldCreateNewSkill 判断是否应该创建新技能
func (e *SkillEvolutionEngine) shouldCreateNewSkill(patterns map[string]int) bool {
	// 1. 至少有2个明显的模式
	if len(patterns) < 2 {
		return false
	}

	// 2. 模式出现的总次数足够多
	totalOccurrences := 0
	for _, count := range patterns {
		totalOccurrences += count
	}

	if totalOccurrences < 3 {
		return false
	}

	// 3. 存在多步骤流程
	if _, hasWorkflow := patterns["多步骤流程"]; hasWorkflow {
		return true
	}

	return false
}

// findNewPatterns 查找新模式
func (e *SkillEvolutionEngine) findNewPatterns(conversation string, skillName string) []string {
	skill, content, err := e.creator.GetSkill(skillName)
	if err != nil {
		return nil
	}

	// 获取现有技能的关键词
	existingKeywords := make(map[string]bool)
	for _, trigger := range skill.Triggers {
		existingKeywords[strings.ToLower(trigger)] = true
	}

	// 检查对话中是否有新关键词
	newPatterns := []string{}
	for _, trigger := range skill.Triggers {
		if !strings.Contains(strings.ToLower(content), strings.ToLower(trigger)) {
			if strings.Contains(strings.ToLower(conversation), strings.ToLower(trigger)) {
				newPatterns = append(newPatterns, trigger)
			}
		}
	}

	return newPatterns
}

// suggestSkillFromPatterns 根据模式建议技能
func (e *SkillEvolutionEngine) suggestSkillFromPatterns(
	patterns map[string]int,
	conversation string,
) *CreateSkillRequest {
	// 提取主要模式
	var mainPattern string
	maxCount := 0
	for pattern, count := range patterns {
		if count > maxCount {
			maxCount = count
			mainPattern = pattern
		}
	}

	// 生成技能名称
	skillName := fmt.Sprintf("auto-%s", strings.ToLower(strings.ReplaceAll(mainPattern, " ", "-")))
	skillName = strings.ReplaceAll(skillName, "/", "-")

	// 生成描述
	description := fmt.Sprintf("自动化%s的技能，根据对话模式自动生成", mainPattern)

	// 提取触发词
	var triggers []string
	for pattern := range patterns {
		if len(pattern) > 2 { // 过滤过短的词
			triggers = append(triggers, pattern)
		}
	}

	// 生成基础内容
	content := fmt.Sprintf(`# %s

## 概述
%s

## 自动检测的模式
%s

## 使用场景
基于用户对话自动生成的技能，用于处理%s相关的任务。

## 执行步骤
1. 分析用户请求
2. 执行%s操作
3. 返回结果

## 注意事项
- 此技能由系统自动生成，建议根据实际使用情况进行优化
- 可以根据具体需求调整触发词和执行流程
`,
		mainPattern,
		description,
		generatePatternList(patterns),
		mainPattern,
		mainPattern)

	return &CreateSkillRequest{
		Name:        skillName,
		Description: description,
		Triggers:    triggers,
		Content:     content,
		Category:    "auto-generated",
	}
}

// generatePatternList 生成模式列表
func generatePatternList(patterns map[string]int) string {
	var sb strings.Builder
	for pattern, count := range patterns {
		sb.WriteString(fmt.Sprintf("- %s: 出现 %d 次\n", pattern, count))
	}
	return sb.String()
}

// ExecuteEvolution 执行进化决策
func (e *SkillEvolutionEngine) ExecuteEvolution(decision *SkillEvolutionDecision) error {
	switch decision.Action {
	case "create":
		if decision.SuggestedSkill != nil {
			return e.creator.CreateSkill(*decision.SuggestedSkill)
		}
	case "update":
		if decision.ExistingSkill != "" && decision.SuggestedSkill != nil {
			return e.creator.UpdateSkill(
				decision.ExistingSkill,
				decision.SuggestedSkill.Name,
				decision.SuggestedSkill.Description,
				decision.SuggestedSkill.Triggers,
				decision.SuggestedSkill.Content,
				decision.SuggestedSkill.Category,
			)
		}
	}

	return nil
}

// SaveEvolutionLog 保存进化日志
func (e *SkillEvolutionEngine) SaveEvolutionLog(decision *SkillEvolutionDecision, conversation string) error {
	logPath := filepath.Join(e.workspace, ".skill_evolution_log.json")

	var logs []map[string]interface{}
	if data, err := os.ReadFile(logPath); err == nil {
		json.Unmarshal(data, &logs)
	}

	logEntry := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"action":      decision.Action,
		"reasoning":   decision.Reasoning,
		"confidence":  decision.Confidence,
		"conversation_length": len(conversation),
		"related_skills": decision.RelatedSkills,
	}

	if decision.ExistingSkill != "" {
		logEntry["existing_skill"] = decision.ExistingSkill
	}

	if decision.SuggestedSkill != nil {
		logEntry["suggested_skill"] = map[string]interface{}{
			"name": decision.SuggestedSkill.Name,
			"description": decision.SuggestedSkill.Description,
		}
	}

	logs = append(logs, logEntry)

	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(logPath, data, 0644)
}
