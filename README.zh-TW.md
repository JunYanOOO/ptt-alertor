# Ptt-Alertor

[English](README.md) | 繁體中文

<img align="right" src="https://raw.githubusercontent.com/Ptt-Alertor/ptt-alertor/master/logo.jpg">

[![Build Status](https://github.com/Ptt-Alertor/ptt-alertor/actions/workflows/main.yml/badge.svg)](https://github.com/Ptt-Alertor/ptt-alertor/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/Ptt-Alertor/ptt-alertor/branch/master/graph/badge.svg)](https://codecov.io/gh/Ptt-Alertor/ptt-alertor)
[![Go Report Card](https://goreportcard.com/badge/github.com/Ptt-Alertor/ptt-alertor)](https://goreportcard.com/report/github.com/Ptt-Alertor/ptt-alertor)
[![Code Climate](https://api.codeclimate.com/v1/badges/f7047295fce56a0465dc/maintainability)](https://codeclimate.com/github/Ptt-Alertor/ptt-alertor/maintainability)
[![StackShare](https://img.shields.io/badge/tech-stack-0690fa.svg?style=flat)](https://stackshare.io/ptt-alertor/ptt-alertor)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## 使用 Docker Compose 啟動

Docker Compose 會啟動應用程式、Redis 與 DynamoDB Local。啟動時，應用程式會等待 DynamoDB 就緒，並在 `boards` 與 `articles` 資料表不存在時自動建立。

```powershell
Copy-Item .env.example .env
docker compose up --build
```

應用程式啟動後，開啟 <http://localhost:19090>。Redis 與 DynamoDB 資料會儲存在 Docker 命名磁碟區中。

所有通知整合皆為選用。LINE Bot 需要同時設定 `LINE_CHANNEL_SECRET` 與 `LINE_CHANNEL_ACCESSTOKEN`。若要啟用 Telegram，請在 `.env` 設定 `TELEGRAM_TOKEN`，並將 `APP_HOST` 設為可從網際網路存取的 HTTPS 網址；應用程式會將 `${APP_HOST}/telegram/${TELEGRAM_TOKEN}` 註冊為 webhook。

追蹤功能由專案根目錄的 `config.yaml` 控制。預設只啟用關鍵字追蹤，每個選項皆附有繁體中文註解；修改後必須重新啟動應用程式。若設定檔遺失、不完整、格式錯誤或含有未知欄位，應用程式會拒絕啟動，避免意外啟用原本停用的功能。

停止服務：

```powershell
docker compose down
```

只有在確定要一併刪除 Redis 與 DynamoDB 本機資料時，才使用 `docker compose down -v`。

## Discord Bot

1. 前往 [Discord Developer Portal](https://discord.com/developers/applications) 建立應用程式與 Bot，並複製 Bot Token。
2. 在 **OAuth2 > URL Generator** 選取 `bot` 與 `applications.commands` scopes，授予 Bot **View Channels** 與 **Send Messages** 權限，再使用產生的網址將 Bot 加入伺服器。
3. 將 Token 與開發階段使用的 Discord 伺服器 ID 填入 `.env`：

```dotenv
DISCORD_TOKEN=your-bot-token
DISCORD_GUILD_ID=your-server-id
```

4. 執行 `docker compose up --build` 啟動或重新建置服務。

具有 **Manage Channels** 權限的成員，可在伺服器文字頻道管理追蹤清單；通知會傳送至相同頻道：

```text
/新增 看板:gossiping 關鍵字:台積電
/清單
/刪除 看板:gossiping 關鍵字:台積電
```

訂閱資料儲存在 Redis，應用程式重新啟動後仍會保留。正式環境可將 `DISCORD_GUILD_ID` 留空，讓指令註冊為全域指令。

## API

以下為可用的 HTTP API。作者、推噓文數與單篇文章追蹤 API 只有在 `config.yaml` 啟用對應功能時才會註冊。

### 看板

- `GET /boards`
- `GET /boards/[看板名稱]/articles`
- `GET /boards/[看板名稱]/articles/[文章代碼]`

### 關鍵字

- `GET /keyword/boards`

### 作者

- `GET /author/boards`

### 推噓文數

- `GET /pushsum/boards`

### 單篇文章

- `GET /articles`

### 使用者（需要 Basic Auth）

- `GET /users`
- `GET /users/[帳號]`
- `POST /users`

```json
{
    "profile": {
        "account": "sample",
        "email": "sample@mail.com"
    },
    "subscribes": [
        {
            "board": "gossiping",
            "keywords": ["問卦", "爆卦", "公告"]
        },
        {
            "board": "lol",
            "keywords": ["閒聊"]
        }
    ]
}
```

- `PUT /users/[帳號]`

```json
{
    "profile": {
        "account": "sample",
        "email": "sample@mail.com"
    },
    "subscribes": []
}
```

## 致謝

### 現實生活

Rose Li、Aries Huang、Scott Kao、Amy Li

### PTT

DMM、bestpika、Zero0910、lucky0509、wbreeze、chang0206、lindo0130、hungys、gyman7788、tooilxui、myamyakoko、whkuo、papago89、timeline、Kamikiri

### Facebook

Mr.clu、Woqeker
