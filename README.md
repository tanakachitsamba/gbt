# GBT Assistant Service

## Overview
The GBT Assistant service provides a thin HTTP layer on top of OpenAI's APIs so that product teams can prototype assistant-style workflows locally. It exposes REST endpoints that mimic the structure of the official Assistants API while delegating model calls to OpenAI's chat completions interface. The implementation focuses on rapid iteration: request validation, deterministic mock resources (threads, assistants, vector stores), and consistent response envelopes make it easy to plug the service into front-end experiments or automated tests.

## Architecture
At a high level the service is composed of three pieces:

1. **HTTP Server (`main.go`, `server.go`)** – Boots a Gorilla Mux router with JSON handlers for each REST resource. Common helpers manage JSON decoding/encoding and error normalization so handlers can stay focused on business logic.
2. **OpenAI Wrapper (`openai_wrapper.go`)** – Encapsulates every interaction with the `go-openai` SDK. It assembles chat completion payloads, translates OpenAI responses into the service's DTOs, and fabricates IDs/metadata for local-only resources such as threads and vector stores.
3. **Supporting Utilities (`pkg/tokenizer`, `docs/openapi.json`, misc. helpers)** – Utility packages and fixtures supply reusable logic (e.g., token counting) and documentation. The bundled OpenAPI document keeps the HTTP contract versioned alongside the codebase and powers schema-aware clients.

The `main` entry point wires these components together: it loads environment variables, instantiates the OpenAI client, wraps it, mounts the API routes, and serves requests behind a CORS layer. Each handler delegates to the wrapper, which either forwards calls to OpenAI (for `/v1/responses`) or synthesizes domain objects (for `/v1/threads`, `/v1/assistants`, `/v1/vector-stores`).

## API Documentation
- **OpenAPI specification:** [`docs/openapi.json`](docs/openapi.json)
- **Generated Swagger endpoint:** `GET /openapi.json` (alias `GET /swagger.json`)

You can import the JSON file into tools such as Postman, Stoplight, or VS Code's REST client to explore the available operations and payload shapes.

## Setup
1. Install Go 1.21 or newer.
2. Clone the repository and move into the project directory.
3. Provide an OpenAI API key via the `OPENAI_KEY` environment variable. For local development you can create a `.env` file with `OPENAI_KEY=sk-...`.
4. Download Go module dependencies:
   ```bash
   go mod download
   ```

## Running the Server
Start the HTTP server with:
```bash
go run ./...
```
The service listens on `http://localhost:8080` and enables CORS for front-ends served from `http://localhost:3000`. Adjust `main.go` if you need to broaden the allowed origins or headers.

## Testing
Run the full Go test suite:
```bash
go test ./...
```
The repository includes unit tests for tokenization utilities and Whisper helpers. Because external API calls are wrapped behind interfaces, tests run without contacting OpenAI.

## Workflow Walkthrough
A typical response generation flow looks like this:

1. A client sends a POST to `/v1/responses` with instructions plus conversation state.
2. `APIServer.handleCreateResponse` validates the payload, normalizes errors, and delegates to the wrapper.
3. `OpenAIWrapper.CreateResponse` builds a chat completion request, calls OpenAI, and maps the result into the service schema (ID, usage stats, message blocks).
4. The handler returns a JSON response compatible with the OpenAI Assistants beta, allowing UI clients to reuse existing adapters.

Thread, assistant, and vector store endpoints follow a similar pattern but fabricate resources locally. This lets downstream systems coordinate metadata and maintain stable IDs without persisting state externally.

## Extending the System
- Add new endpoints by implementing a handler in `server.go`, wiring it to the router in `main.go`, and delegating heavy lifting to either the OpenAI wrapper or a new utility package.
- Update the OpenAPI document (`docs/openapi.json`) so client SDKs stay in sync.
- Keep shared logic (token counting, storage adapters, etc.) inside `pkg/` to avoid circular dependencies.

## Additional Resources
- [OpenAI Go SDK](https://github.com/sashabaranov/go-openai) – Upstream client library used by the wrapper.
- [`policies.md`](policies.md) – Reference for data handling guidelines when operating the service.

With this structure in place you can experiment quickly while maintaining a clear separation between transport concerns, OpenAI integrations, and supporting utilities.
