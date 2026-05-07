# 多 Provider 支持 - 产品需求文档

## Overview
- **Summary**: 扩展 oc-go-cc 项目，支持多个不同的 provider（如 OpenAI、Anthropic、Azure OpenAI 等），每个 provider 可以有独立的配置（base URL、API key）和自定义的 transformer。
- **Purpose**: 当前项目仅支持 OpenCode Go 一个 provider，限制了用户选择其他 AI 服务的灵活性。通过支持多 provider，可以让用户灵活配置不同的后端服务。
- **Target Users**: 需要使用多个 AI 服务提供商的开发者和企业用户。

## Goals
- 支持配置多个 provider，每个 provider 有独立的 base URL 和认证信息
- 每个 provider 可以配置自定义的 transformer（请求/响应转换规则）
- 模型可以绑定到特定的 provider
- 支持 provider 级别的负载均衡（轮询、随机、加权轮询等策略）
- 支持 provider 级别的故障转移（自动切换到备用 provider）
- 支持 provider 级别的 API key 管理（独立 API key、环境变量引用）
- 保持向后兼容性，现有配置格式仍然可用

## Non-Goals (Out of Scope)
- 不支持动态 provider 配置（运行时添加/删除）
- 不实现新的 transformer，仅提供扩展机制

## Background & Context
当前项目结构：
- `internal/client/`: 处理与 OpenCode Go API 的通信
- `internal/transformer/`: 处理 Anthropic ↔ OpenAI 格式转换
- `internal/config/`: 配置加载和管理

**现状分析**：
- 模型配置中已有 `provider` 字段（全部设置为 `opencode-go`）
- 但代码中只有一个硬编码的 `opencode_go` provider 配置
- 模型的 `provider` 字段实际上没有被使用
- provider 选择逻辑是通过模型 ID 硬编码判断的（`IsAnthropicModel`）

需要重构为基于配置的多 provider 支持，同时保持向后兼容性。

## Functional Requirements
- **FR-1**: 支持配置多个 provider，每个 provider 包含 name、base URL、API key、endpoint type 等信息
- **FR-2**: 模型配置可以指定使用哪个 provider
- **FR-3**: 每个 provider 可以配置自定义的 transformer 规则
- **FR-4**: 保持向后兼容性，未指定 provider 的模型默认使用现有行为

## Non-Functional Requirements
- **NFR-1**: 配置格式清晰，易于理解和维护
- **NFR-2**: provider 选择逻辑高效，不影响请求处理性能
- **NFR-3**: 错误处理完善，provider 配置错误时有清晰的错误信息

## Constraints
- **Technical**: Go 语言，保持现有代码风格和架构
- **Dependencies**: 不引入新的第三方库

## Assumptions
- 用户了解不同 provider 的 API 格式差异
- 用户负责正确配置 transformer 规则

## Acceptance Criteria

### AC-1: 配置多个 provider
- **Given**: 配置文件中定义了多个 provider
- **When**: 启动服务
- **Then**: 服务成功加载所有 provider 配置
- **Verification**: `programmatic`

### AC-2: 模型绑定到 provider
- **Given**: 模型配置指定了 provider
- **When**: 发送请求到该模型
- **Then**: 请求被路由到正确的 provider
- **Verification**: `programmatic`

### AC-3: 自定义 transformer
- **Given**: provider 配置了自定义 transformer
- **When**: 请求经过该 provider
- **Then**: 使用 provider 的 transformer 进行格式转换
- **Verification**: `programmatic`

### AC-4: 轮询负载均衡
- **Given**: 配置了多个 provider 和轮询策略
- **When**: 连续发送多个请求
- **Then**: 请求按顺序轮询分配到各个 provider
- **Verification**: `programmatic`

### AC-5: 加权轮询负载均衡
- **Given**: 配置了多个 provider 和加权轮询策略
- **When**: 连续发送多个请求
- **Then**: 请求按权重比例分配到各个 provider
- **Verification**: `programmatic`

### AC-6: provider 故障转移
- **Given**: 主 provider 不可用
- **When**: 发送请求
- **Then**: 请求自动切换到备用 provider
- **Verification**: `programmatic`

### AC-7: 向后兼容
- **Given**: 使用现有格式的配置文件（无 providers 字段）
- **When**: 启动服务
- **Then**: 服务正常运行，使用默认 provider
- **Verification**: `programmatic`

## Open Questions
- [ ] 是否需要支持同一个 provider 的多个实例（如不同区域的 OpenAI API）？