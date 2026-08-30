# Crush

[English](README.md) | 中文

> [!IMPORTANT]
> **本项目是 fork。** Crush 是 [Charm](https://charm.land) 出品的终端 AI 编程助手。本仓库是 [charmbracelet/crush](https://github.com/charmbracelet/crush) 的社区 fork，跟随上游项目，可能带有 fork 特有的改动。原项目、官方文档和发行版请参见[上游仓库](https://github.com/charmbracelet/crush)。

<p align="center">
    <a href="https://stuff.charm.sh/crush/charm-crush.png"><img width="450" alt="Charm Crush Logo" src="https://github.com/user-attachments/assets/cf8ca3ce-8b02-43f0-9d0f-5a331488da4b" /></a><br />
    <a href="https://github.com/f69wrthhqd-blip/crush/releases"><img src="https://img.shields.io/github/release/f69wrthhqd-blip/crush" alt="Latest Release"></a>
    <a href="https://github.com/f69wrthhqd-blip/crush/actions"><img src="https://github.com/f69wrthhqd-blip/crush/actions/workflows/build.yml/badge.svg" alt="Build Status"></a>
</p>

<p align="center">终端里的编程新搭档，<br />无缝接入你的工具、代码与工作流，全面兼容主流 LLM 模型。</p>
<p align="center">Your new coding bestie, now available in your favourite terminal.<br />Your tools, your code, and your workflows, wired into your LLM of choice.</p>

<p align="center"><img width="800" alt="Crush Demo" src="https://github.com/user-attachments/assets/58280caf-851b-470a-b6f7-d5c4ea8a1968" /></p>

## 特性

- **多模型支持：** 可从大量 LLM 中选择，或通过 OpenAI 兼容 / Anthropic 兼容 API 添加你自己的模型
- **灵活切换：** 会话中途可切换 LLM，且保留上下文
- **基于会话：** 每个项目可维护多个工作会话与上下文
- **LSP 增强：** 与人类开发者一样，Crush 借助 LSP 获取额外上下文
- **可扩展：** 通过 MCP（`http`、`stdio` 和 `sse`）添加能力
- **随处可用：** 在 macOS、Linux、Windows（PowerShell 和 WSL）、Android、FreeBSD、OpenBSD 和 NetBSD 的每个终端中提供一流支持
- **工业级：** 基于 Charm 生态构建，为 25,000+ 应用提供动力，从领先的开源项目到关键业务基础设施

## 安装

> [!IMPORTANT]
> 本项目是 fork。下面的包管理器命令安装的是 Charm 发布的**上游** Crush 发行版，而非本 fork。要运行本 fork，请使用下面的 `go install` 命令从源码构建，或从[本 fork 的 releases][releases] 下载二进制文件。

使用包管理器：

```bash
# Homebrew
brew install charmbracelet/tap/crush

# NPM
npm install -g @charmland/crush

# Arch Linux
yay -S crush-bin

# Nix
nix run github:numtide/nix-ai-tools#crush

# FreeBSD
pkg install crush
```

Windows 用户：

```bash
# Winget
winget install charmbracelet.crush

# Scoop
scoop bucket add charm https://github.com/charmbracelet/scoop-bucket.git
scoop install crush
```

<details>
<summary><strong>Nix（NUR）</strong></summary>

Crush 也可通过官方 Charm [NUR](https://github.com/nix-community/NUR) 的 `nur.repos.charmbracelet.crush` 获得，这是在 Nix 中获取 Crush 的最新方式。

你也可以通过 `nix-shell` 试用 NUR 中的 Crush：

```bash
# 添加 NUR 频道。
nix-channel --add https://github.com/nix-community/NUR/archive/main.tar.gz nur
nix-channel --update

# 在 Nix shell 中获取 Crush。
nix-shell -p '(import <nur> { pkgs = import <nixpkgs> {}; }).repos.charmbracelet.crush'
```

### 通过 NUR 使用 NixOS 与 Home Manager 模块

Crush 通过 NUR 提供 NixOS 和 Home Manager 模块。你可以直接在 flake 中导入使用。由于它会自动检测是 home manager 还是 nixos 上下文，因此两种场景的导入方式完全相同：

```nix
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    nur.url = "github:nix-community/NUR";
  };

  outputs = { self, nixpkgs, nur, ... }: {
    nixosConfigurations.your-hostname = nixpkgs.lib.nixosSystem {
      system = "x86_64-linux";
      modules = [
        nur.modules.nixos.default
        nur.repos.charmbracelet.modules.crush
        {
          programs.crush = {
            enable = true;
            settings = {
              providers = {
                openai = {
                  id = "openai";
                  name = "OpenAI";
                  base_url = "https://api.openai.com/v1";
                  type = "openai";
                  api_key = "sk-fake123456789abcdef...";
                  models = [
                    {
                      id = "gpt-4";
                      name = "GPT-4";
                    }
                  ];
                };
              };
              lsp = {
                go = { command = "gopls"; enabled = true; };
                nix = { command = "nil"; enabled = true; };
              };
              options = {
                context_paths = [ "/etc/nixos/configuration.nix" ];
                tui = { compact_mode = true; };
                debug = false;
              };
            };
          };
        }
      ];
    };
  };
}
```

</details>

<details>
<summary><strong>Debian/Ubuntu</strong></summary>

```bash
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://repo.charm.sh/apt/gpg.key | sudo gpg --dearmor -o /etc/apt/keyrings/charm.gpg
echo "deb [signed-by=/etc/apt/keyrings/charm.gpg] https://repo.charm.sh/apt/ * *" | sudo tee /etc/apt/sources.list.d/charm.list
sudo apt update && sudo apt install crush
```

</details>

<details>
<summary><strong>Fedora/RHEL</strong></summary>

```bash
echo '[charm]
name=Charm
baseurl=https://repo.charm.sh/yum/
enabled=1
gpgcheck=1
gpgkey=https://repo.charm.sh/yum/gpg.key' | sudo tee /etc/yum.repos.d/charm.repo
sudo yum install crush
```

</details>

或者下载安装：

- [软件包][releases] 提供 Debian 和 RPM 格式
- [二进制文件][releases] 适用于 Linux、macOS、Windows、FreeBSD、OpenBSD 和 NetBSD

[releases]: https://github.com/f69wrthhqd-blip/crush/releases

或者直接用 Go 安装（安装的是**本 fork**，从源码构建）：

```
go install github.com/f69wrthhqd-blip/crush@latest
```

在 illumos（OpenIndiana、OmniOS）上，上述命令可直接使用。只是该系统没有原生桌面通知；基于终端的通知（OSC）和终端铃声仍然可用。在 Oracle Solaris 上，请添加 `-tags sqlite3_dotlk` 以使用点文件锁定的本地数据库：

```
go install -tags sqlite3_dotlk github.com/f69wrthhqd-blip/crush@latest
```

> [!WARNING]
> 使用 Crush 后生产力可能会提升，初次使用还可能被 nerd snipe 击中。如果症状持续，欢迎加入 [Slack][slack] 或 [Discord][discord]，把我们都 nerd snipe 一遍。

## 快速上手

最快的上手方式是从模型选择器中选择一个 [Hyper][hyper] 模型，按照步骤完成认证即可。

[Hyper] 是 Charm 推出的 Crush 官方提供商。它采用订阅制，提供免费额度，并为 Crush 深度优化。它注重隐私，零数据保留（ZDR），并设计为符合 GDPR。[了解更多 Hyper 信息][hyper]。

<p><a href="https://hyper.charm.land"><img width="340" height="200" alt="Charm Hyper" src="https://github.com/user-attachments/assets/50875289-7992-454d-9f14-9f790413fb5e" /></a></p>

## API Keys

你还可以将 Crush 与许多其他提供商配合使用，例如 Anthropic、OpenAI、Gemini、OpenRouter 等。按 <kbd>ctrl+l</kbd> 打开模型选择器，选择你想要的提供商，粘贴你的 API key 即可。

此外，你也可以为常用的提供商设置环境变量：

| 环境变量                       | 提供商                                           |
| ------------------------------ | ------------------------------------------------ |
| `HYPER_API_KEY`                | [Charm Hyper][hyper]                             |
| `ANTHROPIC_API_KEY`            | Anthropic                                        |
| `OPENAI_API_KEY`               | OpenAI                                           |
| `VERCEL_API_KEY`               | Vercel AI Gateway                                |
| `GEMINI_API_KEY`               | Google Gemini                                    |
| `ZAI_API_KEY`                  | Z.ai                                             |
| `MINIMAX_API_KEY`              | MiniMax                                          |
| `SYNTHETIC_API_KEY`            | Synthetic                                        |
| `HF_TOKEN`                     | Hugging Face Inference                           |
| `CEREBRAS_API_KEY`             | Cerebras                                         |
| `OPENROUTER_API_KEY`           | OpenRouter                                       |
| `IONET_API_KEY`                | io.net                                           |
| `ALIBABA_SINGAPORE_API_KEY`    | Alibaba（新加坡）                                |
| `ALIBABA_US_API_KEY`           | Alibaba（美国）                                  |
| `GROQ_API_KEY`                 | Groq                                             |
| `AVIAN_API_KEY`                | Avian                                            |
| `OPENCODE_API_KEY`             | OpenCode Zen & Go                                |
| `VERTEXAI_PROJECT`             | Google Cloud VertexAI（Gemini）                  |
| `VERTEXAI_LOCATION`            | Google Cloud VertexAI（Gemini）                  |
| `AWS_ACCESS_KEY_ID`            | Amazon Bedrock（Claude）                         |
| `AWS_SECRET_ACCESS_KEY`        | Amazon Bedrock（Claude）                         |
| `AWS_REGION`                   | Amazon Bedrock（Claude）                         |
| `AWS_PROFILE`                  | Amazon Bedrock（自定义 Profile）                 |
| `AWS_BEARER_TOKEN_BEDROCK`     | Amazon Bedrock                                   |
| `AZURE_OPENAI_API_ENDPOINT`    | Azure OpenAI 模型                                |
| `AZURE_OPENAI_API_KEY`         | Azure OpenAI 模型（使用 Entra ID 时可省略）      |
| `AZURE_OPENAI_API_VERSION`     | Azure OpenAI 模型                                |
| `MOONSHOT_API_KEY`             | Moonshot                                         |

[hyper]: https://hyper.charm.land

另外请注意，Crush 几乎可以支持任何提供商，包括[本地模型](#本地模型)。更多信息见下方的[自定义提供商](#自定义提供商)。

### 顺便一提

希望 Crush 支持某个提供商？或者某个现有模型需要更新？

Crush 的默认模型列表由 [Catwalk](https://github.com/charmbracelet/catwalk) 管理，这是一个社区维护的开源 Crush 兼容模型仓库，欢迎贡献。

<a href="https://github.com/charmbracelet/catwalk"><img width="174" height="174" alt="Catwalk Badge" src="https://github.com/user-attachments/assets/95b49515-fe82-4409-b10d-5beb0873787d" /></a>

## 配置

> [!TIP]
> Crush 内置了自我配置的 skill。大多数时候你只需告诉它想配置什么，它就能搞定。

Crush 无需任何配置即可良好运行。当然，如果你需要或想要定制 Crush，可以通过 `crushrc` 实现。

`crushrc` 本质上就是带 Crush 专用内置命令的 Bash，很像为你的 Crush 准备的 `.bashrc`。由于 Crush 内置了原生 Bash 解释器，基于 Bash 的配置在所有平台（包括 Windows）上行为完全一致。

例如：

```bash
# 添加 Ollama。
provider add ollama --type ollama --base-url "http://localhost:11434/v1"

# 在 Ollama 上注册一个模型。
model add ollama/llama3.3 --name "Llama 3.3" --context-window 128000

# 自动批准部分工具。
permissions allow view edit

# 在特定机器上引入其他文件。
if [[ $HOSTNAME == "babysquid" ]]; then
    source ~/my-stuff/babysquid.sh
fi

# 添加一个 MCP 服务器，GitHub API token 存储在 1Password 中。
mcp add github \
  --type http \
  --url "https://api.github.com/mcp/" \
  --header Authorization "Bearer $(op read 'op://my-secret-key')"
```

配置可以放在项目本地，也可以全局设置，优先级如下：

| 优先级 | Unix 类系统                | Windows                                 |
| ------ | -------------------------- | --------------------------------------- |
| 1      | `./.crushrc`               | `.\.crushrc`                            |
| 2      | `./crushrc`                | `.\crushrc`                             |
| 3      | `~/.config/crush/crushrc`  | `%USERPROFILE%\.config\crush\crushrc`   |

（Crush 遵循 [XDG Base Directory 规范][xdg]，因此你的路径可能因 `XDG_CONFIG_HOME` 的值而异。数据目录如 `~/.local/share/crush` 和 `%LOCALAPPDATA%\crush` 只存放 JSON 状态；Crush 不会从这些目录执行 `crushrc`。）

[xdg]: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html

那么旧的 JSON 格式呢？它仍然受支持，但已被视为弃用。详见：[配置文档](./docs/config/)。

> [!TIP]
> 你可以通过设置以下变量覆盖用户与数据配置的位置：
>
> - `CRUSH_GLOBAL_CONFIG`
> - `CRUSH_GLOBAL_DATA`

另外，Crush 还会在另一个位置存放临时数据（例如应用状态）。这是状态数据，不应手动编辑，也不应视为配置。

```bash
# Unix
$HOME/.local/share/crush/crush.json

# Windows
%LOCALAPPDATA%\crush\crush.json
```

#### 关于安全性的说明

`crushrc` 和 `crush.json` 都是受信任的代码；`crushrc` 在完整的 shell 中运行，`crush.json` 中的任何 `$(...)` 都会在加载时执行。不要在未经审查配置的目录中启动 Crush，也不要随意 `source` 来自互联网的文件到你的配置中。

### 环境变量

顶层 `env` 字段在启动时、配置提供商之前设置环境变量。这对于设置影响提供商认证的变量（例如 AWS SDK 凭据链）很有用，无需用 shell 脚本包装 `crush` 命令或在 shell profile 中导出它们：

```json
{
  "$schema": "https://charm.land/crush.json",
  "env": {
    "AWS_PROFILE": "my-sso-profile"
  }
}
```

值支持与其他配置字段相同的 `$VAR` 和 `$(command)` 展开，因此你可以引用现有环境变量或通过 shell 获取值。

### LSP

Crush 可以像开发者一样使用 LSP 获取额外上下文，辅助决策。可以这样手动添加 LSP：

```bash
# crushrc

lsp add go --command "gopls" --env "GOTOOLCHAIN go1.24.5"
lsp add typescript --command "typescript-language-server" --args --stdio
lsp add nix --command "nil"
```

### MCP

Crush 还支持通过三种传输类型接入模型上下文协议（MCP）服务器：命令行服务器使用 `stdio`，HTTP 端点使用 `http`，Server-Sent Events 使用 `sse`。

```bash
# crushrc

# 添加一个运行 Node.js 脚本的本地 MCP 服务器。
mcp add filesystem --command node --args /path/to/mcp-server.js \
  --timeout 10 --disabled-tools some-tool-name --env NODE_ENV production

# 添加一个使用 API token 的 GitHub MCP 服务器。
mcp add github --type http --url https://api.github.com/mcp/ \
  --timeout 10 --header Authorization "Bearer $GH_PAT" \
  --disabled-tools create_issue --disabled-tools create_pull_request

# 添加一个使用 SSE 的流式 MCP 服务器。
mcp add streaming-service --type sse --url "https://example.com/mcp/sse" \
  --timeout 10 --header API-Key "$API_KEY"
```

#### MCP OAuth

需要 OAuth 的 HTTP 和 SSE MCP 服务器可以使用 Crush 内置的授权码流程，而非静态的 `Authorization` 请求头。设置 `"oauth": true` 即可启用：

```json
{
  "mcp": {
    "linear": {
      "type": "http",
      "url": "https://mcp.linear.app/mcp",
      "oauth": true
    }
  }
}
```

##### 预注册客户端

部分服务器（GitHub、Slack）不支持动态客户端注册。对于这些服务器，需要在提供商处注册一个 OAuth 应用并提供凭据。所有值都支持 shell 展开：

```json
{
  "mcp": {
    "github": {
      "type": "http",
      "url": "https://api.github.com/mcp/",
      "oauth": true,
      "oauth_client_id": "Iv1.abc123def456",
      "oauth_client_secret": "$GITHUB_MCP_SECRET",
      "oauth_callback_port": 40704
    }
  }
}
```

当设置了 `oauth_client_id` 时，Crush 会跳过动态客户端注册，以指定客户端身份认证。未设置时，Crush 会自动尝试动态注册（适用于 Linear、Notion 等支持 RFC 7591 的服务器）。

#### 无会话服务器

部分 HTTP MCP 服务器是无会话的——它们从不签发 `Mcp-Session-Id`，并拒绝 Crush 为列表变更通知打开的 `subscriptions/listen` 流，否则会中断连接。Crush 会自动检测已知的无会话服务器（GitHub MCP、`api.githubcopilot.com/mcp`），这些无需额外配置。

对于其他无会话服务器，显式标记 `"sessionless": true`（或在 `crushrc` 中使用 `--sessionless true`）；设为 `false` 可对自动检测的 URL 强制默认行为。权衡是：无会话服务器不会推送实时的工具/提示词/资源列表变更通知。

### Hooks

Crush 对 hooks 提供初步支持。详情参见[hook 指南](./docs/hooks/)。

### 跨客户端共享工作区

当 Crush 连接到共享后端时（例如两个 TUI 连接同一个 `crush serve`），客户端会按解析后的 `--cwd` 归入**工作区**。两个 `--cwd` 相同的客户端会加入同一个底层工作区，共享会话列表、消息历史、权限队列、LSP 和 MCP 状态。

加入是隐式的：将第二个客户端指向同一工作目录即可附着到现有工作区。不过，每次新调用默认都会在新的独立会话中启动。要接续另一个客户端已打开的对话，请使用会话管理器（会话选择器）并选择它。会话在那里会显示两个信号：

- `IsBusy` 表示该会话正在进行 agent 回合。
- `AttachedClients` 表示当前有多少客户端在查看它。

`AttachedClients` 非零（通常与 `IsBusy` 结合）是会话在另一客户端上"进行中"的提示，加入后将实时镜像该视图。

第一个创建工作区的客户端会固定整个进程级标志。特别是 `--yolo` 和 `--debug` 遵循**先到先得**规则：稍后到达同一 `--cwd` 且标志值不同的客户端不会改变运行中的工作区。会输出一条 debug 日志记录该不一致，工作区保留创建时的标志。

工作区在至少一个客户端保持 SSE 事件流连接期间持续存在。当最后一个流断开时，工作区被销毁。`POST /v1/workspaces` 之后有一个短暂的宽限期，以免刚创建工作区但尚未打开事件流的客户端在被附着前就被回收。

### 全局上下文文件

Crush 自动包含两个文件用于跨项目指令。可以把它们看作对系统提示词的个性化补充。

- `~/.config/crush/CRUSH.md`：仅适用于 Crush 的规则，放在这里不会干扰其他 agentic 编码工具。如果你只使用 Crush，只需编辑这一个文件。
- `~/.config/AGENTS.md`：其他编码工具也可能读取的通用指令。避免在这里提及 Crush 特有的功能或工作流。如果你使用多个 agentic 编码工具并希望共享指令，才需要关心这个文件。

你可以用 `option global-context-path` 自定义这些路径。重复该命令可添加多个路径：

```bash
# 加载单个 Markdown 文件。
option global-context-path "~/path/to/custom/context/file.md"

# 递归加载文件夹中的所有 Markdown 文件。
option global-context-path "/full/path/to/folder/of/files/"
```

### 忽略文件

Crush 默认遵循 `.gitignore` 文件，但你也可以创建 `.crushignore` 文件来指定 Crush 应忽略的额外文件和目录。这对于排除你希望纳入版本控制、但不希望 Crush 在提供上下文时考虑的文件很有用。

`.crushignore` 文件使用与 `.gitignore` 相同的语法，可放在项目根目录或子目录中。

### 允许工具

默认情况下，Crush 在运行工具调用前会请求权限。如果你愿意，可以允许工具无需提示即可执行。请谨慎使用。

```bash
permissions allow view ls grep edit mcp_context7_get-library-doc
```

### 禁用内置工具

你也可以拒绝工具，让 agent 完全看不到它们：

```bash
permissions deny bash sourcegraph
```

要禁用 MCP 服务器中的工具，参见 [MCP 配置章节](#mcp)。

### 构建与计划模式

Crush 有两种 agent 模式：**build** 和 **plan**。

**构建模式（build）**是默认模式。agent 可以完整访问其工具集——读写、运行命令——并在有状态变更的工具调用前请求权限，与往常一样。

**计划模式（plan）**是为"先调研后动手"工作流设计的权限缩减。它把 agent 限制为严格只读的工具（view、grep、glob、ls、fetch、sourcegraph、todos、question、LSP 查询、MCP 资源读取等），外加 `present_plan` 工具。写类工具（bash、edit、write、download 等）对模型完全隐藏，任何仍然到达权限层的写入尝试都会被直接拒绝而非提示。

在计划模式下，agent 以只读方式调研你的代码库，准备好后调用 `present_plan` 交给你一份 Markdown 计划以供批准。然后你可以：

- **执行（Execute）**——批准计划并退出计划模式，agent 立即开始实施
- **继续规划（Continue planning）**——保持在计划模式，继续完善计划
- **取消（Cancel）**——不执行，直接关闭计划

在 TUI 中按 <kbd>shift+tab</kbd> 切换计划模式（或通过命令面板 <kbd>ctrl+p</kbd>）。编辑器提示框的 `>` 箭头反映当前模式，计划模式与构建模式下的箭头颜色不同。你也可以运行一次非交互式规划会话：

```bash
crush run --plan "调研代码库并提出重构计划"
```

### 提示词优化

不知道怎么组织你的需求？先在输入框里起草，然后按 <kbd>ctrl+e</kbd>。Crush 会通过一次独立的 LLM 调用把草稿改写成更清晰、更可执行的提示词——改写结合你的项目上下文（工作目录、git 状态、上下文文件），如果会话已开始，还会带上最近的对话内容以消解"这个 bug"之类的模糊指代。

确认对话框会并排展示原始提示词与优化后的提示词：按 <kbd>enter</kbd>（或 <kbd>y</kbd>）将改写结果应用到输入框，或按 <kbd>esc</kbd> 保留原文。优化在后台进行，期间可以继续输入；状态栏的旋转指示器会显示它正在进行。

默认情况下优化使用你当前的对话模型。要改用其他模型，在模型选择器（<kbd>ctrl+l</kbd>，<kbd>tab</kbd> 切换槽位）中选择"提示词优化"槽位，或在 `crushrc` 中配置 `model optimize <provider>/<model>`。

### 主题

Crush 为 TUI 内置了多套配色主题。默认情况下主题跟随所选模型的提供商（例如 Hyper 使用专属主题），但你也可以在**主题中心**中显式选择任意主题——它在命令面板（<kbd>ctrl+p</kbd>）中，以实时预览各主题强调色的方式展示。按 <kbd>enter</kbd> 立即应用主题，或选择**自动（Auto）**恢复为跟随提供商。

也可以在配置中设置主题：

```json
{
  "options": {
    "tui": {
      "theme": "tokyonight"
    }
  }
}
```

可用的主题如下：

| Key                  | 说明                     |
| -------------------- | ------------------------ |
| `pantera`            | Charmtone Pantera（默认）|
| `catppuccin-mocha`   | Catppuccin Mocha         |
| `gruvbox-dark`       | Gruvbox Dark             |
| `tokyonight`         | Tokyo Night              |
| `nord`               | Nord                     |
| `dracula`            | Dracula                  |
| `one-dark`           | One Dark                 |
| `ayu-dark`           | Ayu Dark                 |
| `vesper`             | Vesper                   |
| `rose-pine`          | Rosé Pine                |
| `kanagawa`           | Kanagawa                 |
| `everforest-dark`    | Everforest Dark          |
| `solarized-dark`     | Solarized Dark           |
| `monokai`            | Monokai                  |
| `github-dark`        | GitHub Dark              |
| `ayu-mirage`         | Ayu Mirage               |
| `night-owl`          | Night Owl                |
| `catppuccin-latte`   | Catppuccin Latte（浅色）  |
| `solarized-light`    | Solarized Light          |
| `rose-pine-dawn`     | Rosé Pine Dawn（浅色）    |
| `github-light`       | GitHub Light             |

当终端背景为透明时，Crush 会检测所选主题与你的实际终端背景（深色/浅色）的对比度，并在主题中心给出提示，方便你选择对比度合适的主题。

### UI 语言

TUI 已实现本地化，内置英语和简体中文两套语言包。打开命令面板（<kbd>ctrl+p</kbd>），选择**语言（Language）**并挑选你的偏好语言——界面会立即重新渲染并记住选择。未翻译完整的文案会回退到英语，界面绝不会显示空白。

也可以在配置中设置语言：

```json
{
  "options": {
    "tui": {
      "locale": "zh-CN"
    }
  }
}
```

支持的语言：`en`（默认）和 `zh-CN`。

### 只活一次（--yolo）

你也可以通过 `--yolo` 标志完全跳过所有权限提示。使用此功能请千万小心。

### 禁用 Skills

你可以完全阻止 Crush 使用某些 skill。被禁用的 skill 对 agent 隐藏，包括内置 skill 和从磁盘发现的 skill。

```bash
option disable-skill crush-config
```

### Agent Skills

Crush 支持 [Agent Skills](https://agentskills.io) 开放标准，用可复用的 skill 包扩展 agent 能力。Skill 是包含 `SKILL.md` 文件的文件夹，其中带有 Crush 可以发现并按需激活的指令。

我们查找 skill 的全局路径：

- `$CRUSH_SKILLS_DIR`
- `$XDG_CONFIG_HOME/agents/skills` 或 `~/.config/agents/skills/`
- `$XDG_CONFIG_HOME/crush/skills` 或 `~/.config/crush/skills/`
- `~/.agents/skills/`
- `~/.claude/skills/`
- 在 Windows 上，我们还会查找
  - `%LOCALAPPDATA%\agents\skills\` 或 `%USERPROFILE%\AppData\Local\agents\skills\`
  - `%LOCALAPPDATA%\crush\skills\` 或 `%USERPROFILE%\AppData\Local\crush\skills\`
- 通过 `options.skills_paths` 配置的其他路径

此外，我们还会从以下项目相对路径加载 skill：

- `.agents/skills`
- `.crush/skills`
- `.claude/skills`
- `.cursor/skills`

或者在配置中指定 skill 目录：

```bash
option skill-path "$HOME/squid-skills" "./other-skills"
```

你可以从 [anthropics/skills](https://github.com/anthropics/skills) 的示例 skill 开始：

```bash
# Unix
mkdir -p ~/.config/crush/skills
cd ~/.config/crush/skills
git clone https://github.com/anthropics/skills.git _temp
mv _temp/skills/* . && rm -rf _temp
```

```powershell
# Windows（PowerShell）
mkdir -Force "$env:LOCALAPPDATA\crush\skills"
cd "$env:LOCALAPPDATA\crush\skills"
git clone https://github.com/anthropics/skills.git _temp
mv _temp/skills/* . ; rm -r -force _temp
```

#### 用户可调用 Skills

Skill 可以作为命令从命令面板（<kbd>ctrl+p</kbd>）调用。在 skill 的 YAML frontmatter 中添加 `user-invocable: true`：

```yaml
---
name: my-hot-skill
description: 可作为命令调用的 skill。
user-invocable: true
---
```

用户可调用的 skill 会在命令面板中以 `user:` 或 `project:` 前缀显示：

- 来自全局目录的 skill 显示为 `user:skill-name`
- 来自项目目录的 skill 显示为 `project:skill-name`

调用时，skill 的指令会被加载到对话上下文中。

要阻止模型自动触发某个 skill（同时仍允许用户调用），添加 `disable-model-invocation: true`：

```yaml
---
name: my-skill
description: 仅可由用户调用，模型不可。
user-invocable: true
disable-model-invocation: true
---
```

带有 `disable-model-invocation` 的 skill 不会出现在模型的可用 skill 列表中，但仍可被用户手动调用。

### 桌面通知

当工具调用需要权限以及 agent 完成回合时，Crush 会发送桌面通知。只有在终端窗口未聚焦且你的终端支持上报焦点状态时才会发送。

```bash
# 选择 auto、native、osc、bell 或 disabled。
option notifications disabled
```

`auto` 在本地使用原生通知，在 SSH 上（受支持时）使用 OSC 通知。

### 初始化

初始化项目时，Crush 会分析你的代码库并创建一个上下文文件，帮助它在未来的会话中更有效地工作。默认情况下，该文件名为 `AGENTS.md`，但你可以用 `initialize-as` 选项自定义名称和位置：

```bash
# crushrc
option initialize-as AGENTS.md
```

如果你偏好不同的命名约定，或想放在特定目录（例如 `CRUSH.md` 或 `docs/LLMs.md`），这会很有用。Crush 会用初始化期间发现的构建命令、代码模式、约定等项目特定上下文填充该文件。

### 署名设置

默认情况下，Crush 会为它创建的 Git 提交和 Pull Request 添加署名信息。你可以用 `option` 命令自定义：

```bash
option attribution-trailer-style co-authored-by
option attribution-generated-with true
```

- `trailer_style`：控制添加到提交消息的署名尾注（默认：`assisted-by`）
  - `assisted-by`：按[该约定](https://docs.kernel.org/process/coding-assistants.html#attribution)添加 `Assisted-by: Crush:[ModelID]`
  - `co-authored-by`：添加 `Co-Authored-By: Crush <crush@charm.land>`
  - `none`：不添加署名尾注
- `generated_with`：为 true（默认）时，在提交消息和 PR 描述中添加 `💘 Generated with Crush` 一行

### 自定义提供商

Crush 支持针对 OpenAI 兼容和 Anthropic 兼容 API 的自定义提供商配置。

> [!NOTE]
> 请注意，我们为 OpenAI 支持两种"类型"。请务必选择正确的类型以获得最佳体验！
>
> - 通过 OpenAI 代理或路由请求时，应使用 `openai`。
> - 使用具有 OpenAI 兼容 API 的非 OpenAI 提供商时，应使用 `openai-compat`。

#### OpenAI 兼容 API

下面是使用 OpenAI 兼容 API 的 Deepseek 配置示例。别忘了在环境中设置 `DEEPSEEK_API_KEY`。

```bash
provider add deepseek --type openai-compat \
  --base-url "https://api.deepseek.com/v1" \
  --api-key "$DEEPSEEK_API_KEY"

model add deepseek/deepseek-chat \
  --name "Deepseek V3" \
  --context-window 64000 \
  --default-max-tokens 5000 \
  --price-input 0.27 \
  --price-output 1.1 \
  --price-cache-create 1.1 \
  --price-cache-hit 0.07
```

#### Anthropic 兼容 API

自定义 Anthropic 兼容提供商格式如下：

```bash
provider add custom-anthropic \
  --type anthropic \
  --base-url "https://api.anthropic.com/v1" \
  --api-key "$ANTHROPIC_API_KEY" \
  --extra-header anthropic-version 2023-06-01

model add custom-anthropic/claude-sonnet-4-20250514 \
  --name "Claude Sonnet 4" \
  --context-window 200000 \
  --default-max-tokens 50000 \
  --can-reason true \
  --supports-images true \
  --price-input 3 \
  --price-output 15 \
  --price-cache-create 3.75 \
  --price-cache-hit 0.3
```

### Amazon Bedrock

Crush 目前支持通过 Bedrock 运行 Anthropic 模型，且已禁用缓存。

一旦 Crush 能找到 AWS 凭据，Bedrock 提供商就会出现。你可以通过两种方式之一进行认证：

**API key。** 将 `AWS_BEARER_TOKEN_BEDROCK` 设置为 Bedrock API key。这是最简单的选项，会话中不会过期。

**AWS 凭据链（SSO、profiles、access keys）。** 用 `aws configure` 或 `aws configure sso` 以常规方式配置 AWS。Crush 会使用 AWS SDK 凭据链解析到的任何凭据，包括 `AWS_PROFILE`、`AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` 或 SSO 会话。要选择特定 profile，请在 shell 中设置 `AWS_PROFILE`（`AWS_PROFILE=myprofile crush`）或在顶层 [`env`](#环境变量) 配置中设置。

如果你通过 AWS SSO 认证，会话会定期过期。将 `aws_auth_refresh` 设置为刷新命令。当 Bedrock 返回凭据错误时，Crush 会运行该命令，然后在原地重试请求（不会产生重复消息，也无需手动重启）：

```json
{
  "$schema": "https://charm.land/crush.json",
  "env": {
    "AWS_PROFILE": "my-sso-profile"
  },
  "providers": {
    "bedrock": {
      "aws_auth_refresh": "aws sso login --profile my-sso-profile"
    },
    "bedrock-europe": {
      "aws_auth_refresh": "aws sso login --profile my-eu-sso-profile"
    }
  }
}
```

- `aws_auth_refresh` — AWS 凭据过期时运行的 shell 命令（例如 `aws sso login`）

### Vertex AI 平台

当设置了 `VERTEXAI_PROJECT` 和 `VERTEXAI_LOCATION` 时，Vertex AI 会出现在可用提供商列表中。你还需要完成认证：

```bash
$ gcloud auth application-default login
```

要向配置中添加特定模型，请按如下方式配置：

```bash
# crushrc — 认证仍来自 gcloud 和 VERTEXAI_* 环境变量。
provider add vertexai --type google-vertex

model add vertexai/claude-sonnet-4@20250514 \
  --name "VertexAI Sonnet 4" \
  --context-window 200000 \
  --default-max-tokens 50000 \
  --can-reason true \
  --supports-images true \
  --price-input 3 \
  --price-output 15 \
  --price-cache-create 3.75 \
  --price-cache-hit 0.3
```

### 本地模型

Crush 可以自动发现本地提供商的模型。添加 `type` 为 `llamacpp`、`omlx`、`lmstudio`、`litellm` 或 `ollama` 的自定义提供商，并留空模型列表。Crush 会自动填充模型列表。

```bash
# 小菜一碟。
provider add ollama \
  --name Ollama \
  --type ollama \
  --base-url "http://localhost:11434/v1/"
```

对于 llama.cpp（`llama-server`），指向服务器的 base URL：

```bash
provider add llamacpp \
  --name "llama.cpp" \
  --type llamacpp \
  --base-url "http://localhost:2222"
```

#### 手动模型配置

你仍然可以显式列出模型。用户定义的模型总是优先于自动发现的模型，你设置的任何字段都不会被自动发现覆盖。如果任何 `openai-compat` 提供商的模型列表为空，将自动运行发现；或者传入 `"discover_models": true`，会将发现的模型与你手工配置的模型合并。

```bash
# crushrc
provider add ollama \
  --name Ollama \
  --type ollama \
  --base-url "http://localhost:11434/v1/" \
  --discover-models true

model add ollama/qwen3:30b \
  --name "Qwen 3 30B" \
  --context-window 256000 \
  --default-max-tokens 20000
```

`--discover-models true` 标志会将发现的模型与上面的模型合并；发生冲突时，你显式配置的模型字段优先。

## 日志

有时你需要查看日志。幸运的是，Crush 会记录各种信息。日志存储在项目相对的 `./.crush/logs/crush.log` 中。

CLI 还提供一些辅助命令，方便你翻阅最近的日志：

```bash
# 打印最后 1000 行
crush logs

# 打印最后 500 行
crush logs --tail 500

# 实时跟踪日志
crush logs --follow
```

想要更多日志？用 `--debug` 标志运行 `crush`，或在 `crushrc` 中启用：

```bash
# crushrc
option debug true
option debug-lsp true
```

## 提供商自动更新

默认情况下，Crush 会自动从开源 Crush 提供商数据库 [Catwalk](https://github.com/charmbracelet/catwalk) 检查最新的提供商和模型列表。这意味着当有新的提供商和模型可用，或模型元数据发生变化时，Crush 会自动更新你的本地配置。

### 自定义提供商目录

你也可以覆盖 [Catwalk](https://github.com/charmbracelet/catwalk) 的默认 URL（例如用于测试或使用 fork）。

设置 `CATWALK_URL` 环境变量即可（例如 `export CATWALK_URL=http://localhost:8000`）。

### 禁用自动提供商更新

对于网络受限或偏好离线环境的人，这个功能可能不是你想要的，可以禁用它。

在 `crushrc` 中禁用自动提供商更新：

```bash
option provider-auto-update false
```

或设置 `CRUSH_DISABLE_PROVIDER_AUTO_UPDATE` 环境变量：

```bash
export CRUSH_DISABLE_PROVIDER_AUTO_UPDATE=1
```

### 手动更新提供商

可以用 `crush update-providers` 命令手动更新提供商：

```bash
# 从 Catwalk 远程更新提供商。
crush update-providers

# 从自定义 Catwalk base URL 更新提供商。
crush update-providers https://example.com/

# 从本地文件更新提供商。
crush update-providers /path/to/local-providers.json

# 将提供商重置为构建时嵌入的版本。
crush update-providers embedded

# 更多信息：
crush update-providers --help
```

## 指标

Crush 会记录假名化的使用指标（关联到设备特定哈希），维护者依赖这些指标来确定开发和支持优先级。指标仅包含使用元数据；提示词和响应**绝不会**被收集。

具体收集内容的细节见源代码（[这里](https://github.com/f69wrthhqd-blip/crush/tree/main/internal/event)和[这里](https://github.com/f69wrthhqd-blip/crush/blob/main/internal/event/event.go)）。

你可以在任何时候选择退出指标收集，在环境中设置：

```bash
export CRUSH_DISABLE_METRICS=1
```

Crush 还遵循 [`DO_NOT_TRACK`](https://donottrack.sh/) 约定，可通过 `export DO_NOT_TRACK=1` 启用。

## 常见问题

### 为什么剪贴板复制粘贴不工作？

在 Unix 类环境中可能需要安装额外工具。

| 环境                 | 工具                     |
| -------------------- | ------------------------ |
| Windows              | 原生支持                 |
| macOS                | 原生支持                 |
| Linux/BSD + Wayland  | `wl-copy` 和 `wl-paste`  |
| Linux/BSD + X11      | `xclip` 或 `xsel`        |

## 参与贡献

本仓库是上游 [Crush](https://github.com/charmbracelet/crush) 项目的 fork。不属于本 fork 特有的改进，建议直接贡献到上游，参见[上游贡献指南](https://github.com/charmbracelet/crush?tab=contributing-ov-file#contributing)。fork 特有的改动可通过向本仓库提交 PR 提出。

## 你的想法？

我们很想听听你对该项目的看法。需要帮助？包在我们身上。你可以通过以下方式找到我们：

- [Twitter](https://twitter.com/charmcli)
- [Slack][slack]
- [Discord][discord]
- [The Fediverse](https://mastodon.social/@charmcli)
- [Bluesky](https://bsky.app/profile/charm.land)

[slack]: https://charm.land/slack
[discord]: https://charm.land/discord

## 许可证

[FSL-1.1-MIT](https://github.com/f69wrthhqd-blip/crush/raw/main/LICENSE.md)

---

[Crush](https://github.com/charmbracelet/crush) 的 fork，原项目为 [Charm](https://charm.land) 出品的终端 AI 编程助手。上游仓库：[charmbracelet/crush](https://github.com/charmbracelet/crush)。

<a href="https://charm.land/"><img alt="The Charm logo" width="400" src="https://stuff.charm.sh/charm-banner-softy.jpg" /></a>

<!--prettier-ignore-->
Charm热爱开源 • Charm loves open source
