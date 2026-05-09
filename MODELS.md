# OpenCode Go Models Guide

Comprehensive guide to OpenCode Go models with capabilities, costs, and routing recommendations.

**Source:** [OpenCode Go Documentation](https://opencode.ai/docs/go/)

## Quick Cost Comparison

> 💰 **Cost-conscious routing matters!** GLM-5.1 gives you 880 requests per 5-hour block, while Qwen3.5 Plus gives you **10,200** — that's **11.6x more requests** for the same $12 budget.

| Model            | Requests per $12 (5hr) | Cost Efficiency | Quality |
| ---------------- | ---------------------- | --------------- | ------- |
| **Qwen3.5 Plus** | **10,200**             | ★★★★★           | ★★☆☆☆   |
| **MiniMax M2.5** | **6,300**              | ★★★★★           | ★★☆☆☆   |
| **MiniMax M2.7** | **3,400**              | ★★★★☆           | ★★★☆☆   |
| **Qwen3.6 Plus** | **3,300**              | ★★★★☆           | ★★★☆☆   |
| **MiMo-V2-Omni** | **2,150**              | ★★★☆☆           | ★★★☆☆   |
| **Kimi K2.5**    | **1,850**              | ★★☆☆☆           | ★★★★☆   |
| **MiMo-V2-Pro**  | **1,290**              | ★★☆☆☆           | ★★★★☆   |
| **Kimi K2.6**    | **~1,150**             | ★☆☆☆☆           | ★★★★★   |
| **GLM-5**        | **1,150**              | ★☆☆☆☆           | ★★★★☆   |
| **GLM-5.1**      | **880**                | ☆☆☆☆☆           | ★★★★★   |

## Important: API Endpoints

⚠️ **Critical:** Not all models use the same API endpoint! oc-go-cc handles this automatically, but you should know:

| Models                                                                                                             | Endpoint                                         | Format                   |
| ------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------ | ------------------------ |
| GLM-5, GLM-5.1, Kimi K2.6, Kimi K2.5, MiMo-V2-Pro, MiMo-V2-Omni, Qwen3.5 Plus, Qwen3.6 Plus, DeepSeek V4 Pro/Flash | `https://opencode.ai/zen/go/v1/chat/completions` | OpenAI-compatible        |
| **MiniMax M2.5, MiniMax M2.7**                                                                                     | `https://opencode.ai/zen/go/v1/messages`         | **Anthropic-compatible** |

**Why this matters:** MiniMax models expect Anthropic format natively. oc-go-cc detects MiniMax models and routes them to the correct endpoint automatically without transformation. This means MiniMax models work seamlessly with Claude Code.

DeepSeek V4 Pro and Flash are OpenAI-compatible in OpenCode Go. oc-go-cc transforms Claude Code's Anthropic request into OpenAI Chat Completions format, including tools, tool results, thinking history, `reasoning_effort`, and `thinking`.

For Claude Code and OpenCode-style agent workflows, DeepSeek V4 supports max thinking mode with:

```json
{
  "model_id": "deepseek-v4-pro",
  "reasoning_effort": "max",
  "thinking": {
    "type": "enabled"
  }
}
```

Use `deepseek-v4-pro` for default, complex, thinking, and long-context routing. Use `deepseek-v4-flash` for fast, background, or subagent-style workloads.

## Cost-Conscious Routing Strategy

### Default to Cheap, Upgrade When Necessary

**Most requests should use cheap models.** Only upgrade to expensive models when:

1. **Task complexity demands it** (multi-step reasoning, architecture)
2. **You've tried cheaper models and they failed**
3. **Code quality is critical** (production code review)

### Recommended Routing

```json
{
  "models": {
    "background": {
      // Simple operations
      "model_id": "qwen3.5-plus",
      "max_tokens": 2048
    },
    "default": {
      // Better quality, moderate cost
      "model_id": "kimi-k2.6",
      "max_tokens": 4096
    },
    "long_context": {
      // Large files only
      "model_id": "minimax-m2.5",
      "context_threshold": 80000
    },
    "think": {
      // Reasoning tasks
      "model_id": "glm-5",
      "max_tokens": 8192
    },
    "complex": {
      // Complex architecture only
      "model_id": "glm-5.1",
      "max_tokens": 4096
    },
    "fast": {
      // Streaming requests (prioritize TTFT)
      "model_id": "qwen3.6-plus",
      "max_tokens": 4096
    }
  }
}
```

### Decision Tree

```
Is context > 80K tokens?
├── YES → Use MiniMax M2.5 (1M context, 6,300 req/$12)
│
Is it a complex task (architecture, refactoring, tool operations)?
├── YES → Use GLM-5.1 (880 req/$12)
│
Is it a reasoning/planning task?
├── YES → Use GLM-5 (1,150 req/$12)
│
Is it a simple background task (read file, grep, list dir, no tools)?
├── YES → Use Qwen3.5 Plus (10,200 req/$12)
│
Default → Use Kimi K2.6 (1,850 req/$12, ★★★★★) or Qwen3.6 Plus (3,300 req/$12)
```

## Detailed Model Profiles

### Budget Champions 💰

#### Qwen3.5 Plus — The Workhorse

- **Model ID:** `qwen3.5-plus`
- **Cost:** **10,200 requests per $12** (best value!)
- **Context:** ~128K tokens
- **Quality:** ★★☆☆☆ (adequate for simple tasks)
- **Best For:**
  - File reading operations
  - Directory listing
  - Grep/search
  - Simple questions
  - Bulk operations
  - Background tasks
- **When to Use:** When you need to do lots of operations cheaply

#### MiniMax M2.5 — Long Context on a Budget

- **Model ID:** `minimax-m2.5`
- **Endpoint:** **Anthropic-compatible** (`/v1/messages`)
- **Cost:** **6,300 requests per $12**
- **Context:** **~1M tokens** (1 million!)
- **Quality:** ★★☆☆☆ (acceptable)
- **Speed:** Fast
- **Best For:**
  - Very large files
  - Long conversations
  - Multi-file context
- **When to Use:** When you need 1M context but want to minimize cost
- **Note:** Uses Anthropic endpoint - oc-go-cc handles this automatically

### Balanced Models (Quality + Cost)

#### DeepSeek V4 Pro — Agentic Coding + Max Thinking

- **Model ID:** `deepseek-v4-pro`
- **Endpoint:** **OpenAI-compatible** (`/chat/completions`)
- **Context:** **~1M tokens**
- **Quality:** ★★★★★
- **Best For:**
  - Claude Code agent workflows
  - Complex implementation and debugging
  - Architecture and refactoring
  - Long-context coding tasks
  - Max thinking mode
- **Recommended Config:**

  ```json
  {
    "provider": "opencode-go",
    "model_id": "deepseek-v4-pro",
    "temperature": 0.1,
    "max_tokens": 8192,
    "reasoning_effort": "max",
    "thinking": {
      "type": "enabled"
    }
  }
  ```

#### DeepSeek V4 Flash — Fast Agent Workloads

- **Model ID:** `deepseek-v4-flash`
- **Endpoint:** **OpenAI-compatible** (`/chat/completions`)
- **Context:** **~1M tokens**
- **Quality:** ★★★★☆
- **Best For:**
  - Fast routing
  - Background tasks
  - Subagent-style work
  - Fallback for DeepSeek V4 Pro
- **Recommended Config:**

  ```json
  {
    "provider": "opencode-go",
    "model_id": "deepseek-v4-flash",
    "temperature": 0.1,
    "max_tokens": 4096,
    "reasoning_effort": "max",
    "thinking": {
      "type": "enabled"
    }
  }
  ```

#### Qwen3.6 Plus — Cost-Effective General Coding ⭐ RECOMMENDED DEFAULT

- **Model ID:** `qwen3.6-plus`
- **Cost:** **3,300 requests per $12** (3.8x more than GLM-5.1!)
- **Context:** ~128K tokens
- **Quality:** ★★★☆☆ (good enough for most tasks)
- **Speed:** Fast
- **Best For:**
  - General coding (default choice)
  - Feature implementation
  - Bug fixes
  - Refactoring
- **When to Use:** Default for cost-conscious users

#### Kimi K2.6 — Best Quality at Balanced Cost

- **Model ID:** `kimi-k2.6`
- **Cost:** **~1,850 requests per $12**
- **Context:** ~256K tokens (successor to K2.5 with improvements)
- **Quality:** ★★★★★ (excellent — successor improvements)
- **Speed:** Fast
- **Best For:**
  - Complex coding tasks
  - Code review
  - Architecture discussions
  - General-purpose default (best quality-to-cost ratio)
- **When to Use:** Default choice — better quality than K2.5 at similar cost

#### Kimi K2.5 — Quality + Reasonable Cost (Predecessor)

- **Model ID:** `kimi-k2.5`
- **Cost:** **1,850 requests per $12**
- **Context:** ~256K tokens (2x most others)
- **Quality:** ★★★★☆ (excellent)
- **Speed:** Fast
- **Best For:**
  - Complex coding tasks
  - Code review
  - Architecture discussions
  - When you need better quality than budget models
- **When to Use:** When quality matters more than maximum cost savings

### Premium Models (Use Sparingly!)

#### GLM-5 — Reasoning Specialist

- **Model ID:** `glm-5`
- **Cost:** **1,150 requests per $12** (9x more expensive than Qwen3.5 Plus!)
- **Context:** ~200K tokens
- **Quality:** ★★★★☆ (excellent)
- **Best For:**
  - Multi-step reasoning
  - Complex planning
  - Algorithm design
  - Difficult debugging
- **When to Use:** When reasoning/planning is required and budget models fail

#### GLM-5.1 — Maximum Quality

- **Model ID:** `glm-5.1`
- **Cost:** **880 requests per $12** (11.6x more expensive than Qwen3.5 Plus!)
- **Context:** ~200K tokens
- **Quality:** ★★★★★ (best available)
- **Speed:** Moderate
- **Best For:**
  - Critical architectural decisions
  - Complex multi-file refactoring
  - Production code review
  - When you need the absolute best quality
- **When to Use:** Only when cheaper models can't handle the task

## Usage Limits

OpenCode Go limits:

- **5-hour limit:** $12 of usage
- **Weekly limit:** $30 of usage
- **Monthly limit:** $60 of usage

### Cost Comparison Example

**Scenario:** You want to make 5,000 requests this month.

| Model        | Cost | Can you do it?        |
| ------------ | ---- | --------------------- |
| Qwen3.5 Plus | ~$6  | ✅ Yes, easily        |
| MiniMax M2.5 | ~$10 | ✅ Yes                |
| Qwen3.6 Plus | ~$18 | ✅ Yes                |
| Kimi K2.5    | ~$32 | ❌ Exceeds $30 weekly |
| GLM-5        | ~$52 | ❌ Exceeds limits     |
| GLM-5.1      | ~$68 | ❌ Exceeds limits     |

### Optimizing Your Usage

**Strategy 1: Tiered Approach**

```
1. Start with Qwen3.6 Plus (cheap, good quality)
2. If it fails, try Kimi K2.5 (better quality)
3. If still failing, use GLM-5 (reasoning)
4. Only for critical tasks: GLM-5.1 (premium)
```

**Strategy 2: Task-Based Selection**

```
Background ops (grep, ls, cat) → Qwen3.5 Plus
General coding → Qwen3.6 Plus or Kimi K2.5
Complex features → Kimi K2.5
Architecture/Planning → GLM-5
Critical review → GLM-5.1 (rarely)
```

## Fallback Chains for Cost Efficiency

```json
{
  "fallbacks": {
    "background": [
      { "model_id": "qwen3.6-plus" },
      { "model_id": "minimax-m2.5" }
    ],
    "long_context": [{ "model_id": "minimax-m2.5" }],
    "default": [{ "model_id": "mimo-v2-pro" }, { "model_id": "qwen3.6-plus" }],
    "think": [{ "model_id": "kimi-k2.6" }],
    "complex": [{ "model_id": "glm-5" }],
    "fast": [{ "model_id": "qwen3.5-plus" }, { "model_id": "minimax-m2.5" }]
  }
}
```

**Rule of thumb:** If a task succeeds with a cheap model, it doesn't need an expensive one. Only fall back to expensive models when necessary.

## Quick Reference

| Task Type             | Recommended  | Cost (req/$12) | Fallback     |
| --------------------- | ------------ | -------------- | ------------ |
| Read file, ls, grep   | Qwen3.5 Plus | 10,200         | Qwen3.6 Plus |
| General coding        | Qwen3.6 Plus | 3,300          | Kimi K2.5    |
| Complex features      | Kimi K2.6    | 1,850          | Kimi K2.5    |
| Long context (>80K)   | MiniMax M2.5 | 6,300          | MiniMax M2.7 |
| Reasoning/planning    | GLM-5        | 1,150          | Kimi K2.5    |
| Critical architecture | GLM-5.1      | 880            | GLM-5        |
| Bulk operations       | Qwen3.5 Plus | 10,200         | MiniMax M2.5 |

## Cost-Saving Tips

1. **Use Qwen3.6 Plus as default** — 3,300 req/$12 is plenty for most tasks
2. **Reserve GLM-5.1 for critical tasks only** — 880 req/$12 drains budget fast
3. **Use Qwen3.5 Plus for simple operations** — 10,200 req/$12 is unbeatable
4. **MiniMax M2.5 for long context** — 6,300 req/$12 with 1M context is amazing value
5. **Monitor your usage** in the [OpenCode console](https://opencode.ai/auth)

## Complexity Analysis System

oc-go-cc automatically analyzes each request's complexity using 6 dimensions to route it to the appropriate model. This ensures cheap models handle simple tasks while expensive models are reserved for genuinely complex work.

### Complexity Levels

| Total Score | Level | Recommended Model | Scenario |
|-------------|-------|------------------|----------|
| 0~40 | **simple** | Qwen3.5 Plus | `fast` |
| 41~60 | **medium** | Kimi K2.6 | `default` |
| 61~80 | **complex** | GLM-5 | `think` |
| 81~100 | **extreme** | GLM-5.1 | `complex` |

### 6 Scoring Dimensions

| Dimension | Weight | Scoring Rules |
|----------|--------|---------------|
| **Token Usage** | 30 pts | >70%→30, 50%~70%→20, 25%~50%→10, <25%→0 |
| **Constraints** | 20 pts | ≥15→20, 8~14→13, 4~7→7, 1~3→3, 0→0 |
| **Memory Load** | 10 pts | ≥8→10, 4~7→6, 1~3→3, 0→0 |
| **File Scope** | 15 pts | >15/>800→15, 8~15/400~800→10, 3~7/100~400→5, <3/<100→0 |
| **Conversation** | 10 pts | >15/>12→10, 10~15/8~12→7, 5~9/4~7→3, <5/<4→0 |
| **Logic Steps** | 15 pts | >12→15, 8~12→10, 4~7→5, <4→0 |

### Constraint Keywords (Strict Detection)

**Chinese:** `必须`, `禁止`, `遵循`, `不得`, `保证`  
**English:** `must not`, `shall not`, `forbidden`, `never use/modify/delete/execute`, `must implement/follow/use/adhere/ensure/comply`

Constraints are only counted when they appear as explicit restriction patterns, preventing false positives from casual conversation.

### Conversation Scoring (Turns Only)

Conversation scoring is based on **user turns** only, not tool count. Tool-heavy requests (like Claude Code with 80+ tools) won't automatically get high conversation scores.

### Routing Priority

```
1. Long Context (> threshold tokens) → long_context scenario
2. Model Tier (sonnet → think, haiku → fast)
3. Simple Read-Only Tools (explore, search, grep, read, ls, cat, etc.) → fast
4. Simple Message Patterns (list directory, what is, cat file, etc.) → fast
5. Complex Write Tools (write, edit, execute, delete, bash, etc.) → default/think/complex
6. Complexity Analysis → routes based on score:
   - 0~40 simple → fast (if no write tools)
   - 41~60 medium → default
   - 61~80 complex → think
   - 81~100 extreme → complex
```

### Simple Read-Only Patterns

**Tools:** `explore`, `search`, `grep`, `read`, `view`, `cat`, `ls`, `list`, `find`, `fetch`, `get`, `query`, `inspect`

**Message Patterns:** `list directory`, `ls -`, `show file`, `view file`, `cat file`, `what is this`, `tell me about`, `check status`

### Complex Write Patterns

**Tools:** `write`, `edit`, `delete`, `remove`, `execute`, `run`, `bash`, `shell`, `create`, `modify`, `update`

Requests with any write tools will route to at least `default` scenario, regardless of complexity score.

### Example Log Output

```
routing request scenario=complex model=glm-5 tokens=9795 scenario_reason="complexity score 78 (complex): token_usage=20/30, constraints=13/20, memory=6/10, files=10/15, conversation=7/10, logic_steps=15/15 (Recommand GLM-5)"
```

## See Also

- [OpenCode Go Documentation](https://opencode.ai/docs/go/)
- [oc-go-cc Configuration](../configs/config.example.json)
- [README.md](../README.md) for setup instructions
