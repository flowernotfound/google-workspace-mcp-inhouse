# google-workspace-mcp-inhouse

> **日本語版:** [README_ja.md](README_ja.md)

A read-only MCP server for Google Docs.
Provides 6 tools to read, list, search, and comment on Google Docs — no write access.

---

## Available Tools

| Tool | Description |
|------|-------------|
| `read_document` | Fetch document body as Markdown or plain text |
| `list_documents` | List Google Docs in your Drive |
| `search_documents` | Search documents by keyword |
| `get_document_info` | Get document metadata (title, owner, timestamps) |
| `list_comments` | List comments on a document |
| `get_comment` | Get a specific comment and its reply thread |

<details>
<summary>Tool parameters</summary>

#### `read_document`
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `document_id` | string | Yes | — | Google Docs document ID |
| `format` | string | No | `"markdown"` | Output format: `"markdown"` or `"text"` |

#### `list_documents`
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `folder_id` | string | No | — | Restrict to a specific folder |
| `max_results` | number | No | `20` | Max items to return (limit: 100) |
| `order_by` | string | No | `"modifiedTime desc"` | Sort order |

#### `search_documents`
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `query` | string | Yes | — | Search keyword |
| `max_results` | number | No | `10` | Max items to return (limit: 50) |

#### `get_document_info`
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `document_id` | string | Yes | — | Google Docs document ID |

#### `list_comments`
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `document_id` | string | Yes | — | Google Docs document ID |
| `include_resolved` | boolean | No | `false` | Whether to include resolved comments |

#### `get_comment`
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `document_id` | string | Yes | — | Google Docs document ID |
| `comment_id` | string | Yes | — | Comment ID |

</details>

---

## Installation

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/flowernotfound/google-workspace-mcp-inhouse/master/install.sh | bash
```

The binary is installed to `~/bin/google-workspace-mcp-inhouse`.

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/flowernotfound/google-workspace-mcp-inhouse/master/install.ps1 | iex
```

The binary is installed to `%LOCALAPPDATA%\Programs\google-workspace-mcp-inhouse\`.
The installer automatically adds the directory to your user `PATH`.

---

## Setup

### 1. Place `credentials.json`

Obtain `credentials.json` from the GCP Console and place it in the config directory.

```bash
# macOS / Linux
mv ~/Downloads/credentials.json ~/.config/google-workspace-mcp-inhouse/credentials.json

# Windows (PowerShell)
Move-Item ~\Downloads\credentials.json ~\.config\google-workspace-mcp-inhouse\credentials.json
```

### 2. Authenticate

```bash
# macOS / Linux
google-workspace-mcp-inhouse auth

# Windows
google-workspace-mcp-inhouse.exe auth
```

A browser window opens. Sign in with your Google account to grant read-only access.

### 3. Register with Claude Code

Add the following to `.mcp.json`:

```json
{
    "mcpServers": {
        "google-workspace-mcp-inhouse": {
            "type": "stdio",
            "command": "${HOME}/bin/google-workspace-mcp-inhouse"
        }
    }
}
```

> **Windows:** Replace `${HOME}/bin/google-workspace-mcp-inhouse` with the full path, e.g.
> `C:\Users\yourname\AppData\Local\Programs\google-workspace-mcp-inhouse\google-workspace-mcp-inhouse.exe`

---

## Updating

```bash
# macOS / Linux
google-workspace-mcp-inhouse update

# Windows — re-run the installer
irm https://raw.githubusercontent.com/flowernotfound/google-workspace-mcp-inhouse/master/install.ps1 | iex
```

---

## Prerequisites (admin setup)

Before individual engineers can install this tool, an admin needs to set up the GCP project once.
See [README_ja.md — GCP 初期設定](README_ja.md#gcp-初期設定管理者-1-回のみ) for the full setup guide.

---

## Development

```bash
mise run build     # build binary
mise run test      # run tests
mise run lint      # run linter
mise run fmt       # format code
```

Requires [mise](https://mise.jdx.dev/) and Go 1.23+.
