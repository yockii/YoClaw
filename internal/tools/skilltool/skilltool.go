package skilltool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yockii/wangshu/internal/agent"
	"github.com/yockii/wangshu/pkg/tools"
	"github.com/yockii/wangshu/pkg/tools/types"
)

// SkillTool 技能管理工具
type SkillTool struct {
	agent *agent.Agent
}

// NewSkillTool 创建技能管理工具
func NewSkillTool() *SkillTool {
	return &SkillTool{}
}

// SetAgent 设置Agent引用
func (t *SkillTool) SetAgent(a *agent.Agent) {
	t.agent = a
}

// Name 工具名称
func (t *SkillTool) Name() string {
	return "skill"
}

// Description 工具描述
func (t *SkillTool) Description() string {
	return "管理技能（创建、更新、删除、列出、搜索、验证技能）"
}

// ParametersSchema 参数模式
func (t *SkillTool) ParametersSchema() tools.Schema {
	return tools.Schema{
		Type: "object",
		Properties: map[string]tools.SchemaProperty{
			"action": {
				Type:        "string",
				Description: "操作类型: create, update, delete, list, get, search, validate, stats, from_conversation",
				Enum:        []string{"create", "update", "delete", "list", "get", "search", "validate", "stats", "from_conversation"},
			},
			"name": {
				Type:        "string",
				Description: "技能名称（对于create/update操作，这是新名称；对于其他操作，这是目标技能名称）",
			},
			"new_name": {
				Type:        "string",
				Description: "新技能名称（仅用于update操作）",
			},
			"description": {
				Type:        "string",
				Description: "技能描述",
			},
			"triggers": {
				Type:        "array",
				Items:       &tools.SchemaProperty{Type: "string"},
				Description: "触发关键词数组",
			},
			"content": {
				Type:        "string",
				Description: "技能内容（Markdown格式）",
			},
			"category": {
				Type:        "string",
				Description: "技能分类（可选）",
			},
			"query": {
				Type:        "string",
				Description: "搜索查询词（用于search操作）",
			},
			"conversation": {
				Type:        "string",
				Description: "对话文本（用于from_conversation操作）",
			},
		},
		Required: []string{"action"},
	}
}

// Execute 执行工具
func (t *SkillTool) Execute(ctx context.Context, args map[string]any, toolCtx types.ToolContext) (string, error) {
	if t.agent == nil {
		return "", fmt.Errorf("agent not initialized")
	}

	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action parameter is required")
	}

	switch action {
	case "create":
		return t.createSkill(args)
	case "update":
		return t.updateSkill(args)
	case "delete":
		return t.deleteSkill(args)
	case "list":
		return t.listSkills()
	case "get":
		return t.getSkill(args)
	case "search":
		return t.searchSkills(args)
	case "validate":
		return t.validateSkill(args)
	case "stats":
		return t.getStats()
	case "from_conversation":
		return t.createFromConversation(args)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

// createSkill 创建技能
func (t *SkillTool) createSkill(args map[string]any) (string, error) {
	name, _ := args["name"].(string)
	description, _ := args["description"].(string)
	content, _ := args["content"].(string)
	category, _ := args["category"].(string)

	// 解析triggers
	var triggers []string
	if triggersRaw, ok := args["triggers"].([]any); ok {
		for _, trigger := range triggersRaw {
			if triggerStr, ok := trigger.(string); ok {
				triggers = append(triggers, triggerStr)
			}
		}
	}

	if name == "" || description == "" || content == "" {
		return "", fmt.Errorf("name, description, and content are required for create action")
	}

	return t.agent.CreateSkill(name, description, triggers, content, category)
}

// updateSkill 更新技能
func (t *SkillTool) updateSkill(args map[string]any) (string, error) {
	name, _ := args["name"].(string)
	newName, _ := args["new_name"].(string)
	description, _ := args["description"].(string)
	content, _ := args["content"].(string)
	category, _ := args["category"].(string)

	// 解析triggers
	var triggers []string
	if triggersRaw, ok := args["triggers"].([]any); ok {
		for _, trigger := range triggersRaw {
			if triggerStr, ok := trigger.(string); ok {
				triggers = append(triggers, triggerStr)
			}
		}
	}

	if name == "" {
		return "", fmt.Errorf("name is required for update action")
	}

	return t.agent.UpdateSkill(name, newName, description, triggers, content, category)
}

// deleteSkill 删除技能
func (t *SkillTool) deleteSkill(args map[string]any) (string, error) {
	name, _ := args["name"].(string)

	if name == "" {
		return "", fmt.Errorf("name is required for delete action")
	}

	return t.agent.DeleteSkill(name)
}

// listSkills 列出技能
func (t *SkillTool) listSkills() (string, error) {
	return t.agent.ListSkills()
}

// getSkill 获取技能详情
func (t *SkillTool) getSkill(args map[string]any) (string, error) {
	name, _ := args["name"].(string)

	if name == "" {
		return "", fmt.Errorf("name is required for get action")
	}

	return t.agent.GetSkill(name)
}

// searchSkills 搜索技能
func (t *SkillTool) searchSkills(args map[string]any) (string, error) {
	query, _ := args["query"].(string)

	if query == "" {
		return "", fmt.Errorf("query is required for search action")
	}

	return t.agent.SearchSkills(query)
}

// validateSkill 验证技能
func (t *SkillTool) validateSkill(args map[string]any) (string, error) {
	name, _ := args["name"].(string)

	if name == "" {
		return "", fmt.Errorf("name is required for validate action")
	}

	return t.agent.ValidateSkill(name)
}

// getStats 获取统计信息
func (t *SkillTool) getStats() (string, error) {
	return t.agent.GetSkillStats()
}

// createFromConversation 从对话创建技能
func (t *SkillTool) createFromConversation(args map[string]any) (string, error) {
	conversation, _ := args["conversation"].(string)
	skillName, _ := args["name"].(string)
	description, _ := args["description"].(string)

	if conversation == "" {
		return "", fmt.Errorf("conversation is required for from_conversation action")
	}

	return t.agent.CreateSkillFromConversation(conversation, skillName, description)
}

// MarshalJSON JSON序列化
func (t *SkillTool) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":        t.Name(),
		"description": t.Description(),
		"parameters":  t.ParametersSchema(),
	})
}
