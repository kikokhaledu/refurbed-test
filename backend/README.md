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

## Endpoints

### `GET /health`
Simple health check endpoint.

### `GET /products`
Returns merged product data from `data/metadata.json` + `data/details.json`, with server-side filtering and pagination.

#### Query parameters
- `search` (string): case-insensitive name search.
- `color` (string): color filter; supports repeated params and comma-separated values (e.g. `color=blue&color=red` or `color=blue,red`).
- `bestseller` (bool): strict `true` or `false`.
- `minPrice` (number): inclusive minimum discounted price.
- `maxPrice` (number): inclusive maximum discounted price.
- `limit` (int): page size. Default `6`, max `100`.
- `offset` (int): pagination offset. Default `0`.

#### Example
```bash
curl "http://localhost:8080/products?search=iphone&color=blue&bestseller=true&minPrice=100&maxPrice=800&limit=6&offset=0"
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
      "image_url": "",
      "stock": 34
    }
  ],
  "total": 1,
  "limit": 6,
  "offset": 0,
  "has_more": false
}
```

## Behavior and Design Notes
- The full aggregated product list is cached in memory for `30s` TTL.
- Filters/pagination are applied per request on top of cached data.
- Cache is guarded by `sync.RWMutex` and uses defensive copy semantics to avoid accidental mutation.
- Invalid query params return `400` with a descriptive JSON error.

## Data Files
- `data/metadata.json` - Product metadata (id, name, base_price, image_url)
- `data/details.json` - Product details (id, discount_percent, bestseller, colors, stock)
