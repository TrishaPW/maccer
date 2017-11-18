# Maccer

The Bay Area Roleplay Discord Bot

---

## Development

The code is derived from CJ, the SA:MP Discord bot. It's a bit messy and has some features left over that are unused but it's quite simple.

```bash
go get github.com/Southclaws/maccer
```

All environment variables are inside the Makefile apart from the two secrets: `FORUM_KEY` which is the Invision Community API key and `DISCORD_TOKEN` which is the bot's Discord token (not the app token, the bot token!). So make sure you create an `.env` file to contain these:

```bash
echo FORUM_KEY=abc123 >> .env
echo DISCORD_TOKEN=xyz789 >> .env
```

Then you can hack away and run locally with

```make
make local
```

Docker is my deployment method. To build the image:

```make
make build
```

To run it locally:

```make
make run
```

And (you'll probably never need this but in case I go M.I.A. or something)

To deploy it on a server:

```make
make run-prod
```
