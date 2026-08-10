# StyleTailor MCP Server

A PocketBase-based MCP (Model Context Protocol) server for personalized virtual try-on and fashion styling. All AI operations are delegated to the Bailian CLI (`bl`) so no Python or local GPU is required.

## Features

- **MCP tools** exposed over HTTP:
  - `styletailor_generate_look` – generate a virtual try-on look from user portrait, body data and preference, using hierarchical negative feedback.
  - `styletailor_feedback` – submit user feedback to trigger an iteration.
  - `styletailor_get_result` – fetch the status and result URLs of a request.
- **All model calls** go through the Bailian CLI (`bl text chat`, `bl image generate`, etc.).
- **Written in Go** using a PocketBase backend.
- **Negative-feedback loop**: generated images are scored by a vision/text model and the prompt is revised until the score passes a threshold (or max iterations reached).

## Build

```bash
source /home/cuizheng/.local/go_env.sh
cd /home/cuizheng/Projects/styletailor-mcp
go build -o /tmp/styletailor-mcp ./styletailor
```

## Run

```bash
/tmp/styletailor-mcp serve --http=0.0.0.0:8090
```

## Test

```bash
# List tools
curl -X POST http://127.0.0.1:8090/mcp/tools/list

# Generate a look
curl -X POST http://127.0.0.1:8090/mcp/tools/call \
  -H 'Content-Type: application/json' \
  -d '{
    "tool": "styletailor_generate_look",
    "args": {
      "user_id": "u-123",
      "body_data": "{\"height_cm\":170,\"weight_kg\":60,\"chest\":\"88cm\",\"waist\":\"66cm\",\"hips\":\"90cm\"}",
      "portrait_url": "https://example.com/portrait.png",
      "occasion": "daily office",
      "preference": "elegant minimalist, neutral colors, comfortable"
    }
  }'
```

## Architecture

```
User / AI Agent
      |
      v
MCP HTTP endpoints (/mcp/tools/*)
      |
      v
PocketBase server (Go)
      |
      +-- Bailian CLI for chat/vision/image generation
      +-- (future) PocketBase collections for users, products, requests, feedback
```

## Next Steps

- Replace the in-memory `queryCatalog`/`persistRequest` with real PocketBase collections.
- Add vision evaluation (`bl vision describe`) for scoring generated looks.
- Add reference-image workflows for garment transfer and multi-view generation.
- Add SSE/gRPC streaming MCP transport.
