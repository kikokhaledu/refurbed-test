# Frontend - Vue 3 + Tailwind CSS

This is the primary implemented frontend track for this repository.

## Prerequisites
- Node.js 18+
- Backend API running on `http://localhost:8080`

## Local Development
From this directory:

```bash
npm install
npm run dev
```

Frontend is available at `http://localhost:5173`.

## Environment Variables
- `VITE_API_BASE_URL` (optional): backend base URL.
  - Default fallback is `http://localhost:8080`.

Example:

```bash
VITE_API_BASE_URL=http://localhost:8080 npm run dev
```

## Docker
This frontend has a Dockerfile at [Dockerfile](./Dockerfile) and is wired into root Compose.

From repository root:

```bash
make up
```

Then open:
- Frontend: `http://localhost:5173`
- Backend: `http://localhost:8080`

## Build
```bash
npm run build
```

## Tests
```bash
npm run test:run
```

Current unit coverage includes:
- `FiltersPanel` dual-range clamp behavior and sticky bottom action controls visibility.
- `ProductCard` out-of-stock color hiding, color-based image switching, and image fallback behavior.
- `SortSelect` sort value emission.
- `ActiveFilterChips` chip removal and clear-all interactions.

## Tech Stack
- Vue 3 (`<script setup>`, Composition API)
- Vite
- Tailwind CSS
- Vitest + Vue Test Utils (unit tests)

## Implemented UI Features
- Debounced search auto-apply (500ms after typing stops).
- Draft-vs-applied filtering for non-search controls (`Apply` button inside filter panel).
- Filters panel actions are position-aware: action controls stay at the top when visible, and switch to a sticky bottom action bar (with unapplied-changes warning) when the top controls scroll out of view.
- URL sync for applied query state (`search`, `category`, `color`, `condition`, `bestseller`, `onSale`, `inStock`, `sort`, `minPrice`, `maxPrice`, `offset`).
- Filter panel with category, colors, condition, bestseller, on-sale toggle, stock availability (`all`/`in stock`), and a dual-thumb single-track price slider.
- Price slider bounds are data-driven from backend response (`price_min`/`price_max`), not hardcoded.
- Color filter options are data-driven from backend `available_colors`.
- Sort control supports backend popularity ranking (`sort=popularity`).
- Active filter chips allow single-filter removal and `Clear all` from the results area.
- Product card color swatches change the card image using backend `image_urls_by_color` mapping (fallback to `image_url` if a color mapping is missing).
- Product images include lazy loading, async decoding, explicit dimensions, and runtime fallback to a local placeholder image when a URL fails.
- Out-of-stock colors are hidden from card swatches and backend-provided color filter options.
- Stock label updates per selected color (from `stock_by_color`).
- Load-more pagination (`limit`/`offset`) with append behavior.
- Loading skeletons, error state with retry, and empty state with clear filters.
- Dark mode toggle with persisted preference (`localStorage`).
- Mobile UX: filter modal with focus trap + `Esc` close + body scroll lock, plus floating actions (`Filters` and `Top` buttons).
