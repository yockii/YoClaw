package skills

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// SkillUsageStats 技能使用统计
type SkillUsageStats struct {
	TotalUsage       int       `json:"total_usage"`        // 总使用次数
	LastUsedAt       time.Time `json:"last_used_at"`       // 最后使用时间
	FirstUsedAt      time.Time `json:"first_used_at"`      // 首次使用时间
	SuccessRate      float64   `json:"success_rate"`       // 成功率
	RecentUsageCount int       `json:"recent_usage_count"` // 最近30天使用次数
}

// SkillLifecycleManager 技能生命周期管理器
type SkillLifecycleManager struct {
	workspace       string
	statsPath       string
	archivePath     string
	statsFile       string
	stats           map[string]*SkillUsageStats
	archiveThreshold int // 归档阈值（天数）
	evolutionThreshold int // 进化阈值（使用次数）
}

// NewSkillLifecycleManager 创建技能生命周期管理器
func NewSkillLifecycleManager(workspace string) *SkillLifecycleManager {
	archivePath := filepath.Join(workspace, ".skills_archive")
	statsFile := filepath.Join(workspace, ".skills_stats.json")

	return &SkillLifecycleManager{
		workspace:         workspace,
		archivePath:       archivePath,
		statsFile:         statsFile,
		stats:            make(map[string]*SkillUsageStats),
		archiveThreshold:  90,  // 90天不使用则归档
		evolutionThreshold: 10, // 使用10次后考虑进化
	}
}

// LoadStats 加载技能使用统计
func (m *SkillLifecycleManager) LoadStats() error {
	data, err := os.ReadFile(m.statsFile)
	if err != nil {
		if os.IsNotExist(err) {
			m.stats = make(map[string]*SkillUsageStats)
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &m.stats)
}

// SaveStats 保存技能使用统计
func (m *SkillLifecycleManager) SaveStats() error {
	data, err := json.MarshalIndent(m.stats, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.statsFile, data, 0644)
}

// RecordUsage 记录技能使用
func (m *SkillLifecycleManager) RecordUsage(skillName string, success bool) error {
	if m.stats == nil {
		if err := m.LoadStats(); err != nil {
			return err
		}
	}

	now := time.Now()
	stats, exists := m.stats[skillName]
	if !exists {
		stats = &SkillUsageStats{
			FirstUsedAt: now,
		}
		m.stats[skillName] = stats
	}

	stats.TotalUsage++
	stats.LastUsedAt = now

	// 计算成功率
	if stats.TotalUsage == 1 {
		stats.SuccessRate = 1.0
	} else {
		// 简单的移动平均
		currentSuccess := 1.0
		if !success {
			currentSuccess = 0.0
		}
		stats.SuccessRate = (stats.SuccessRate*float64(stats.TotalUsage-1) + currentSuccess) / float64(stats.TotalUsage)
	}

	return m.SaveStats()
}

// ShouldArchive 判断技能是否应该归档
func (m *SkillLifecycleManager) ShouldArchive(skillName string) bool {
	stats, exists := m.stats[skillName]
	if !exists {
		return false
	}

	// 90天未使用且总使用次数少于5次
	daysSinceLastUsed := time.Since(stats.LastUsedAt).Hours() / 24
	return daysSinceLastUsed >= float64(m.archiveThreshold) && stats.TotalUsage < 5
}

// ShouldEvolve 判断技能是否应该进化
func (m *SkillLifecycleManager) ShouldEvolve(skillName string) bool {
	stats, exists := m.stats[skillName]
	if !exists {
		return false
	}

	// 使用次数达到阈值且成功率大于80%
	return stats.TotalUsage >= m.evolutionThreshold && stats.SuccessRate > 0.8
}

// ArchiveSkill 归档技能
func (m *SkillLifecycleManager) ArchiveSkill(skillName string) error {
	// 创建归档目录
	if err := os.MkdirAll(m.archivePath, 0755); err != nil {
		return err
	}

	// 源技能目录
	srcDir := filepath.Join(m.workspace, skillName)
	// 目标归档目录
	dstDir := filepath.Join(m.archivePath, skillName)

	// 检查源目录是否存在
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return fmt.Errorf("skill '%s' does not exist", skillName)
	}

	// 移动到归档目录
	if err := os.Rename(srcDir, dstDir); err != nil {
		return err
	}

	// 移除统计信息
	delete(m.stats, skillName)
	if err := m.SaveStats(); err != nil {
		slog.Warn("Failed to save stats after archiving", "error", err)
	}

	slog.Info("Skill archived",
		"skill", skillName,
		"archive_path", dstDir)

	return nil
}

// GetStats 获取技能统计信息
func (m *SkillLifecycleManager) GetStats(skillName string) (*SkillUsageStats, bool) {
	if m.stats == nil {
		if err := m.LoadStats(); err != nil {
			slog.Error("Failed to load stats", "error", err)
			return nil, false
		}
	}

	stats, exists := m.stats[skillName]
	return stats, exists
}

// GetAllStats 获取所有技能统计信息
func (m *SkillLifecycleManager) GetAllStats() map[string]*SkillUsageStats {
	if m.stats == nil {
		if err := m.LoadStats(); err != nil {
			slog.Error("Failed to load stats", "error", err)
			return make(map[string]*SkillUsageStats)
		}
	}

	return m.stats
}

// SuggestEvolution 建议技能进化
func (m *SkillLifecycleManager) SuggestEvolution(skillName string) (string, error) {
	creator := NewSkillCreator(m.workspace, GetDefaultLoader())

	skill, content, err := creator.GetSkill(skillName)
	if err != nil {
		return "", err
	}

	stats, _ := m.GetStats(skillName)

	suggestion := fmt.Sprintf(`# 技能进化建议: %s

## 当前状态
- 总使用次数: %d
- 成功率: %.1f%%
- 首次使用: %s
- 最后使用: %s
- 最近30天使用: %d次

## 进化建议
基于使用统计数据，建议对技能进行以下优化：

1. **性能优化**: 根据成功率和使用频率优化执行流程
2. **扩展触发词**: 考虑添加更多触发关键词以提高技能发现率
3. **优化描述**: 更新技能描述以反映实际使用场景
4. **添加参数**: 考虑增加可配置参数以提高灵活性

## 当前技能内容
%s
`, skillName,
		stats.TotalUsage,
		stats.SuccessRate*100,
		stats.FirstUsedAt.Format("2006-01-02"),
		stats.LastUsedAt.Format("2006-01-02"),
		stats.RecentUsageCount,
		content)

	return suggestion, nil
}

// RunLifecycleCheck 运行生命周期检查
func (m *SkillLifecycleManager) RunLifecycleCheck() ([]string, error) {
	var results []string

	// 加载统计信息
	if err := m.LoadStats(); err != nil {
		return nil, err
	}

	// 检查需要归档的技能
	for skillName := range m.stats {
		if m.ShouldArchive(skillName) {
			skillPath := filepath.Join(m.workspace, skillName)
			if _, err := os.Stat(skillPath); err == nil {
				if err := m.ArchiveSkill(skillName); err != nil {
					results = append(results, fmt.Sprintf("❌ 归档失败: %s - %v", skillName, err))
				} else {
					results = append(results, fmt.Sprintf("✅ 已归档: %s", skillName))
				}
			}
		}
	}

	return results, nil
}

// GetInactiveSkills 获取不活跃技能列表
func (m *SkillLifecycleManager) GetInactiveSkills(thresholdDays int) ([]string, error) {
	if m.stats == nil {
		if err := m.LoadStats(); err != nil {
			return nil, err
		}
	}

	var inactiveSkills []string
	now := time.Now()

	for skillName, stats := range m.stats {
		daysSinceLastUsed := int(now.Sub(stats.LastUsedAt).Hours() / 24)
		if daysSinceLastUsed >= thresholdDays {
			inactiveSkills = append(inactiveSkills, fmt.Sprintf(
				"%s (未使用: %d天, 总使用: %d次)",
				skillName, daysSinceLastUsed, stats.TotalUsage))
		}
	}

	return inactiveSkills, nil
}

// GetEvolutionCandidates 获取需要进化的技能
func (m *SkillLifecycleManager) GetEvolutionCandidates() ([]string, error) {
	if m.stats == nil {
		if err := m.LoadStats(); err != nil {
			return nil, err
		}
	}

	var candidates []string

	for skillName, stats := range m.stats {
		if m.ShouldEvolve(skillName) {
			candidates = append(candidates, fmt.Sprintf(
				"%s (使用: %d次, 成功率: %.1f%%)",
				skillName, stats.TotalUsage, stats.SuccessRate*100))
		}
	}

	return candidates, nil
}
