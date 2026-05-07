# 多 Provider 支持 - 实现计划

## [ ] Task 1: 扩展配置结构，支持多 provider、负载均衡和 API key 管理
- **Priority**: P0
- **Depends On**: None
- **Description**: 
  - 扩展 `Config` 结构体，添加 `Providers` 字段、负载均衡策略配置和 API key 管理配置
  - 创建 `ProviderConfig` 结构体，包含 name、base URL、API key（支持环境变量引用）、endpoint type、transformer 配置、权重等
  - 添加 `LoadBalancerConfig` 结构体，支持 round_robin、random、weighted_round_robin 策略
  - 更新配置加载逻辑，支持新的配置格式和环境变量解析
  - **向后兼容**: 保留现有的 `opencode_go` 配置块，自动转换为 provider 配置
- **Acceptance Criteria Addressed**: AC-1, AC-4, AC-5, AC-7, AC-8, AC-9
- **Test Requirements**:
  - `programmatic` TR-1.1: 加载包含多个 provider 的配置文件不报错
  - `programmatic` TR-1.2: 加载旧格式配置文件（无 providers 字段）不报错
  - `programmatic` TR-1.3: 加载包含负载均衡配置的文件不报错
  - `programmatic` TR-1.4: 正确解析 API key 环境变量引用
  - `programmatic` TR-1.5: 现有 `opencode_go` 配置块自动转换为 provider

## [ ] Task 2: 创建 Provider 接口和注册机制
- **Priority**: P0
- **Depends On**: Task 1
- **Description**: 
  - 创建 `Provider` 接口，定义 provider 的核心方法
  - 创建 provider 注册中心，管理所有 provider
  - 实现默认的 OpenCode Go provider
- **Acceptance Criteria Addressed**: AC-2
- **Test Requirements**:
  - `programmatic` TR-2.1: 可以注册和获取 provider
  - `programmatic` TR-2.2: 未注册的 provider 返回错误

## [ ] Task 3: 重构 Client 层，支持多 provider
- **Priority**: P0
- **Depends On**: Task 2
- **Description**: 
  - 创建通用的 HTTP client，支持根据 provider 配置选择 endpoint
  - 修改 `OpenCodeClient` 支持多 provider 配置
  - 移除硬编码的 `IsAnthropicModel` 判断，改为基于 provider 配置
- **Acceptance Criteria Addressed**: AC-2
- **Test Requirements**:
  - `programmatic` TR-3.1: 请求正确路由到指定 provider
  - `programmatic` TR-3.2: 不同 provider 使用不同的 base URL

## [ ] Task 4: 实现 provider 级别的 transformer
- **Priority**: P1
- **Depends On**: Task 2
- **Description**: 
  - 创建 transformer 接口，定义请求/响应转换方法
  - 创建 transformer 注册机制
  - 修改消息处理逻辑，根据 provider 选择 transformer
- **Acceptance Criteria Addressed**: AC-3
- **Test Requirements**:
  - `programmatic` TR-4.1: 不同 provider 使用不同的 transformer
  - `programmatic` TR-4.2: 未配置 transformer 的 provider 使用默认 transformer

## [ ] Task 5: 实现负载均衡器
- **Priority**: P0
- **Depends On**: Task 2
- **Description**: 
  - 创建负载均衡器接口，定义选择 provider 的方法
  - 实现轮询（round_robin）策略
  - 实现随机（random）策略
  - 实现加权轮询（weighted_round_robin）策略
  - 集成到 provider 选择逻辑中
- **Acceptance Criteria Addressed**: AC-4, AC-5
- **Test Requirements**:
  - `programmatic` TR-5.1: 轮询策略正确按顺序分配请求
  - `programmatic` TR-5.2: 随机策略随机分配请求
  - `programmatic` TR-5.3: 加权轮询策略按权重分配请求

## [ ] Task 6: 实现 provider 故障转移
- **Priority**: P0
- **Depends On**: Task 5
- **Description**: 
  - 创建 provider 健康检查机制
  - 实现故障检测（基于请求失败率）
  - 实现自动故障转移到备用 provider
  - 实现故障恢复后自动重新加入
- **Acceptance Criteria Addressed**: AC-6
- **Test Requirements**:
  - `programmatic` TR-6.1: 主 provider 失败时自动切换到备用
  - `programmatic` TR-6.2: 故障恢复后自动重新加入
  - `programmatic` TR-6.3: 失败率阈值配置有效

## [ ] Task 7: 更新模型路由和 fallback 逻辑
- **Priority**: P1
- **Depends On**: Task 6
- **Description**: 
  - 修改模型路由逻辑，支持跨 provider 的 fallback
  - 更新 `ModelConfig`，添加 `ProviderName` 字段
  - 修改 fallback 配置，支持指定 provider
- **Acceptance Criteria Addressed**: AC-2
- **Test Requirements**:
  - `programmatic` TR-7.1: 模型正确绑定到指定 provider
  - `programmatic` TR-7.2: fallback 链可以跨 provider

## [ ] Task 8: 更新配置示例和文档
- **Priority**: P2
- **Depends On**: Task 1
- **Description**: 
  - 更新 `config.example.json`，添加多 provider 和负载均衡配置示例
  - 更新 README.md，说明如何配置多 provider 和负载均衡策略
- **Acceptance Criteria Addressed**: AC-1
- **Test Requirements**:
  - `human-judgment` TR-8.1: 配置示例清晰易懂
  - `human-judgment` TR-8.2: 文档说明完整准确

## [ ] Task 9: 测试和验证
- **Priority**: P0
- **Depends On**: All
- **Description**: 
  - 运行现有测试，确保向后兼容
  - 添加新的单元测试，覆盖多 provider、负载均衡和故障转移功能
- **Acceptance Criteria Addressed**: All
- **Test Requirements**:
  - `programmatic` TR-9.1: 所有现有测试通过
  - `programmatic` TR-9.2: 新的 provider 测试通过
  - `programmatic` TR-9.3: 负载均衡策略测试通过
  - `programmatic` TR-9.4: 故障转移测试通过