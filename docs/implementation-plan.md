# Knowblazer P0 实现计划

日期：2026-04-29

## 1. 实现目标

P0 目标是实现一个可本地运行的 CLI，跑通 MVP 规格中的核心路径：

```text
init
  ↓
capture
  ↓
scan
  ↓
promote
  ↓
recall
```

P0 不依赖网络、不依赖官方账号、不依赖 LLM API、不自动 commit、不自动 push。

## 2. 推荐技术栈

推荐使用 Go。

原因：

- 单二进制分发适合 CLI 工具。
- 标准库对文件、路径、时间、正则、测试支持足够。
- 跨平台成本较低。
- 对 Git/Markdown 这类本地文件工作流很合适。
- P0 不需要复杂运行时或外部服务。

可以使用少量依赖，但 P0 应优先标准库。CLI 参数解析可以先用标准库 `flag`，后续需要更好 UX 时再引入 Cobra 或 urfave/cli。

## 3. 代码结构建议

建议结构：

```text
knowblazer/
├── cmd/
│   └── knowblazer/
│       └── main.go
├── internal/
│   ├── cli/
│   ├── repo/
│   ├── templates/
│   ├── scan/
│   ├── capture/
│   ├── promote/
│   ├── recall/
│   └── markdown/
├── templates/
│   └── default-memory-repo/
├── docs/
├── go.mod
└── README.md
```

模块职责：

- `cli/`：命令分发、参数解析、错误输出。
- `repo/`：记忆库发现、路径校验、配置读取、目录创建。
- `templates/`：嵌入默认模板并写入目标目录。
- `scan/`：敏感信息规则、扫描结果、脱敏输出。
- `capture/`：读取 Markdown、扫描、写入 inbox/quarantine。
- `promote/`：目标路径校验、扫描、移动文件、更新状态。
- `recall/`：关键词匹配、候选文件选择、上下文包生成。
- `markdown/`：标题提取、front matter 读写、slug 生成。

## 4. 实施阶段

### 4.1 阶段一：项目骨架

任务：

- 初始化 Go module。
- 创建 CLI 入口。
- 支持 `knowblazer --help`。
- 支持命令分发：`init`、`scan`、`capture`、`promote`、`recall`。
- 增加基础测试目录。

验收：

- `go test ./...` 通过。
- 未实现命令给出清晰错误或 help。

### 4.2 阶段二：repo 与 init

任务：

- 实现 `.knowblazer/config.json` 识别。
- 实现 repo 查找顺序：
  1. 当前目录。
  2. 父级目录。
  3. `KNOWBLAZER_REPO`。
  4. 默认路径。
- 使用 `templates/default-memory-repo/` 生成记忆库。
- 保证 init 幂等，不覆盖已有文件。
- 非空非 Knowblazer 目录默认失败。

验收：

- 空目录 init 生成完整结构。
- 重复 init 不覆盖用户内容。
- 非空普通目录 init 失败并提示。

### 4.3 阶段三：secret scan

任务：

- 实现文件和目录扫描。
- 递归扫描目录中的 Markdown、文本、env 类文件。
- 实现规则：
  - private key block
  - token/api_key/secret 字段
  - password 字段
  - database URL
  - common cloud secret env
- 实现 risk level：`clean`、`warning`、`high`。
- 输出行号、规则名、脱敏片段。
- 避免完整输出 secret。

验收：

- mock private key 被识别。
- mock database URL 被识别。
- mock password 被识别。
- 输出脱敏。
- clean 返回 exit code 0，风险返回 exit code 1，命令错误返回 exit code 2。

### 4.4 阶段四：Markdown 工具

任务：

- 提取一级标题。
- 从文件名生成 slug。
- 生成 front matter。
- 更新 front matter 状态字段。
- 保持正文可读，不破坏原文。

验收：

- 无标题文件也能生成可读标题。
- 已有 front matter 的文件不会被粗暴破坏。
- 生成的 Markdown 人类可读。

### 4.5 阶段五：capture

任务：

- 只接受 `.md` / `.markdown`。
- capture 前执行 scan。
- clean/warning 写入 `inbox/YYYY-MM-DD/`。
- high 写入 `quarantine/YYYY-MM-DD/`。
- 目标文件名使用 `YYYYMMDD-HHMMSS-<slug>.md`。
- 写入 front matter。
- 不自动 commit/push。

验收：

- clean Markdown 进入 inbox。
- high Markdown 进入 quarantine。
- 输出明确保存路径和扫描结果。
- 不会修改源文件。

### 4.6 阶段六：promote

任务：

- 支持 `promote <file> --to <target>`。
- 源文件重新 scan。
- high 文件阻止 promote。
- 目标只允许 `experience/`、`projects/`、`profile/`。
- 阻止路径逃逸。
- 默认不覆盖已有目标。
- 更新 front matter：`status`、`promoted_at`、`promoted_to`。

验收：

- inbox 文件可提升到 `experience/deployment/`。
- high 文件不可提升。
- `--to ../../x` 被阻止。
- 已存在目标不覆盖。

### 4.7 阶段七：recall

任务：

- 支持 `recall --task <text>`。
- 读取：
  - `profile/preferences.md`
  - `profile/decision-principles.md`
  - `projects/<project>.md`
  - 匹配的 `experience/**/*.md`
  - 今天和昨天的 `daily/*.md`
- 默认排除 `inbox/`、`quarantine/`、`recall/`。
- 实现简单关键词匹配和文件排序。
- 生成 Markdown recall pack。
- 支持 `--output <file>`。
- 控制输出长度。

验收：

- 无 task 失败。
- 输出包含任务和生成时间。
- 默认不包含 inbox/quarantine。
- 匹配 experience 不拼接整个仓库。
- `--output` 能写文件。

### 4.8 阶段八：端到端验证

任务：

- 用临时目录模拟完整路径。
- 加入 README 中的命令示例验证。
- 增加一个 fixture：
  - clean lesson
  - secret lesson
  - project context
  - experience note
  - daily note

验收：

- `go test ./...` 通过。
- 手动执行 MVP flow 成功。
- 无网络环境可运行。

## 5. 测试策略

P0 以单元测试和临时目录集成测试为主。

必须测试：

- repo 查找和路径安全。
- init 幂等。
- scan 规则和脱敏。
- capture 目标选择。
- promote 目标限制。
- recall 排除规则。

测试应使用临时目录，不依赖用户真实 home，不读写真实 Git remote。

## 6. 路径安全要求

所有写入都必须在 Knowblazer repo 内，除非用户显式指定读取源文件。

必须防止：

- `../` 路径逃逸。
- 通过 symlink 写出 repo。
- promote 到 `.knowblazer/`。
- promote 到 `quarantine/` 或 `recall/`。
- recall 读取 `quarantine/`。

## 7. 输出风格

CLI 输出应简洁、明确、可脚本化。

推荐：

- 成功输出保存路径。
- 风险输出命中规则和脱敏片段。
- 错误输出下一步建议。

避免：

- 打印完整 secret。
- 输出大段无关说明。
- 自动执行 Git 操作。

## 8. 暂不实现

P0 不实现：

- Git commit/push。
- Git remote 配置。
- SpecStory 导入。
- AI 工具 hook。
- MCP server。
- SQLite 或向量索引。
- LLM 总结。
- Web UI。
- 后台任务。

## 9. 建议里程碑

建议按以下顺序提交：

1. Go module and CLI skeleton.
2. Repo discovery and init from templates.
3. Secret scanner.
4. Capture command.
5. Promote command.
6. Recall command.
7. End-to-end tests and README command alignment.

每个里程碑都应保持 `go test ./...` 通过。
