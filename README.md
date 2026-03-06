# Beans News API & MCP (README)

**Version:** 0.1

Beans is an intelligent news & blogs aggregation and search service that curates fresh content from RSS feeds using AI-powered natural language queries and vector search. This README describes every public route, its parameters, expected outputs, and provides `curl` examples for quick experimentation.

---

## General Notes

- All endpoints return JSON and, unless otherwise noted, respond with HTTP status codes indicating success (`200`, `204`, etc.) or the appropriate error (`400`, `401`, `500`, ...).
- Authentication is optional; if the `API_KEYS` configuration map is non‑empty the server requires a matching header for each request. Example header in curl: `-H "X-Api-Key: <your-key>"`.
- Concurrency is limited via a channel; requests beyond the limit are automatically queued by Go's scheduler.
- Responses follow the OpenAPI definitions in `router/routes.go` (see the types near the top of the file).

---

## Common curl flags

```bash
BASE_URL="http://localhost:8080"               # adjust to your server address
API_KEY="my-secret"                            # only if API key enforcement enabled
COMMON="-H 'Content-Type: application/json'"   # send/receive JSON
AUTH="-H 'X-API-KEY: $API_KEY'"               # include conditionally
```

---

## Routes

### 1. Health Check

```bash
curl -s $COMMON $AUTH "$BASE_URL/health" | jq
```

- **Method:** `GET`
- **Path:** `/health`
- **Description:** Simple liveness probe used by load‑balancers or orchestrators.
- **Request body:** none
- **Response body:**
  ```json
  {
    "status": "alive"
  }
  ```

---

### 2. Favicon

```bash
curl -i $COMMON $AUTH "$BASE_URL/favicon.ico"
```

- **Method:** `GET`
- **Path:** `/favicon.ico`
- **Description:** Redirects (302) to a static image URL defined by `FAVICON_PATH`.
- **Response:** HTTP redirect; no JSON.

---

### 3. Categories

```bash
curl -s $COMMON $AUTH "$BASE_URL/tags/categories?offset=0&limit=20" | jq
```

- **Method:** `GET`
- **Path:** `/tags/categories`
- **Query parameters:**
  - `offset` (int ≥ 0, default 0)
  - `limit` (int 1–100, default 16)
- **Response:** array of strings, e.g. `["AI", "Cybersecurity", "Politics"]`.

---

### 4. Entities

```bash
curl -s $COMMON $AUTH "$BASE_URL/tags/entities?limit=50" | jq
```

- **Method:** `GET`
- **Path:** `/tags/entities`
- **Query parameters:** same as categories.
- **Response:** array of named-entity strings discovered in articles.

---

### 5. Regions

```bash
curl -s $COMMON $AUTH "$BASE_URL/tags/regions?offset=10" | jq
```

- **Method:** `GET`
- **Path:** `/tags/regions`
- **Query parameters:** same as categories.
- **Response:** array of geographic regions.

---

### 6. Publishers

```bash
curl -s $COMMON $AUTH "$BASE_URL/publishers?sources=example.com&offset=0&limit=5" | jq
```

- **Method:** `GET`
- **Path:** `/publishers`
- **Query parameters:**
  - `sources` (required, comma-separated list)
  - `offset`, `limit` same as above
- **Response:** array of publisher objects:
  ```json
  {
    "source": "example.com",
    "base_url": "https://example.com",
    "site_name": "Example News",
    "description": "A sample publisher",
    "favicon": "https://example.com/favicon.ico"
  }
  ```

---

### 7. Publisher Sources

```bash
curl -s $COMMON $AUTH "$BASE_URL/publishers/sources?limit=10" | jq
```

- **Method:** `GET`
- **Path:** `/publishers/sources`
- **Query parameters:** `offset` and `limit` as before.
- **Response:** list of unique `source` IDs (strings) present in the database.

---

### 8. Latest Articles

This is the primary search route. It supports both full-text **vector search** when `q` is provided and traditional tag‑based direct filtering (`categories`, `entities`, `regions`, etc.). The rich filters documented below can be combined with either mode.

```bash
curl -s $COMMON $AUTH \
  "$BASE_URL/articles/latest?q=artificial+intelligence&acc=0.8&kind=news&categories=AI,Tech&offset=0&limit=5" \
  | jq
```

- **Method:** `GET`
- **Path:** `/articles/latest`
- **Query parameters:**
  - `q` (string, 3–512 chars): semantic search query; embedding is computed server-side.
  - `acc` (float 0–1, default 0.75): cosine similarity threshold (1 = exact).
  - `kind` (`news` or `blog`).
  - `tags`/`entities`/`regions`/`sources` (comma-separated lists).
  - `published_since` (RFC3339 datetime).
  - `trending_since` (RFC3339 datetime).
  - `with_content` (boolean): only return items with `content` populated.
  - `offset`, `limit` for pagination.
- **Response:** JSON array of `Bean` objects. Example element:
  ```json
  {
    "url": "https://example.com/story",
    "kind": "news",
    "source": "example.com",
    "title": "Breaking AI News",
    "summary": "A short summary...",
    "content": null,
    "image_url": null,
    "author": "Jane Doe",
    "created": "2024-02-28T12:34:56Z",
    "entities": ["OpenAI", "GPT"],
    "regions": ["US"],
    "categories": ["AI"],
    "sentiments": ["positive"]
  }
  ```

---

### 9. Trending Articles

Same parameters as **Latest Articles** but results are ordered by internal `trend_score` instead of date. Use when you want socially‑relevant or high‑engagement pieces.

```bash
curl -s $COMMON $AUTH \
  "$BASE_URL/articles/trending?q=golang&acc=0.7&limit=10" | jq
```

- **Method:** `GET`
- **Path:** `/articles/trending`
- **Query parameters:** identical to `/articles/latest`
- **Response:** array of `BeanAggregate` objects; each includes the `trend_score` field along with the usual article data.

---

## Development & Running

1. **Build:** `go build` (the project uses modules defined in `go.mod`).
2. **Configuration:**
   - `PORT` – port to listen on (default 8080).
   - `API_KEYS` – optional comma-separated `Header:Value` pairs used by `verifyAPIKey` middleware.
   - `PG_CONNECTION_STRING` – PostgreSQL DSN used by `beansack/pgsack.go`; **required for production**.
   - `MAX_CONCURRENT_REQS` – limits concurrent requests processed by the in‑memory queue (defaults to `1` if unset).
   - `EMBEDDER_BASE_URL` – URL of the embedding service used by `nlp.Embedder`.
   - `EMBEDDER_API_KEY` – API key sent to the embedder service (if required).
3. **Run:** `./gobeansack` or via `docker-compose up` (Dockerfile and compose manifest included).  Environment variables can be placed in a `.env` file and loaded by the `docker-compose` setup.

```bash
# example running locally
export PORT=8080
export API_KEYS="X-API-KEY:foo"
export PG_CONNECTION_STRING="postgres://user:pass@localhost:5432/beans"
export MAX_CONCURRENT_REQUESTS=512
export EMBEDDER_BASE_URL="https://my-embedder.local"  # e.g. OpenAI or local service
export EMBEDDER_API_KEY="secret123"
./go-beans-api
```

### Deployment Notes

- Ensure PostgreSQL is reachable and the connection string includes SSL settings if needed.
- The embedder service must implement the same API as expected by `nlp/embedder.go`; you can swap in any hosted or self‑hosted vector generator.
- Tuning `MAX_CONCURRENT_REQS` allows the binary to be run on higher‑capacity machines or behind load‑balancers without dropping requests.
- API key enforcement is disabled when `API_KEYS` is empty; set a value before exposing the service publicly.

---

## License

This repository is licensed under the **MIT License**. See [`LICENSE`](LICENSE) for details.

---

*Documentation generated from `router/routes.go` on 2026-03-06.*
