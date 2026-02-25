# Backend

## Prerequisites
- Go 1.22 or higher

## Running the Server

```bash
cd backend
go run .
```

The server will start on `http://localhost:8080`

## Testing

```bash
go test ./...
```

Optional race check:

```bash
go test ./... -race
```

Note: `-race` requires CGO + a C compiler (for example `gcc`).

Linting (from repository root):

```bash
make lint
```

`make lint` runs `golangci-lint` via `go run` with a pinned version. First run may download tooling dependencies.

## Endpoints

### `GET /health`
Simple health check endpoint (helper for local/dev checks; not part of assignment scoring).

### `GET /products`
Returns merged product data from `data/metadata.json` + `data/details.json`, with server-side filtering and pagination.

#### Query parameters
- `search` (string): case-insensitive name search.
- `category` (string): category filter; supports repeated params and comma-separated values.
- `color` (string): color filter; supports repeated params and comma-separated values (e.g. `color=blue&color=red` or `color=blue,red`).
- `condition` (string): condition filter; supports repeated params and comma-separated values.
- `bestseller` (bool): strict `true` or `false`.
- `inStock` (bool): strict `true` or `false`; maps to effective stock `> 0` (effective stock is color-scoped when `color` filter is present).
- `onSale` (bool): strict `true` or `false`; maps to `discount_percent > 0`.
- `minStock` (int): inclusive minimum effective stock quantity.
- `minPrice` (number): inclusive minimum discounted price.
- `maxPrice` (number): inclusive maximum discounted price.
- `sort` (string): optional sort mode. Supported value: `popularity`.
- `limit` (int): page size. Default `6`, max `100`.
- `offset` (int): pagination offset. Default `0`.
- Any unsupported query parameter returns `400` (strict allowlist).

#### Example
```bash
curl "http://localhost:8080/products?search=iphone&category=smartphones&color=blue&condition=refurbished&bestseller=true&onSale=true&inStock=true&minStock=1&minPrice=100&maxPrice=800&sort=popularity&limit=6&offset=0"
```

#### Response shape
```json
{
  "items": [
    {
      "id": "p1",
      "name": "iPhone 12",
      "price": 311.24,
      "discount_percent": 25,
      "bestseller": true,
      "colors": ["blue", "red", "green"],
      "image_urls_by_color": {
        "blue": "https://files.refurbed.com/ii/iphone-12-1607327732.jpg?h=600&t=fitdesign&w=800",
        "red": "https://files.refurbed.com/pi/iphone-12-1627375405.jpg?t=resize&h=600&w=600",
        "green": "https://files.refurbed.com/pi/iphone-13-mini-1647245009.jpg?t=resize&h=600&w=600"
      },
      "stock_by_color": {
        "blue": 12,
        "red": 0,
        "green": 22
      },
      "image_url": "",
      "stock": 34,
      "category": "smartphones",
      "brand": "apple",
      "condition": "refurbished",
      "popularity_rank": 2
    }
  ],
  "total": 1,
  "limit": 6,
  "offset": 0,
  "has_more": false,
  "available_colors": ["blue", "green", "red", "silver"],
  "price_min": 99.99,
  "price_max": 1424.99
}
```

## Behavior and Design Notes
- The full aggregated product list is cached in memory for `30s` TTL.
- Filters/pagination are applied per request on top of cached data.
- The response includes `available_colors` derived from the aggregated dataset (unique, normalized, sorted) and limited to in-stock colors.
- The response includes `price_min` and `price_max` derived from the aggregated dataset (discounted prices), used by the frontend price slider bounds.
- Per-product `image_urls_by_color` is supported for strict color-to-image mapping.
- Per-product `stock_by_color` is supported; `stock` is computed as the sum of color stocks when `stock_by_color` exists.
- When `color` filters are used, stock-based filters (`inStock`, `minStock`) are evaluated against color-scoped stock.
- Optional popularity sorting is supported via `sort=popularity` using `data/popularity.json`.
- Cache refreshes are guarded to avoid stampedes (only one refresh runs after expiry).
- Cached product slices are returned via defensive copy semantics to avoid accidental mutation.
- If a cache refresh fails and stale cache exists, stale data is served and refresh is retried shortly after.
- Invalid query params return `400` with a descriptive JSON error.
- Requested `offset` is echoed as-is in the response, even when it is greater than `total`.
- Discounted prices are computed using cent-based arithmetic internally to avoid floating-point drift.
- Records that cannot be merged by `id` are skipped (only products present in both sources are returned).

## Assignment Requirement Coverage
- Two internal data sources: implemented via `data/metadata.json` and `data/details.json` read by `FileProductSource`.
- Aggregator endpoint: `GET /products` returns merged products with computed `price`.
- Search and filters: supports assignment filters plus extended filters (`category`, `condition`, `onSale`, `inStock`, `minStock`).
- Bonus sorting: supports popularity ranking via external source (`data/popularity.json`) and `sort=popularity`.
- Load more pagination: implemented server-side with `limit` and `offset`.
- In-memory cache: full aggregated product list cached for `30s` TTL.

## Explicit Tradeoffs
- Cache refresh is intentionally detached from request cancellation (`context.WithoutCancel`) once refresh begins, to prevent repeated canceled requests from starving cache refresh.
- On refresh failure, stale cached data is returned (if available) instead of surfacing `500`; this favors availability over immediate freshness/error visibility.
- Cached product slices are cloned before returning to callers; this avoids shared-mutation bugs at the cost of extra per-request allocations.
- Missing/`null` scalar fields in source JSON currently fall back to Go zero values (for example `discount_percent -> 0`) to keep ingestion resilient for assignment scope; production should enforce stricter schema validation plus data-quality monitoring/alerts.
- Type mismatches in source JSON (for example string instead of number) fail decode and surface as backend load failures (stale cache is served when available).
- Popularity source failures are non-fatal; products are still served without popularity ranks/sorting influence.

## Data Files
- `data/metadata.json` - Product metadata (`id`, `name`, `base_price`, `image_url`, `category`, `brand`)
- `data/details.json` - Product details (`id`, `discount_percent`, `bestseller`, `colors`, `image_urls_by_color`, `stock`, `stock_by_color`, `condition`)
- `data/popularity.json` - Optional popularity ranking source (`id`, `rank`)
