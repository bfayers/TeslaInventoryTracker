# Tesla Tracker - Webhook Bot

Quick script that can track approved used Tesla's for sale in the UK. Written to handle Model 3 originally but should work for all models listed below.

## Required settings
Settings are handled with env vars, in the current state you need these:
```
DISCORD_WEBHOOK_URL - Discord channel webhook url
DISCORD_NEW_CAR_THREAD - Thread ID for new cars
DISCORD_CHANGED_CAR_THREAD - Thread ID for changes to already tracked cars
MODEL - One of: ms, m3, mx, my
YEARS - Comma seperated list of years to search (eg: 2024,2025)
TRIMS - Comma seperated list of trims. See below for options
```

Available settings for `TRIMS`
- When model is `ms`
    ```
    MSPLAID,MSAWD,75D
    ```
- When model is `m3`
    ```
    PAWD,LRAWD,LRRWD,M3RWD
    ```
- When model is `mx`
    ```
    MXPERF,MXAWD,75D
    ```
- When model is `my`
    ```
    PAWD,LRAWD,LRRWD,MYRWD
    ```
