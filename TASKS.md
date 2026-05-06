# 靶身高 App API 开发任务清单

> 基于问卷设计文档，对照当前代码状态，逐项列出已完成/待完成工作。
> 规则：**做好一个打一个勾**，避免一次性改动过大导致会话中断。

---

## 一、数据模型层 (models)

### 1.1 环境问卷评估模型
- [x] `EnvironmentAssessment` 模型已定义（含5个模块原始JSON、计算得分、预测结果、行动计划）
- [ ] **待完善**：当前模型用 `NutritionRaw` 等5个字符串字段存储JSON，如需要按具体问题拆分存储，需新增子表或扩展JSON结构

### 1.2 预警模型
- [x] `HeightAlert` 模型已定义（含类型、等级、标题、描述、维度、指标值等）
- [x] 预警相关常量已定义（7种预警类型）

### 1.3 宝宝/记录模型扩展
- [x] `Child` 新增 `Region`, `GrowthStage`, `StageConfirmedAt`, `LastHeightChangeDate`
- [x] `GrowthRecord` 新增 `BoneAge`, `BoneAgeSource`, `BoneAgeDiff`

---

## 二、DTO / 请求响应结构 (dto)

### 2.1 环境问卷请求DTO
- [x] `CreateEnvironmentAssessmentRequest` 已定义（含5个模块）
- [x] `NutritionRequest` 已定义（6个评分字段）
- [x] `SleepRequest` 已定义（5个评分字段）
- [x] `ExerciseRequest` 已定义（4个评分字段）
- [x] `HealthRequest` 已定义（3个评分字段）
- [x] `MentalRequest` 已定义（3个评分字段）

### 2.2 环境问卷响应DTO
- [x] `EnvironmentAssessmentResponse` 已定义
- [x] `WeeklyActionPlan` + `ActionPlanItem` 已定义
- [x] `EnvironmentAssessmentHistoryResponse` + `EnvironmentAssessmentSummary` 已定义
- [x] **缺失类型已补充**：`ModuleScore`, `KhamisRocheResult`, `QuantitativeGeneticsResult`, `ComprehensivePredictionResult`, `AgeLayeredWeights`
  > 实际发现这些类型已在 `standards.go` 中定义，删除了 `dto.go` 中的重复声明。

### 2.3 预警DTO
- [x] `AlertResponse`, `AlertListRequest`, `AlertListResponse`, `AlertSummaryResponse`
- [x] `SetGrowthStageRequest`, `DismissAlertRequest`
- [x] `BoneAgeSummary`

### 2.4 生长速度与靶身高DTO
- [x] `GrowthVelocityResponse` 已定义
- [x] `TargetHeightComparisonResponse` 已定义

---

## 三、数据库迁移 (main.go)

- [x] `HeightAlert` 已加入 `autoMigrate`
- [x] `EnvironmentAssessment` 已加入 `autoMigrate`

---

## 四、路由层 (handler)

### 4.1 预警路由（已完成）
- [x] `POST /children/:id/growth-stage` — 设置生长阶段
- [x] `GET /children/:id/alerts` — 获取宝宝预警列表
- [x] `POST /alerts/:alertId/read` — 标记预警已读
- [x] `POST /alerts/:alertId/dismiss` — 忽略预警
- [x] `GET /alerts/summary` — 获取预警摘要

### 4.2 环境问卷评估路由（已完成）
- [x] `POST /children/:id/environment-assessment` — 提交环境问卷评估
- [x] `GET /children/:id/environment-assessment/latest` — 获取最新评估结果
- [x] `GET /children/:id/environment-assessment/history` — 获取评估历史列表

### 4.3 靶身高与生长速度路由（已完成）
- [x] `GET /children/:id/target-height-comparison` — 靶身高综合分析
- [x] `GET /children/:id/growth-velocity` — 生长速度监测

---

## 五、业务逻辑层 (service)

### 5.1 预警系统（已完成）
- [x] `SetGrowthStage` — 设置生长阶段
- [x] `GetChildAlerts` — 获取预警列表
- [x] `MarkAlertRead` — 标记已读
- [x] `DismissAlert` — 忽略预警
- [x] `GetAlertsSummary` — 获取摘要
- [x] `buildBoneAgeSummary` — 构建骨龄摘要
- [x] `evaluateAndSaveAlerts` — 运行预警引擎并保存（已在创建记录时调用）

### 5.2 环境问卷评估（**5个方法均未实现**）
- [ ] `CreateEnvironmentAssessment`
  - 参数校验（宝宝归属权）
  - 读取宝宝信息（性别、父母身高）
  - 计算遗传靶身高（Khamis-Roche 公式）
  - 计算各模块得分和加权总分
  - 计算环境增量（总分 × 0.2 cm，上限10 cm）
  - 计算预测身高 = 遗传靶身高 + 环境增量
  - 判定干预分区（high/medium/low）
  - 生成个性化行动计划（取最薄弱模块的3个优先行动）
  - 保存到数据库（包括原始JSON和计算结果）
  - 返回完整响应（需补充缺失的DTO类型）

- [ ] `GetLatestEnvironmentAssessment`
  - 查询该宝宝最新一条评估记录
  - 反序列化JSON为响应结构
  - 返回 `EnvironmentAssessmentResponse`

- [ ] `GetEnvironmentAssessmentHistory`
  - 分页查询历史评估记录
  - 返回摘要列表

### 5.3 靶身高对比（**未实现**）
- [ ] `GetTargetHeightComparison`
  - 计算遗传靶身高（Khamis-Roche）
  - 定量遗传学分析结果
  - 获取最新环境评估预测结果
  - 当前身高百分位查询
  - 遗传潜力达成状态判定
  - 生长速度数据（可复用 GetGrowthVelocity 逻辑）

### 5.4 生长速度监测（**未实现**）
- [ ] `GetGrowthVelocity`
  - 取最近 N 个月的身高记录（默认6个月或12个月）
  - 计算年生长速度（cm/年）
  - 与同龄期望最低值对比
  - 判定状态（正常/偏慢/过慢）
  - 返回预警等级和行动建议

---

## 六、预警引擎 (internal/alert)

- [x] `Engine.Evaluate` — 评估所有预警维度
- [x] `Engine.SaveAlerts` — 保存预警到数据库
- [x] `Engine.GetActiveAlerts` — 获取活跃预警
- [x] `Engine.GetSummary` — 获取预警摘要
- [x] `Engine.GetChildAlertList` — 分页获取预警列表
- [x] `Engine.MarkAlertRead` — 标记已读
- [x] `Engine.DismissAlert` — 忽略预警
- [x] 6种预警检查逻辑：靶身高差距、区域偏差、骨龄超前/延迟、生长停滞、速度偏慢、百分位下降
- [x] 预警去重逻辑

---

## 七、问卷设计文档 vs 当前代码差距

> 当前代码采用**简化模式**：前端计算好每个模块的得分（0-2/0-3等整数），传给后端。后端只做汇总计算和行动计划生成。
> 文档中 Q7-Q26 的详细问题（如"昨天吃了多少种食物""睡前是否用电子设备"等）由前端负责展示和转化为模块得分。

| 模块 | 文档问题数 | 当前DTO字段数 | 说明 |
|------|-----------|--------------|------|
| 营养 | Q7-Q12 (6题) | 6个字段 | 一一对应，简化合理 |
| 睡眠 | Q13-Q17 (5题) | 5个字段 | 一一对应 |
| 运动 | Q18-Q21 (4题) | 4个字段 | 一一对应 |
| 健康 | Q22-Q23 (2题) | 3个字段 | 多一个 `medication_safety` |
| 心理 | Q24-Q26 (3题) | 3个字段 | 一一对应 |

**结论**：DTO 结构与文档问卷结构基本匹配，无需新增字段。前端负责将具体问题的答案映射为模块得分传给后端。

---

## 八、开发顺序建议（降低风险）

### 阶段1：补齐编译基础（必须先做，否则无法编译）
1. 补充5个缺失的DTO类型（ModuleScore 等）
2. `main.go` 添加 `EnvironmentAssessment` 到 AutoMigrate
3. Docker 编译验证

### 阶段2：环境问卷评估（核心功能）
4. 实现 `CreateEnvironmentAssessment` 服务方法
5. 实现 `GetLatestEnvironmentAssessment` 服务方法
6. 实现 `GetEnvironmentAssessmentHistory` 服务方法
7. 注册环境问卷路由（3个端点）
8. Docker 编译验证 + 简单测试

### 阶段3：靶身高与生长速度
9. 实现 `GetGrowthVelocity`
10. 实现 `GetTargetHeightComparison`
11. 注册对应路由（2个端点）
12. Docker 编译验证

### 阶段4：整合与测试
13. 确认 `CreateRecord` 时正确触发预警评估
14. 端到端API测试
15. 提交代码

---

## 九、与问卷文档对应的建议文案（后端需返回）

当前 `ActionPlanItem` 结构支持返回标题、描述、原因、开始方式、难度。需要后端根据得分最低模块，从以下建议库中抽取3条：

- **营养模块建议库**（基于文档Q7-Q12的达成建议）
- **睡眠模块建议库**（基于文档Q13-Q17的达成建议）
- **运动模块建议库**（基于文档Q18-Q21的达成建议）
- **健康模块建议库**（基于文档Q22-Q23的达成建议）
- **心理模块建议库**（基于文档Q24-Q26的达成建议）

> 建议将各模块的建议文案硬编码为常量数组，按得分区间匹配返回。

---

*任务清单创建时间：2026-05-03*
*当前代码分支：main*
*未提交修改：6个文件修改 + 2个新增文件（internal/alert/）*
