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
curl -fsSL https://raw.githubusercontent.com/KiBlazer/knowblazer/main/install.sh | sh
```

### 2. 源码编译安装 (Go 开发者)
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

在您的任何项目目录中启动：

```bash
cd /path/to/project
knowblazer start
```

`knowblazer start` 会在 `~/knowblazer-notes` 创建或复用您的记忆库，根据当前目录名推断项目名称，并自动连接检测到的 AI 编程 CLI 工具。
*   如果检测到 **Claude Code**，它将写入并配置 `CLAUDE.md` 以通过 MCP 连接。
*   如果检测到 **Codex CLI**，它将写入 `AGENTS.md` 以通过 MCP 连接。

然后直接启动您的 AI 客户端：

```bash
claude
# 或者
codex
```

### 每日高频指令：

```bash
knowblazer remember "部署流程需要先运行数据库迁移脚本"
knowblazer recall "如何进行项目部署？"
knowblazer status
knowblazer sync
```

*   `remember`: 记录值得保留的经验；干净的笔记将写入 `experience/auto/` 并自动整合，包含密钥等高危内容则会被隔离。
*   `status`: 检查您的动态记忆统计。
*   `sync`: 只有在您主动调用时，它才会对修改的文件进行安全扫描，并提交推送至您的私有远程 Git 仓库。

---

## MVP 运行逻辑

```text
从项目目录启动
  ↓
连接 Claude Code / Cursor 到私有记忆
  ↓
自动/手动捕获有用的经验
  ↓
安全扫描（检测密钥与敏感信息）
  ↓
干净内容进入 experience/auto/，高危内容进入 quarantine/ (隔离区)
  ↓
自动将新鲜记忆整合（Consolidate）为高密度上下文
  ↓
在编码时，根据当前 Task 动态生成 Recall Pack (召回包) 注入 AI 窗口
```

---

## MCP 服务器集成配置

若要在 **Cursor** 或 **Claude Desktop** 中使用 Knowblazer 的 MCP 服务，可在客户端 MCP settings 中添加如下配置：

```json
{
  "mcpServers": {
    "knowblazer": {
      "command": "knowblazer",
      "args": ["serve"],
      "env": {
        "KNOWBLAZER_NOTES_PATH": "/path/to/your/knowblazer-notes"
      }
    }
  }
}
```

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

| 特性 | Knowblazer | Nocturne Memory / 传统 MCP 记忆服务 |
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
