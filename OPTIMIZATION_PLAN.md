# 望舒项目优化计划

## 优化目标

1. **精简Agent上下文内容**
   - 移除不必要的心跳机制上下文
   - 优化SystemPrompt，减少冗余信息
   - 精简Profile文件加载逻辑

2. **增加自我学习能力**
   - 实现基于Go的技能创建系统
   - 支持动态学习和保存技能
   - 类似hermes的技能管理机制

3. **清理冗余内置技能**
   - 评估并移除不常用的内置技能
   - 保留核心必要技能

4. **纯Go实现**
   - 确保不依赖Node.js、Python等外部环境
   - 使用Go标准库和可用包

## 优化详情

### 1. 上下文优化

#### 1.1 移除心跳相关上下文
- 当前状态：已注释，但代码中仍有引用
- 优化方案：彻底移除相关代码和常量

#### 1.2 精简SystemPrompt
**当前问题**：
- 提示词过长（~46行）
- 包含重复说明
- Cron表达式计算说明过于详细

**优化方案**：
- 压缩到核心内容（~20行）
- 移除冗余的Cron计算说明（简化为"使用绝对时间"）
- 保留关键的行为准则

#### 1.3 精简Profile加载
**当前加载文件**：
- AGENTS.md
- BOOTSTRAP.md
- HEARTBEAT.md（已注释）
- IDENTITY.md（已注释）
- SOUL.md（已注释）
- TOOLS.md（已注释）
- USER.md（已注释）
- MEMORY.md（已注释）
- SPRITE.md

**优化方案**：
- 只保留AGENTS.md和BOOTSTRAP.md
- 移除其他不必要的profile文件
- 简化Sprite逻辑（仅在Live2D启用时加载）

### 2. 自我学习系统

#### 2.1 技能创建工具
**功能**：
- 基于对话历史自动生成技能
- 支持技能验证和优化
- 纯Go实现技能解析

**实现方案**：
```go
type SkillCreator struct {
    agent *Agent
    skillLoader *skills.Loader
}

// 方法
- CreateSkillFromConversation() // 从对话创建技能
- ValidateSkill() // 验证技能完整性
- SaveSkill() // 保存技能到文件系统
- ListSkills() // 列出所有技能
- UpdateSkill() // 更新技能
- DeleteSkill() // 删除技能
```

#### 2.2 技能管理工具
**功能**：
- 搜索和检索技能
- 技能分类和标签
- 技能使用统计
- 技能依赖管理

### 3. 技能清理

#### 3.1 内置技能评估
**保留核心技能**：
- `feishu_create` - 飞书集成
- `skill-creator` - 技能创建
- 基础文件操作相关

**移除/可选技能**：
- `algorithmic-art` - 算法艺术（不常用）
- `webapp-testing` - Web应用测试（特定场景）
- `web-artifacts-builder` - Web构建器（特定场景）
- `theme-factory` - 主题工厂（不常用）
- `slack-gif-creator` - Slack GIF（特定平台）
- `canvas-design` - Canvas设计（不常用）
- `brand-guidelines` - 品牌指南（特定场景）

### 4. 纯Go实现策略

#### 4.1 技能执行引擎
- 不依赖外部解释器
- 使用Go模板引擎
- 内置常用工具函数

#### 4.2 文件处理
- 使用标准库处理Markdown
- 使用gopkg.in/yaml.v3处理YAML
- 避免依赖外部CLI工具

## 实施步骤

### Phase 1: 上下文优化
1. 清理`pkg/constant/prompts.go`中的冗余提示词
2. 优化`internal/agent/agent_messages.go`中的上下文构建逻辑
3. 移除心跳相关代码

### Phase 2: 技能清理
1. 评估当前内置技能
2. 移除不常用技能
3. 更新技能加载器

### Phase 3: 自我学习系统
1. 实现SkillCreator
2. 添加技能管理工具
3. 集成到Agent系统

### Phase 4: 测试和优化
1. 测试上下文大小变化
2. 测试新技能系统
3. 性能优化

## 预期效果

### 上下文优化效果
- **Token使用减少**: 预计减少30-40%的系统提示词Token
- **响应速度提升**: 减少上下文传输时间
- **成本降低**: 减少API调用成本

### 自学习能力效果
- **技能动态扩展**: 支持运行时学习新技能
- **用户自定义**: 用户可以创建和分享技能
- **社区生态**: 支持技能市场概念

### 纯Go实现效果
- **零依赖**: 不需要安装Node.js、Python等
- **跨平台**: 更好的跨平台支持
- **部署简单**: 单一二进制文件即可运行

## 风险评估

### 低风险
- 上下文精简：不影响核心功能
- 技能清理：移除不常用技能

### 中风险
- 自我学习系统：需要充分测试
- 新工具集成：可能影响现有功能

### 缓解措施
- 分阶段实施，逐步验证
- 保留回滚机制
- 充分测试每个阶段

## 时间估算

- Phase 1: 1-2小时
- Phase 2: 30分钟
- Phase 3: 2-3小时
- Phase 4: 1-2小时

总计：4-7小时
