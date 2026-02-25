# Refurbed Engineering - Senior Fullstack Assignment

This repository contains the backend plus both frontend tracks for the assignment.

## Documentation Index
- Assignment brief: [assignment.md](./assignment.md)
- Backend guide: [backend/README.md](./backend/README.md)
- Vue track overview: [assignment_vue/README.md](./assignment_vue/README.md)
- Vue frontend guide: [assignment_vue/frontend-vue/README.md](./assignment_vue/frontend-vue/README.md)
- Vanilla track overview: [assignment_vanilla/README.md](./assignment_vanilla/README.md)
- Vanilla frontend guide: [assignment_vanilla/frontend-vanilla/README.md](./assignment_vanilla/frontend-vanilla/README.md)

## Repository Layout
- `backend/` - Go API (`GET /products`, caching, filtering, load-more pagination).
- `assignment_vue/frontend-vue/` - Vue + Tailwind frontend implementation.
- `assignment_vanilla/frontend-vanilla/` - Vanilla JS starter track (optional alternative).

## Quick Start (Backend Only)
```bash
cd backend
go run .
```

Backend runs at `http://localhost:8080`.

## Quick Start (Vue Frontend Only)
```bash
cd assignment_vue/frontend-vue
npm install
npm run dev
```

Frontend runs at `http://localhost:5173`.

## Docker Compose (Full Stack)
From repository root:

```bash
docker compose up --build -d
```

Stop and remove containers:

```bash
docker compose down
```

Notes:
- Compose file: [docker-compose.yml](./docker-compose.yml)
- Backend image definition: [backend/Dockerfile](./backend/Dockerfile)
- Frontend image definition: [assignment_vue/frontend-vue/Dockerfile](./assignment_vue/frontend-vue/Dockerfile)
- `backend/data` is mounted read-only into the container at `/app/data`.
- Frontend is exposed on port `5173` and calls backend via `VITE_API_BASE_URL=http://localhost:8080`.

## Makefile Commands
Common developer commands are available from repository root:

```bash
make help
make up
make seed
make down
make reset
make lint
make pre-push
```

`make up` now starts both backend and Vue frontend.

`make pre-push` runs:
- `make fmt`
- `make vet`
- `make lint`
- `make test`
- `make compose-check`

## What I Would Improve/Change For Production
- Replace in-memory cache with Redis for shared cache across instances, stronger invalidation options, and better horizontal scaling.
- Add explicit stale-while-revalidate semantics with cache metadata/headers and stronger refresh observability.
- Move from file-based sources to resilient upstream APIs or persisted storage with schema validation and contract checks.
- Introduce a search service (Elasticsearch/OpenSearch/Solr) only when catalog size and product discovery complexity justify it (fuzzy matching, relevance ranking, faceting, typo tolerance).
- Add authentication/rate limiting where required, plus stricter CORS policy per environment.
- Add end-to-end monitoring stack: structured logs, metrics dashboards, alerting, and distributed tracing.
- Extend CI quality gates with race tests in Linux CI runner, coverage threshold enforcement, and artifact/image scanning.

## Notes On Architecture, Decisions Or Other Comments
- The backend is organized into separable layers for parsing (`query.go`), data loading (`repository.go`), business logic/cache (`service.go`), and HTTP transport (`http.go`).
- Product data from two sources is merged once per cache refresh cycle and then filtered/paginated in memory per request.
- Response contract uses a paginated envelope (`items`, `total`, `limit`, `offset`, `has_more`) and also exposes `available_colors` for data-driven color filters.
- Product payload supports `image_urls_by_color` for strict per-color image rendering in the frontend.
- Product payload supports `stock_by_color` for color-level stock representation.
- In addition to assignment-required filters, the API supports richer filtering for `category`, `condition`, `onSale`, `inStock`, and `minStock` (backend-level support).
- Vue frontend uses debounced search auto-apply (500ms), explicit apply for non-search filters, sortable results (`sort=popularity`), and mobile-friendly floating actions for filters + scroll-to-top.
- Filter panel actions are position-aware: top controls are used when visible, and a sticky bottom action bar appears after scrolling past them (including unapplied-changes warning).
- Active filter chips support per-filter removal and clear-all behavior.
- Price range slider bounds are sourced from backend dataset-level price bounds (`price_min`/`price_max`) instead of static UI constants.
- Mobile filter UX now uses an accessible modal flow (focus trap, `Esc` close, body scroll lock).
- Product cards use explicit color-to-image mapping (`image_urls_by_color`) with fallback placeholder handling for broken image URLs.
- Query parsing is strict (allowlist + explicit validation) to fail fast on malformed inputs.
- Price calculation uses cent-based arithmetic internally to reduce floating-point drift.
- Documented backend tradeoffs: refresh work is detached from request cancellation and stale cache can be served on refresh failure to prioritize availability.
- Source JSON ingestion currently tolerates missing/`null` scalar fields via zero-value defaults (assignment pragmatism), while type mismatches still fail decoding; for production this should be backed by stricter schema validation and data-quality alerting.
- Dev ergonomics are supported with Docker Compose + Makefile commands for consistent local setup.

## Final Thoughts
- Current implementation intentionally prioritizes clarity, correctness, and testability over premature complexity.
- For this assignment scope, in-memory filtering and caching are sufficient and avoid overengineering.
- This README section is a living summary and can be extended as frontend implementation and final integration notes are completed.
