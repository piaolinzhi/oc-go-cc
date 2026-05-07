# 多 Provider 支持 - 验证检查清单

## 配置结构验证
- [ ] 配置文件支持多个 provider 定义
- [ ] 每个 provider 可以配置独立的 base URL
- [ ] 每个 provider 可以配置独立的 API key
- [ ] API key 支持环境变量引用（如 `$OPENAI_API_KEY`）
- [ ] 每个 provider 可以配置 endpoint type
- [ ] 每个 provider 可以配置权重
- [ ] 支持配置负载均衡策略（轮询、随机、加权轮询）
- [ ] 支持配置故障转移参数（失败率阈值、恢复时间）
- [ ] 模型配置可以指定 provider 名称
- [ ] 旧格式配置文件（无 providers 字段）仍然可用

## Provider 接口验证
- [ ] Provider 接口定义完整
- [ ] 可以注册新的 provider
- [ ] 可以根据名称获取 provider
- [ ] 未注册的 provider 返回错误
- [ ] 默认 provider 自动注册

## Client 层验证
- [ ] 请求正确路由到指定 provider
- [ ] 不同 provider 使用不同的 base URL
- [ ] 硬编码的模型判断已移除
- [ ] 支持 OpenAI 和 Anthropic 格式的 endpoint

## Transformer 验证
- [ ] Transformer 接口定义完整
- [ ] 可以注册自定义 transformer
- [ ] 根据 provider 选择正确的 transformer
- [ ] 默认 transformer 行为正常

## 负载均衡验证
- [ ] 轮询策略正确按顺序分配请求
- [ ] 随机策略随机分配请求
- [ ] 加权轮询策略按权重分配请求
- [ ] 负载均衡器线程安全

## 故障转移验证
- [ ] provider 失败时自动切换到备用 provider
- [ ] 故障恢复后自动重新加入
- [ ] 失败率阈值配置有效
- [ ] 健康检查机制正常工作

## 模型路由验证
- [ ] 模型正确绑定到指定 provider
- [ ] fallback 链可以跨 provider
- [ ] 路由逻辑性能不受影响

## 向后兼容性验证
- [ ] 现有测试全部通过
- [ ] 旧配置格式仍然有效
- [ ] 现有 API 行为保持不变
- [ ] 现有 `opencode_go` 配置块自动转换为 provider
- [ ] 模型已有的 `provider` 字段可以直接使用

## 文档验证
- [ ] 配置示例更新
- [ ] README 文档更新
- [ ] 配置说明清晰易懂