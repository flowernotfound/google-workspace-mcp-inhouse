# google-workspace-mcp-inhouse

Google Docs・Google Sheets 読み取り専用 MCP サーバー。AIエージェントが Google Docs と Google Sheets の内容を参照するためのツールです。

---

## 提供ツール一覧

### Google Docs

| ツール名            | 概要                                                           |
| ------------------- | -------------------------------------------------------------- |
| `read_document`     | ドキュメント本文を Markdown またはプレーンテキストで取得       |
| `list_documents`    | Google Drive 内のドキュメント一覧を取得                        |
| `search_documents`  | キーワードでドキュメントを検索                                 |
| `get_document_info` | ドキュメントのメタ情報（タイトル・オーナー・更新日時等）を取得 |
| `list_comments`     | ドキュメントのコメント一覧を取得                               |
| `get_comment`       | 個別コメントと返信スレッドを取得                               |

### Google Sheets

| ツール名               | 概要                                                                 |
| ---------------------- | -------------------------------------------------------------------- |
| `read_spreadsheet`     | スプレッドシートの内容を CSV または JSON 形式で取得                  |
| `get_spreadsheet_info` | スプレッドシートのメタ情報（タイトル・シート一覧・ロケール等）を取得 |
| `list_spreadsheets`    | Google Drive 内のスプレッドシート一覧を取得                          |
| `search_spreadsheets`  | キーワードでスプレッドシートを検索                                   |
| `get_sheet_range`      | 指定範囲（A1 記法）のセル値を取得                                    |

<details>
<summary>各ツールのパラメータ詳細</summary>

#### `read_document`

| パラメータ    | 型     | 必須 | デフォルト   | 説明                                     |
| ------------- | ------ | ---- | ------------ | ---------------------------------------- |
| `document_id` | string | Yes  | —            | Google Docs のドキュメント ID            |
| `format`      | string | No   | `"markdown"` | 出力形式（`"markdown"` または `"text"`） |

#### `list_documents`

| パラメータ    | 型     | 必須 | デフォルト            | 説明                                        |
| ------------- | ------ | ---- | --------------------- | ------------------------------------------- |
| `folder_id`   | string | No   | —                     | 特定フォルダ内のみ取得する場合のフォルダ ID |
| `max_results` | number | No   | `20`                  | 最大取得件数（上限: 100）                   |
| `order_by`    | string | No   | `"modifiedTime desc"` | 並び順                                      |

#### `search_documents`

| パラメータ    | 型     | 必須 | デフォルト | 説明                     |
| ------------- | ------ | ---- | ---------- | ------------------------ |
| `query`       | string | Yes  | —          | 検索キーワード           |
| `max_results` | number | No   | `10`       | 最大取得件数（上限: 50） |

#### `get_document_info`

| パラメータ    | 型     | 必須 | デフォルト | 説明                          |
| ------------- | ------ | ---- | ---------- | ----------------------------- |
| `document_id` | string | Yes  | —          | Google Docs のドキュメント ID |

#### `list_comments`

| パラメータ         | 型      | 必須 | デフォルト | 説明                          |
| ------------------ | ------- | ---- | ---------- | ----------------------------- |
| `document_id`      | string  | Yes  | —          | Google Docs のドキュメント ID |
| `include_resolved` | boolean | No   | `false`    | 解決済みコメントを含めるか    |

#### `get_comment`

| パラメータ    | 型     | 必須 | デフォルト | 説明                          |
| ------------- | ------ | ---- | ---------- | ----------------------------- |
| `document_id` | string | Yes  | —          | Google Docs のドキュメント ID |
| `comment_id`  | string | Yes  | —          | コメント ID                   |

#### `read_spreadsheet`

| パラメータ       | 型     | 必須 | デフォルト | 説明                                     |
| ---------------- | ------ | ---- | ---------- | ---------------------------------------- |
| `spreadsheet_id` | string | Yes  | —          | Google Sheets のスプレッドシート ID      |
| `sheet_name`     | string | No   | —          | 取得するシート名（省略時は最初のシート） |
| `format`         | string | No   | `"csv"`    | 出力形式（`"csv"` または `"json"`）      |

#### `get_spreadsheet_info`

| パラメータ       | 型     | 必須 | デフォルト | 説明                                |
| ---------------- | ------ | ---- | ---------- | ----------------------------------- |
| `spreadsheet_id` | string | Yes  | —          | Google Sheets のスプレッドシート ID |

#### `list_spreadsheets`

| パラメータ    | 型     | 必須 | デフォルト            | 説明                                        |
| ------------- | ------ | ---- | --------------------- | ------------------------------------------- |
| `folder_id`   | string | No   | —                     | 特定フォルダ内のみ取得する場合のフォルダ ID |
| `max_results` | number | No   | `20`                  | 最大取得件数（上限: 100）                   |
| `order_by`    | string | No   | `"modifiedTime desc"` | 並び順                                      |

#### `search_spreadsheets`

| パラメータ    | 型     | 必須 | デフォルト | 説明                     |
| ------------- | ------ | ---- | ---------- | ------------------------ |
| `query`       | string | Yes  | —          | 検索キーワード           |
| `max_results` | number | No   | `10`       | 最大取得件数（上限: 50） |

#### `get_sheet_range`

| パラメータ       | 型     | 必須 | デフォルト | 説明                                       |
| ---------------- | ------ | ---- | ---------- | ------------------------------------------ |
| `spreadsheet_id` | string | Yes  | —          | Google Sheets のスプレッドシート ID        |
| `range`          | string | Yes  | —          | A1 記法での範囲指定（例: `Sheet1!A1:D10`） |

</details>

---

## インストール

### Mac / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/flowernotfound/google-workspace-mcp-inhouse/master/install.sh | bash
```

バイナリは `~/bin/google-workspace-mcp-inhouse` に配置されます。

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/flowernotfound/google-workspace-mcp-inhouse/master/install.ps1 | iex
```

バイナリは `%LOCALAPPDATA%\Programs\google-workspace-mcp-inhouse\` に配置されます。
インストーラが自動的にユーザー `PATH` へ追加します。

---

## セットアップ

### 1. `credentials.json` を配置する

GCP コンソールから `credentials.json` をダウンロードして、以下のパスに配置します。

```bash
# Mac / Linux
mv ~/Downloads/credentials.json ~/.config/google-workspace-mcp-inhouse/credentials.json

# Windows (PowerShell)
Move-Item ~\Downloads\credentials.json ~\.config\google-workspace-mcp-inhouse\credentials.json
```

> `credentials.json` の取得方法は「[GCP 初期設定（管理者）](#gcp-初期設定管理者-1-回のみ)」を参照してください。

### 2. 個人認証を行う

```bash
# Mac / Linux
google-workspace-mcp-inhouse auth

# Windows
google-workspace-mcp-inhouse.exe auth
```

ブラウザが開くので、自分の Google アカウントでログインして読み取り権限を許可します。認証情報は `~/.config/google-workspace-mcp-inhouse/token.json` に保存されます。

### 3. Claude Code に登録する

`.mcp.json` に以下を追記します。

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

> **Windows の場合：** `${HOME}/bin/google-workspace-mcp-inhouse` の代わりに、フルパスを指定してください。
> 例: `C:\Users\yourname\AppData\Local\Programs\google-workspace-mcp-inhouse\google-workspace-mcp-inhouse.exe`

---

## アップデート

```bash
# Mac / Linux
google-workspace-mcp-inhouse update

# Windows — インストーラを再実行
irm https://raw.githubusercontent.com/flowernotfound/google-workspace-mcp-inhouse/master/install.ps1 | iex
```

---

## GCP 初期設定（管理者、1 回のみ）

各エンジニアがインストールする前に、管理者が GCP プロジェクトをセットアップする必要があります。

### 1. GCP プロジェクトを作成する

組織の GCP 配下に MCP 専用プロジェクト（例: `google-workspace-mcp`）を作成します。

### 2. API を有効化する

以下の 3 つの API を有効化します。

- Google Docs API
- Google Drive API
- Google Sheets API

### 3. OAuth 同意画面を設定する

| 項目     | 設定値                                                          |
| -------- | --------------------------------------------------------------- |
| 公開範囲 | **内部**（組織メンバーのみ。外部からの不正利用を防止）          |
| スコープ | `documents.readonly`、`drive.readonly`、`spreadsheets.readonly` |

### 4. OAuth クライアント ID を発行する

- 種類: **デスクトップアプリ**
- 発行後、`credentials.json` をダウンロードする

### 5. IAM 権限を付与する

利用エンジニアに GCP プロジェクトの **Viewer** ロールを付与します。Viewer 権限があれば、各エンジニアが GCP コンソールから `credentials.json` を自分でダウンロードできます（管理者への都度依頼が不要）。

```
GCP コンソール
└── APIs & Services → Credentials
    └── OAuth 2.0 クライアント ID の行にある「ダウンロード」ボタン
```

---

## 開発者向け

```bash
mise run build     # バイナリビルド
mise run test      # テスト実行
mise run lint      # リンター実行
mise run fmt       # コードフォーマット
```

[mise](https://mise.jdx.dev/) と Go 1.23+ が必要です。
