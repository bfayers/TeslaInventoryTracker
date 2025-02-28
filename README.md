# Tesla Tracker - Webhook Bot

Quick script that can track approved used Tesla's for sale in the UK. Written to handle Model 3.

## Required settings
Settings are handled with env vars, in the current state you need these:
```
DISCORD_WEBHOOK_URL - Discord channel webhook url
DISCORD_NEW_CAR_THREAD - Thread ID for new cars
DISCORD_CHANGED_CAR_THREAD - Thread ID for changes to already tracked cars
```