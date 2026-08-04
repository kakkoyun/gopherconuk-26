# gopherconuk-26

GopherCon UK 2026 talks.

## Creative MCP tools

The project-level `.mcp.json` gives Claude Code a small media toolkit for
writing and designing talks.

| Server | Use |
| --- | --- |
| `meme` | Search meme templates and create captioned images |
| `giphy` | Search rated GIFs and stickers |
| `readwise` | Search highlights and Reader documents |
| `fal` | Find generative-media models and run them when requested |
| `pexels` | Search stock photos and videos |
| `pixabay` | Search stock photos, illustrations, vectors, and videos |

The credential-backed servers require Node.js 20 or newer and the 1Password
CLI. Their API keys remain as `op://` references and are resolved only when each
server starts. Clones outside the author's 1Password account must replace those
references and the `--account` value.

Claude Code asks for approval before starting project MCP servers. Run `claude`,
open `/mcp`, approve the servers, and authenticate Readwise. `claude mcp list`
shows their connection status.

Download selected media into `talks/<talk>/slides/assets/` instead of depending
on remote URLs. Keep the provider attribution with the asset. Animated GIFs work
in HTML decks, but PDF exports capture a static frame.

fal model execution can incur charges. Check pricing before running a model.
