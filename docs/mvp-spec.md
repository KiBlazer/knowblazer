# Knowblazer MVP 规格文档

日期：2026-04-29

## 1. 目标

本文把 `requirements.md` 中的 P0 范围细化为可实现、可测试的 MVP 规格。

MVP 只验证一条核心路径：

```text
初始化私有工程记忆库
  ↓
保存一条 Markdown 经验
  ↓
扫描敏感信息
  ↓
提升为长期工程记忆
  ↓
为一次 AI Coding 任务生成 recall 上下文
```

MVP 不追求自动化程度最高，也不追求智能检索。它优先验证：私有 Git + Markdown + 分层记忆 + 安全召回这条路径是否真实有用。

## 2. MVP 命令

P0 命令包括：

```bash
knowblazer init [path]
knowblazer capture <file> [--repo <path>]
knowblazer scan <path> [--repo <path>]
knowblazer promote <file> --to <target> [--repo <path>]
knowblazer recall --task <text> [--repo <path>] [--project <name>] [--output <file>]
```

`--repo` 用来显式指定 Knowblazer 记忆库路径。没有传入时，工具按以下顺序寻找：

1. 当前目录是否是 Knowblazer 记忆库。
2. 当前目录的父级目录是否包含 Knowblazer 记忆库标记。
3. 环境变量 `KNOWBLAZER_REPO`。
4. 用户主目录下的默认路径，例如 `~/knowblazer-notes`。

具体默认路径可在实现阶段确认，但 MVP 需要保留这个查找顺序。

## 3. 记忆库结构

`init` 默认生成以下结构：

```text
knowblazer-notes/
├── .knowblazer/
│   └── config.json
├── AI-SETUP.md
├── inbox/
├── daily/
├── profile/
│   ├── preferences.md
│   └── decision-principles.md
├── projects/
├── experience/
│   ├── deployment/
│   ├── frontend/
│   ├── backend/
│   ├── ai-tools/
│   └── operations/
├── system/
│   ├── memory-policy.md
│   └── privacy-policy.md
├── quarantine/
└── recall/
```

`.knowblazer/config.json` 是仓库标记和基础配置文件。它不保存敏感信息。

推荐初始内容：

```json
{
  "version": 1,
  "created_by": "knowblazer",
  "memory_repo": true
}
```

## 4. 文件元信息

Knowblazer 生成或移动的 Markdown 文件应尽量包含简单 front matter，便于人读和工具处理。

推荐格式：

```markdown
---
title: "Deploy frontend rollback lesson"
type: "inbox"
status: "candidate"
source: "/path/to/original.md"
captured_at: "2026-04-29T12:30:00+08:00"
---

# Deploy frontend rollback lesson

...
```

MVP 不要求支持复杂 schema，但应保证：

- 标题可读。
- 来源可追踪。
- 状态清楚。
- 时间明确。

## 5. `init`

### 5.1 用法

```bash
knowblazer init [path]
```

示例：

```bash
knowblazer init ~/knowblazer-notes
```

### 5.2 行为

`init` 应：

- 创建记忆库根目录。
- 创建第一版最小目录结构。
- 创建 `.knowblazer/config.json`。
- 创建 `AI-SETUP.md`。
- 创建 `system/memory-policy.md`。
- 创建 `system/privacy-policy.md`。
- 创建 `profile/preferences.md`。
- 创建 `profile/decision-principles.md`。

### 5.3 幂等性

重复执行 `init` 时：

- 不覆盖用户已经修改过的文件。
- 缺失的目录可以补齐。
- 缺失的模板文件可以创建。
- 如果目标目录不是 Knowblazer 记忆库但非空，应提示用户确认或失败退出。MVP 可以选择失败退出，避免误写。

### 5.4 输出

成功输出应包括：

```text
Initialized Knowblazer memory repo: <path>
Next steps:
  1. Edit AI-SETUP.md
  2. Capture a lesson with: knowblazer capture <file> --repo <path>
  3. Generate recall with: knowblazer recall --task "<task>" --repo <path>
```

### 5.5 验收

- 空目录执行后生成完整最小结构。
- 重复执行不会覆盖已有内容。
- 非空且非 Knowblazer 目录不会被静默改写。

## 6. `scan`

### 6.1 用法

```bash
knowblazer scan <path> [--repo <path>]
```

`path` 可以是文件或目录。

### 6.2 扫描范围

MVP 至少识别以下高风险模式：

- Private key block，例如 `-----BEGIN PRIVATE KEY-----`。
- 常见 token 字段，例如 `token=...`、`api_key=...`、`secret=...`。
- password 字段，例如 `password=...`。
- 数据库连接串，例如 `mysql://`、`postgres://`、`mongodb://`。
- `.env` 风格的密钥变量，例如 `AWS_SECRET_ACCESS_KEY=...`。

MVP 可以先用规则匹配，不要求接入外部 secret scanning 服务。

### 6.3 结果等级

MVP 使用三个等级：

```text
clean       未发现明显风险
warning     发现可疑内容，但不一定是 secret
high        发现高风险敏感信息
```

### 6.4 输出

扫描输出应包含：

- 文件路径。
- 风险等级。
- 命中规则名称。
- 行号。
- 脱敏后的片段。

不要在输出中完整打印敏感值。

示例：

```text
HIGH  deploy.md:12  database-url  postgres://user:****@host/db
```

### 6.5 退出码

推荐：

```text
0  clean
1  warning or high
2  command error
```

如果实现希望区分 warning 和 high，可以后续扩展。

### 6.6 验收

- mock private key 会被识别。
- mock database URL 会被识别。
- mock password 会被识别。
- 输出不会泄露完整 secret。

## 7. `capture`

### 7.1 用法

```bash
knowblazer capture <file> [--repo <path>]
```

### 7.2 输入要求

MVP 只支持 Markdown 文件。

如果输入不是 Markdown 文件，应失败并提示：

```text
Only Markdown files are supported in MVP.
```

### 7.3 行为

`capture` 应：

1. 读取输入文件。
2. 执行 `scan`。
3. 如果结果为 `clean` 或 `warning`，复制到 `inbox/YYYY-MM-DD/`。
4. 如果结果为 `high`，复制到 `quarantine/YYYY-MM-DD/`。
5. 写入或补充 front matter。
6. 输出保存路径和扫描结果。

### 7.4 命名

目标文件名应可读且稳定。

推荐规则：

```text
YYYYMMDD-HHMMSS-<slug>.md
```

`slug` 可以来自原文件名或一级标题。

### 7.5 不自动 commit

MVP 的 `capture` 不自动执行 `git commit` 或 `git push`。

如果后续增加 sync helper，也必须先 scan 再允许 commit/push。

### 7.6 输出

clean 或 warning：

```text
Captured to inbox: <repo>/inbox/2026-04-29/20260429-123000-deploy-lesson.md
Scan result: clean
```

high：

```text
Sensitive content detected.
Moved to quarantine: <repo>/quarantine/2026-04-29/20260429-123000-deploy-lesson.md
Review and sanitize before promoting or committing.
```

### 7.7 验收

- Markdown 文件能进入 `inbox/YYYY-MM-DD/`。
- 高风险 Markdown 文件进入 `quarantine/YYYY-MM-DD/`。
- capture 输出明确告诉用户下一步。
- capture 不自动 commit/push。

## 8. `promote`

### 8.1 用法

```bash
knowblazer promote <file> --to <target> [--repo <path>]
```

示例：

```bash
knowblazer promote inbox/2026-04-29/deploy.md --to experience/deployment
knowblazer promote inbox/2026-04-29/preferences.md --to profile/preferences.md
knowblazer promote inbox/2026-04-29/project.md --to projects/kiblazer.md
```

### 8.2 允许目标

MVP 只允许提升到：

- `experience/`
- `projects/`
- `profile/`

不允许提升到：

- `quarantine/`
- `recall/`
- `.knowblazer/`
- 记忆库之外的路径

### 8.3 行为

`promote` 应：

1. 确认源文件存在。
2. 对源文件重新执行 `scan`。
3. 如果结果为 `high`，阻止 promote。
4. 确认目标路径在允许范围内。
5. 移动或复制文件到目标路径。
6. 更新 front matter：
   - `status: "promoted"`
   - `promoted_at: "<timestamp>"`
   - `promoted_to: "<target>"`

MVP 默认可以使用移动语义。后续如果需要保留 inbox 原件，可以增加 `--copy`。

### 8.4 目标冲突

如果目标文件已存在，默认不覆盖。

可以提示用户改名。MVP 暂不需要 `--force`。

### 8.5 验收

- clean 文件可以提升到 `experience/`。
- high 文件不能提升。
- 目标路径不能逃逸到记忆库之外。
- 已存在目标不会被覆盖。

## 9. `recall`

### 9.1 用法

```bash
knowblazer recall --task <text> [--repo <path>] [--project <name>] [--output <file>]
```

示例：

```bash
knowblazer recall --task "deploy new frontend" --project kiblazer
```

### 9.2 输入

必填：

- `--task`

可选：

- `--project`
- `--repo`
- `--output`

### 9.3 候选来源

MVP recall 使用以下来源：

1. `profile/preferences.md`
2. `profile/decision-principles.md`
3. `projects/<project>.md`，如果传入 `--project`
4. `experience/**.md` 中与任务关键词匹配的文件
5. 今天和昨天的 `daily/*.md`

默认不读取：

- `inbox/`
- `quarantine/`
- `recall/`

### 9.4 匹配规则

MVP 使用简单关键词匹配：

- 从 `--task` 中提取英文单词、数字、中文连续词段。
- 文件名命中加权。
- 标题命中加权。
- 正文命中计分。
- 优先选择分数最高的若干文件。

不要求语义检索，不要求 embedding。

### 9.5 输出格式

输出应是 Markdown：

```markdown
# Knowblazer Recall Pack

Task: deploy new frontend
Generated at: 2026-04-29T12:30:00+08:00

## Developer Preferences

...

## Project Context

...

## Relevant Experience

...

## Recent Daily Notes

...

## Cautions

- This pack excludes inbox and quarantine by default.
- Verify commands and secrets before running anything.
```

### 9.6 长度控制

MVP 应有默认长度上限。

建议：

- 最多 5 个 experience 文件。
- 每个文件最多摘取前 120 行或前 8KB。
- 输出总量默认不超过约 20KB。

具体数值可以实现时调整，但必须避免拼接整个记忆库。

### 9.7 输出位置

默认输出到 stdout。

如果传入 `--output <file>`，写入指定文件。推荐用户写入 `recall/`，但不强制。

### 9.8 验收

- 没有 `--task` 时失败。
- 输出包含标题、任务、生成时间。
- 输出默认不包含 `inbox/` 和 `quarantine/`。
- 匹配 experience 时不会把整个仓库拼进去。
- `--output` 能写入文件。

## 10. 错误处理

MVP 错误信息应直接说明问题和下一步。

典型错误：

```text
Knowblazer repo not found. Run `knowblazer init <path>` or set KNOWBLAZER_REPO.
```

```text
Target path is outside the Knowblazer repo and is not allowed.
```

```text
Sensitive content detected. File was not promoted.
```

```text
Only Markdown files are supported in MVP.
```

## 11. 安全约束

MVP 必须遵守：

- 不上传任何用户记忆到 Knowblazer 官方服务。
- 不要求用户登录官方账号。
- 不自动 commit。
- 不自动 push。
- 不在 scan 输出中完整打印 secret。
- 不从 `quarantine/` 生成 recall。
- 不把 `inbox/` 默认纳入 recall。

## 12. 实现无关要求

本文不强制实现语言。

但实现应满足：

- 跨平台文件路径处理。
- 对 Git 仓库友好。
- 不依赖网络运行 P0。
- 测试能在临时目录中完成。
- 不要求外部 LLM API。

## 13. P0 测试清单

最小测试应覆盖：

- `init` 创建目录和模板。
- `init` 幂等，不覆盖已有文件。
- `capture` clean 文件进入 inbox。
- `capture` high 文件进入 quarantine。
- `scan` 能识别 mock secret。
- `scan` 输出脱敏。
- `promote` clean 文件进入长期目录。
- `promote` high 文件被阻止。
- `promote` 阻止路径逃逸。
- `recall` 输出 Markdown。
- `recall` 排除 inbox 和 quarantine。
- `recall --output` 写文件。

## 14. 后续非 P0

以下内容不在 MVP 规格内：

- 自动 Git commit/push。
- 官方云服务。
- SpecStory 导入。
- AI 工具 hook 安装。
- MCP server。
- 向量检索。
- 后台自动 promotion。
- Web UI。

这些可以在 P0 验证后单独写规格。
