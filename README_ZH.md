# Knowblazer

[English](./README.md) | 中文

**面向 AI 辅助编程的本地优先、Git 友好的动态工程记忆系统。**

Knowblazer 是一个用于维护您个人私有工程知识库的开源工具。它能持续收集 AI 编码过程中的有用信号，进行敏感信息扫描，将新鲜的记忆沉淀为 Markdown 格式，并自动将其整合（Consolidate）为高密度的工程上下文，重新注入到 Claude Code、Codex、Gemini、Cursor 等 AI 编程工具中。

Knowblazer **不会**托管您的记忆。您的知识库存储在您自己选择的地方：本地磁盘、GitHub 私有仓库、GitLab 私有仓库、自建 Git、NAS 或其他您拥有的 Git 远程端。

---

## 为什么需要它？

AI 辅助编程正逐渐分散到多个不同的工具中。这台工具可能知道您的部署习惯，那台工具可能知道某个特定的项目约束，而第三台工具里则可能保留着解释了某个经常发生的失败的调试会话。

原始的对话历史（Transcripts）很有用，但它们不等于持久的工程记忆。Knowblazer 专注于原始对话捕获之上的这一层：

*   开发者偏好（Developer preferences）
*   项目背景上下文（Project context）
*   可复用的部署与调试经验（Reusable deployment and debugging lessons）
*   决策原则（Decision principles）
*   每日工作笔记（Daily working notes）
*   面向 AI 编程任务的、隐私安全的召回包（Recall packs）

---

## 它是与不是

**Knowblazer 是：**
*   一个用于私有工程记忆库的本地 CLI 工作流
*   一种纯 Markdown / Git 的记忆存储结构
*   一套开箱即用的模板和策略
*   一套“隐私优先”的捕获、扫描、整合与召回流程

**Knowblazer 不是：**
*   云端记忆服务（Cloud memory service）
*   官方托管笔记的平台
*   完全的 AI 会话录制器（AI transcript recorder）
*   通用的向量数据库（Vector database）
*   又一个复杂的数据库守护进程 MCP 服务器
*   团队知识库（Team wiki）

---

## 安装

### 1. 快捷一键安装 (Mac & Linux)
直接下载并安装预编译二进制文件，无需本地安装 Go 环境：

```bash
curl -fsSL https://raw.githubusercontent.com/KiBlazer/knowblazer/main/install.sh | bash
```

### 2. 查看版本与更新
检查当前版本并自动更新至最新发布版：

```bash
# 查看版本
knowblazer version

# 原地一键自更新
knowblazer update
```

或重新运行一键安装脚本进行覆盖更新：
```bash
curl -fsSL https://raw.githubusercontent.com/KiBlazer/knowblazer/main/install.sh | bash
```

### 3. 源码编译安装 (Go 开发者)
如果您拥有 Go 环境，可直接通过 `go install` 安装最新版：

```bash
go install github.com/KiBlazer/knowblazer/cmd/knowblazer@latest
```

或克隆仓库后在本地编译安装：

```bash
make install
```

---

## 快速开始

### 1. 初始化记忆库并连接 AI 工具

在您的任何项目目录中启动：

```bash
cd /path/to/project
knowblazer setup
```

或者开启全局记忆支持，在所有工作区中自动启用 Knowblazer：

```bash
knowblazer setup --global
```

`knowblazer setup` 会在 `~/knowblazer-notes` 创建或复用您的记忆库，根据当前目录名推断项目名称，并自动连接检测到的 AI 编程工具：
*   **Claude Code**：自动写入 `CLAUDE.md` 策略文件，并自动在后台配置 MCP 服务。
*   **Codex CLI**：自动写入 `AGENTS.md` 策略文件，并自动在后台配置 MCP 服务。
*   **Antigravity (agy)**：动态更新全局 `mcp_config.json` 配置文件，自动在后台注册 MCP 服务。

然后直接启动您的 AI 客户端：

```bash
claude
# 或者
codex
```

### 2. 每日核心高频指令：

```bash
knowblazer remember "部署流程需要先运行数据库迁移脚本"       # 沉淀工程经验或 Markdown 文件
knowblazer remember "已发布 v0.3.0 稳定版" --daily          # 追加到当天的工程师日志
knowblazer recall "如何进行项目部署？"                      # 检索并生成精准上下文召回包
knowblazer status                                           # 检查工作区映射与记忆状态
knowblazer sync                                             # 安全扫描并同步私有 Git 仓库
```

*   `remember`: 记录值得保留的经验；干净的笔记将写入 `experience/auto/` 并自动整合，包含密钥等高危内容则会被隔离。
*   `status`: 检查当前工作区映射、指令注入文件与动态记忆统计。
*   `sync`: 只有在您主动调用时，它才会对修改的文件进行安全扫描，并提交推送至您的私有远程 Git 仓库。

---

## 完整命令体系一览

Knowblazer 经过重构，采用简洁优雅的三层指令分类：

### 核心操作 (Core Commands)
* `knowblazer setup [--global] [--path <dir>] [--tool <name>] [--skip-mcp]`：初始化记忆库并连接 AI 工具（支持全局注入）。
* `knowblazer remember <文本|文件> [-d|--daily]`：沉淀经验、解决方案、Markdown 文档或追加日常日志。
* `knowblazer recall <任务描述> [--project <项目名>] [--output <文件>]`：检索并生成针对当前编码任务的高密度召回包。
* `knowblazer status [--path <目录>]`：检查当前工作区项目映射、工具指令注入状态与记忆量统计。
* `knowblazer sync [status|commit|push|pull] [-m <提交信息>]`：带敏感信息预检的 Git 同步。

### 经验管理与维护 (Management & Curation)
* `knowblazer memory list`：列出收件箱中未审核的候选记忆。
* `knowblazer memory review [suggest]`：审核收件箱或生成经验提炼整合建议。
* `knowblazer memory promote <文件> --to <目标分类>`：将候选记忆正式提升归档到经验库。
* `knowblazer memory reject <文件>`：拒绝并删除收件箱候选。
* `knowblazer memory consolidate`：触发新鲜记忆整合，浓缩成高信息密度 Markdown 摘要。
* `knowblazer memory search <关键词>`：在整个记忆库中进行 BM25 / 词频检索。
* `knowblazer memory capture <文件>`：直接捕获外部 Markdown 笔记至收件箱（带安全拦截）。
* `knowblazer memory import specstory <路径>`：将历史 SpecStory AI 对话提炼导入为经验候选。
* `knowblazer memory scan <路径>`：安全扫描指定文件或目录，检查是否有泄漏的 Token 或密钥。
* `knowblazer daily [show|add <文本>]`：查看或记录当天的工作日志。
* `knowblazer project [list|set <名称>|clear]`：管理工作区路径与项目名称的绑定映射。
* `knowblazer doctor`：全面体检记忆库结构完整性与敏感信息隔离状态。
* `knowblazer backup <create|restore> [--passphrase <口令>]`：创建或恢复加密备份包。

### 系统与服务 (System)
* `knowblazer mcp [serve]`：启动标准 Model Context Protocol 服务，供 AI 工具通过 stdio 实时调用。
* `knowblazer update`：自更新 Knowblazer 二进制程序至 GitHub 最新发布版本。
* `knowblazer version`：打印版本号、编译 commit 与架构信息。

---

---

## 记忆库目录结构

默认的记忆库完全由人类可读的 Markdown 文件组成，存储在本地：

```text
knowblazer-notes/
├── .knowblazer/
│   └── config.json       # 索引与项目映射配置
├── AI-SETUP.md
├── inbox/                # 待审核/导入的候选记忆
├── daily/                # 短期工作日志、今日笔记
├── projects/             # 长期项目事实、约束和运行说明
├── experience/           # 可复用的工程经验
│   ├── auto/             # 新鲜捕获的记忆缓存区
│   └── synthesized/      # 自动整合后的高密度记忆层
└── quarantine/           # 隔离区（敏感信息，默认不参与召回）
```

---

## 隐私与安全模型

Knowblazer 坚持 **本地优先**。

核心命令不需要任何以下网络支持即可工作：
*   无需 Knowblazer 注册账号
*   无需连接 Knowblazer 云端服务器
*   无需网络连接
*   无需配置 Git 远程端
*   基础的本地工作流无需 LLM API 参与

安全扫描器运行在本地，防止 AI 在记忆中捕获诸如 AWS 密钥、数据库连接密码等高危内容。

---

## 系统对比 (Knowblazer vs 其他记忆方案)

| 特性 | Knowblazer | SQLite-based / 传统 MCP 记忆服务 |
| :--- | :--- | :--- |
| **存储介质** | **纯文本 Markdown 文件夹**（开发者最爱） | SQLite / PostgreSQL 等数据库（黑盒） |
| **Git 友好度** | **天然支持**（可随时 `git diff` / `commit` / 冲突解决） | 较差（无法直接 diff 数据库二进制） |
| **安全隔离** | **主动扫描并隔离密钥**（Quarantine 机制） | 无，盲目接受 AI 写入的一切内容 |
| **部署成本** | **零依赖**（单 Go 二进制，即装即用） | 需要运行 Python 后端守护进程、数据库、Node.js UI |
| **召回方式** | 针对当前任务，动态拼装打包的 Recall Pack | 低阶的 KV 检索或模糊搜索 |

---

## 关系与 SpecStory 的互补

SpecStory 非常适合捕获实时的 AI 对话和意图历史。
Knowblazer 是它的互补工具：它把对话历史、日常笔记和经验整理成动态的、高内聚的私有工程记忆库。

```text
SpecStory: 原始对话与意图归档
Knowblazer: 面向 AI 编码的动态私有工程记忆库
```

---

## 授权协议

本项目采用 MIT 授权协议。详情见 [LICENSE](LICENSE)。
