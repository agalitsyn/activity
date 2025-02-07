# Activity

This is POC of activity tracking system.


## Local run

Copy config template:

```sh
cp .env.example .env
```

Fill config variables values in `.env`

Then run needed app:

```sh
make run-server
# OR
make run-agent
```

## Architecture

TBD

### Scaling

TBD

## What was uncovered?

### Agent

- Better MacOS intergration (maybe rewrite in Swift?)
    - Apps introspection
    - Menubar
    - Settings
    - Updates
- More platforms
    - Windows
    - Linux
- Binary self-protection
- Handling websocket reconnections
- Send parts of data from log after being offline

### Server

- Auth
- Secure connection (wss)
- Blocking and removing agents
- Performance testing
