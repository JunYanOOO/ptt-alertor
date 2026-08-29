# Ptt-Alertor

<img align="right" src="https://raw.githubusercontent.com/Ptt-Alertor/ptt-alertor/master/logo.jpg">

[![Build Status](https://github.com/Ptt-Alertor/ptt-alertor/actions/workflows/main.yml/badge.svg)](https://github.com/Ptt-Alertor/ptt-alertor/actions/workflows/main.yml)
[![codecov](https://codecov.io/gh/Ptt-Alertor/ptt-alertor/branch/master/graph/badge.svg)](https://codecov.io/gh/Ptt-Alertor/ptt-alertor)
[![Go Report Card](https://goreportcard.com/badge/github.com/Ptt-Alertor/ptt-alertor)](https://goreportcard.com/report/github.com/Ptt-Alertor/ptt-alertor)
[![Code Climate](https://api.codeclimate.com/v1/badges/f7047295fce56a0465dc/maintainability)](https://codeclimate.com/github/Ptt-Alertor/ptt-alertor/maintainability)
[![StackShare](https://img.shields.io/badge/tech-stack-0690fa.svg?style=flat)](https://stackshare.io/ptt-alertor/ptt-alertor)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Run with Docker Compose

Docker Compose starts the application, Redis, and DynamoDB Local. On startup,
the application waits for DynamoDB and creates the required `boards` and
`articles` tables when they do not already exist.

```powershell
Copy-Item .env.example .env
docker compose up --build
```

Open <http://localhost:19090> after the application has started. Redis and
DynamoDB data are stored in named Docker volumes.

Notification integrations are optional. LINE Bot requires both
`LINE_CHANNEL_SECRET` and `LINE_CHANNEL_ACCESSTOKEN`. To enable Telegram, set
`TELEGRAM_TOKEN` and set `APP_HOST` to a public HTTPS URL in `.env`; the
application registers `${APP_HOST}/telegram/${TELEGRAM_TOKEN}` as its webhook.

Stop the stack with:

```powershell
docker compose down
```

Use `docker compose down -v` only when you also want to delete the local Redis
and DynamoDB data.

## Discord Bot

1. Create an application and bot in the
   [Discord Developer Portal](https://discord.com/developers/applications), then
   copy the bot token.
2. In **OAuth2 > URL Generator**, select the `bot` and
   `applications.commands` scopes. Grant the bot **View Channels** and
   **Send Messages**, then use the generated URL to add it to your server.
3. Put the token and, during development, your Discord server ID in `.env`:

```dotenv
DISCORD_TOKEN=your-bot-token
DISCORD_GUILD_ID=your-server-id
```

4. Start or rebuild the stack with `docker compose up --build`.

Members with **Manage Channels** permission can manage the watch list in a
server text channel. Alerts are sent back to the same channel:

```text
/新增 看板:gossiping 關鍵字:台積電
/清單
/刪除 看板:gossiping 關鍵字:台積電
```

Subscriptions are stored in Redis and survive application restarts. Leave
`DISCORD_GUILD_ID` empty in production to register the commands globally.

## API

### Board

* GET /boards

* GET /boards/[board name]/articles

* GET /boards/[board name]/articles/[article code]

### Keyword

* GET /keyword/boards

### Author

* GET /author/boards

### PushSum

* GET /pushsum/boards

### Articles

* GET /articles

### User (Auth)

* GET /users

* GET /users/[account]

* POST /users

```json
{
    "profile":{
        "account": "sample",
        "email":"sample@mail.com"
    },
    "subscribes":[
        {
            "board":"gossiping",
            "keywords":["問卦","爆卦","公告"]
        },
        {
            "board":"lol",
            "keywords":["閒聊"]
        }
    ]
}
```

* PUT /users/[account]

```json
{
    "profile":{
        "account": "sample",
        "email":"sample@mail.com"
    },
    "subscribes":[]
}
```

## Credits

### Real Life

Rose Li, Aries Huang, Scott Kao, Amy Li

### Ptt

DMM, oas, bestpika, Zero0910, lucky0509, wbreeze, chang0206, lindo0130, hungys, gyman7788, tooilxui, myamyakoko, whkuo, papago89, timeline, Kamikiri

### Facebook

Mr.clu, Woqeker
