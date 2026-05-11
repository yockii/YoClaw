package skills

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yockii/wangshu/internal/config"
	"github.com/yockii/wangshu/internal/types"
	"github.com/yockii/wangshu/pkg/constant"
	"gopkg.in/yaml.v3"
)

// 技能文件前端匹配正则
var skillFrontmatterReg = regexp.MustCompile(`---\s*\n([\s\S]*?)\n---\s*|\n`)
// SkillCreator 技能创建器，支持从对话历史自动生成技能
type SkillCreator struct {
	workspace string
	loader    *Loader
}

// NewSkillCreator 创建技能创建器
func NewSkillCreator(workspace string, loader *Loader) *SkillCreator {
	return &SkillCreator{
		workspace: workspace,
		loader:    loader,
	}
}

// CreateSkillRequest 创建技能的请求
type CreateSkillRequest struct {
	Name        string   `json:"name"`        // 技能名称
	Description string   `json:"description"` // 技能描述
	Triggers    []string `json:"triggers"`    // 触发关键词
	Content     string   `json:"content"`     // 技能内容（Markdown）
	Category    string   `json:"category"`    // 技能分类（可选）
}

// CreateSkill 创建新技能
func (sc *SkillCreator) CreateSkill(req CreateSkillRequest) error {
	// 验证请求
	if err := sc.validateCreateSkillRequest(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// 创建技能目录
	skillDir := filepath.Join(sc.workspace, req.Name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		return fmt.Errorf("failed to create skill directory: %w", err)
	}

	// 生成技能文件内容
	skillContent := sc.generateSkillContent(req)

	// 写入SKILL.md文件
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(skillContent), 0644); err != nil {
		return fmt.Errorf("failed to write skill file: %w", err)
	}

	slog.Info("Skill created successfully",
		"name", req.Name,
		"location", skillFile)

	return nil
}

// validateCreateSkillRequest 验证创建技能请求
func (sc *SkillCreator) validateCreateSkillRequest(req CreateSkillRequest) error {
	if req.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	if req.Description == "" {
		return fmt.Errorf("skill description is required")
	}
	if req.Content == "" {
		return fmt.Errorf("skill content is required")
	}

	// 检查技能是否已存在
	skillPath := filepath.Join(sc.workspace, req.Name, "SKILL.md")
	if _, err := os.Stat(skillPath); err == nil {
		return fmt.Errorf("skill '%s' already exists", req.Name)
	}

	// 验证技能名称格式（只允许字母、数字、短横线、下划线）
	if !isValidSkillName(req.Name) {
		return fmt.Errorf("skill name '%s' contains invalid characters (only letters, numbers, hyphens, and underscores are allowed)", req.Name)
	}

	return nil
}

// generateSkillContent 生成技能文件内容
func (sc *SkillCreator) generateSkillContent(req CreateSkillRequest) string {
	var sb strings.Builder

	// YAML Frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("name: %s\n", req.Name))
	sb.WriteString(fmt.Sprintf("description: %s\n", req.Description))
	if len(req.Triggers) > 0 {
		sb.WriteString("triggers:\n")
		for _, trigger := range req.Triggers {
			sb.WriteString(fmt.Sprintf("  - %s\n", trigger))
		}
	}
	if req.Category != "" {
		sb.WriteString(fmt.Sprintf("category: %s\n", req.Category))
	}
	sb.WriteString("---\n\n")

	// 技能内容
	sb.WriteString(req.Content)

	return sb.String()
}

// isValidSkillName 验证技能名称格式
func isValidSkillName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// UpdateSkill 更新现有技能
func (sc *SkillCreator) UpdateSkill(name string, req CreateSkillRequest) error {
	// 检查技能是否存在
	skillPath := filepath.Join(sc.workspace, name, "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return fmt.Errorf("skill '%s' does not exist", name)
	}

	// 验证请求
	if req.Name == "" {
		req.Name = name
	}

	// 如果名称变更，需要重命名目录
	if req.Name != name {
		oldDir := filepath.Join(sc.workspace, name)
		newDir := filepath.Join(sc.workspace, req.Name)

		// 检查新名称是否已存在
		if _, err := os.Stat(newDir); err == nil {
			return fmt.Errorf("skill '%s' already exists", req.Name)
		}

		if err := os.Rename(oldDir, newDir); err != nil {
			return fmt.Errorf("failed to rename skill directory: %w", err)
		}

		skillPath = filepath.Join(newDir, "SKILL.md")
	}

	// 生成新内容
	skillContent := sc.generateSkillContent(req)

	// 写入文件
	if err := os.WriteFile(skillPath, []byte(skillContent), 0644); err != nil {
		return fmt.Errorf("failed to update skill file: %w", err)
	}

	slog.Info("Skill updated successfully",
		"old_name", name,
		"new_name", req.Name,
		"location", skillPath)

	return nil
}

// DeleteSkill 删除技能
func (sc *SkillCreator) DeleteSkill(name string) error {
	skillDir := filepath.Join(sc.workspace, name)

	// 检查技能是否存在
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return fmt.Errorf("skill '%s' does not exist", name)
	}

	// 删除技能目录
	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("failed to delete skill directory: %w", err)
	}

	slog.Info("Skill deleted successfully", "name", name)

	return nil
}

// GetSkill 获取技能详细信息
func (sc *SkillCreator) GetSkill(name string) (*types.Skill, string, error) {
	skillPath := filepath.Join(sc.workspace, name, "SKILL.md")

	// 检查技能是否存在
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("skill '%s' does not exist", name)
	}

	// 读取文件
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read skill file: %w", err)
	}

	// 解析frontmatter
	content := string(data)
	matches := skillFrontmatterReg.FindStringSubmatch(content)
	if len(matches) < 2 {
		return nil, "", fmt.Errorf("invalid skill format: missing frontmatter")
	}

	// 解析YAML
	var skill types.Skill
	if err := yaml.NewDecoder(strings.NewReader(matches[1])).Decode(&skill); err != nil {
		return nil, "", fmt.Errorf("failed to parse skill metadata: %w", err)
	}

	// 提取Markdown内容（去掉frontmatter）
	markdownContent := skillFrontmatterReg.ReplaceAllString(content, "")
	markdownContent = strings.TrimSpace(markdownContent)

	return &skill, markdownContent, nil
}

// ListSkills 列出所有技能
func (sc *SkillCreator) ListSkills() ([]types.Skill, error) {
	return sc.loader.LoadSkills()
}

// CreateSkillFromConversation 从对话历史创建技能
// 这是一个高级功能，使用LLM分析对话并生成技能
func (sc *SkillCreator) CreateSkillFromConversation(conversation string) (*CreateSkillRequest, error) {
	// TODO: 实现从对话生成技能的逻辑
	// 这需要调用LLM来分析对话，提取可复用的模式和步骤
	// 目前先返回一个示例实现

	// 示例：从对话中提取关键信息
	req := CreateSkillRequest{
		Name:        "generated-skill-" + time.Now().Format("20060102-150405"),
		Description: "Auto-generated skill from conversation",
		Triggers:    []string{},
		Content:     "# Auto-generated Skill\n\nThis skill was generated from conversation history.",
	}

	return &req, nil
}

// ValidateSkill 验证技能完整性
func (sc *SkillCreator) ValidateSkill(name string) error {
	skill, content, err := sc.GetSkill(name)
	if err != nil {
		return err
	}

	// 基本验证
	if skill.Name == "" {
		return fmt.Errorf("skill name is empty")
	}
	if skill.Description == "" {
		return fmt.Errorf("skill description is empty")
	}
	if content == "" {
		return fmt.Errorf("skill content is empty")
	}

	// 检查是否有明确的触发条件
	if len(skill.Triggers) == 0 {
		slog.Warn("Skill has no triggers defined", "name", name)
	}

	return nil
}

// SearchSkills 搜索技能
func (sc *SkillCreator) SearchSkills(query string) ([]types.Skill, error) {
	allSkills, err := sc.ListSkills()
	if err != nil {
		return nil, err
	}

	var results []types.Skill
	query = strings.ToLower(query)

	for _, skill := range allSkills {
		// 搜索名称、描述和触发词
		if strings.Contains(strings.ToLower(skill.Name), query) ||
			strings.Contains(strings.ToLower(skill.Description), query) ||
			sc.hasMatchingTrigger(skill, query) {
			results = append(results, *skill)
		}
	}

	return results, nil
}

// hasMatchingTrigger 检查技能是否有匹配的触发词
func (sc *SkillCreator) hasMatchingTrigger(skill *types.Skill, query string) bool {
	for _, trigger := range skill.Triggers {
		if strings.Contains(strings.ToLower(trigger), query) {
			return true
		}
	}
	return false
}

// GetSkillStats 获取技能统计信息
func (sc *SkillCreator) GetSkillStats() map[string]interface{} {
	skills, err := sc.ListSkills()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}

	totalSkills := len(skills)
	categories := make(map[string]int)

	for _, skill := range skills {
		if skill.Category != "" {
			categories[skill.Category]++
		} else {
			categories["uncategorized"]++
		}
	}

	return map[string]interface{}{
		"total_skills":  totalSkills,
		"categories":    categories,
		"workspace":     sc.workspace,
		"last_updated": time.Now().Format(time.RFC3339),
	}
}
