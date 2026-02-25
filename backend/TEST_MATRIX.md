# Backend Test Coverage Matrix

Status legend:
- `Covered`: directly verified by one or more tests.
- `Partially covered`: key path is tested, but some branches or operational concerns are not.
- `Not covered`: no direct automated test in current suite.

| Area | Status | Notes |
|---|---|---|
| Query parsing defaults, validation, boundaries | Covered | `query_test.go` covers defaults, strict validation, invalid inputs, boundaries, and degenerate token input. |
| Filtering behavior (`search`, `category`, `condition`, `color`, `bestseller`, `onSale`, `inStock`, `minStock`, price bounds) | Covered | `service_test.go` has table-style scenarios for filter combinations, inclusive bounds, and color-scoped stock semantics. |
| Pagination semantics (`limit`, `offset`, load-more shape) | Covered | `service_test.go` validates paging behavior including `offset > total` response semantics. |
| Aggregation/merge correctness from two sources | Covered | `service_test.go` validates merge output, duplicate/empty IDs, price calculation, stock/image normalization behavior. |
| Price computation precision | Covered | `TestDiscountedPriceCents_RoundsAtCentPrecision`. |
| Cache TTL, refresh, stale fallback, anti-stampede, wait cancellation | Covered | `service_test.go` includes TTL hit/miss, stale-on-error, single refresh fan-in, and cancellation while waiting. |
| Popularity sort (`sort=popularity`) and non-fatal popularity source failure | Covered | `service_test.go` and `query_test.go`. |
| Repository file loading (missing file, malformed JSON, context cancel, null/missing scalar behavior) | Covered | `repository_test.go`. |
| HTTP handler method validation, bad query, success path, internal error JSON, CORS OPTIONS | Covered | `http_test.go`. |
| CORS behavior for non-OPTIONS requests | Partially covered | OPTIONS preflight is tested; non-OPTIONS header assertions are not explicitly isolated. |
| Popularity data normalization edge cases (duplicate IDs, empty IDs, invalid rank values) | Partially covered | Happy path and source failure are covered; invalid ranking payload branches are not directly unit-tested. |
| Defensive-copy immutability guarantees of cached products | Partially covered | Behavior is indirectly exercised; no focused test mutating returned slices/maps to assert isolation. |
| `/health` endpoint behavior | Not covered | Endpoint exists in `main.go` but has no dedicated test. |
| Server bootstrap and graceful shutdown wiring (`main.go`, signal handling, timeout config) | Not covered | No integration test around process lifecycle. |
| Logging middleware output format/content | Not covered | Middleware exists, but log assertions are not included. |
| Load/performance/soak behavior | Not covered | No benchmark or load-test suite in repository. |
| Race detector execution in CI-like environment | Partially covered | Code is race-conscious and `go test -race` is documented, but compiler/runtime availability can block it in some environments. |

## Additional Edge-Case Gaps (Current)

| Edge Case | Status | Notes |
|---|---|---|
| Repeated conflicting boolean params (`bestseller=true&bestseller=false`) | Not covered | Parser currently uses `Get`, so first-value behavior should be asserted explicitly. |
| Case-insensitive sort parsing (`sort=POPULARITY`) | Partially covered | Implementation lowercases sort input; no direct test for mixed/upper-case values. |
| Unknown color filter against `stock_by_color` map | Partially covered | Color-scoped stock logic is covered, but unknown-color behavior is not isolated as a dedicated case. |
| Dataset with `stock_by_color` present but `colors` missing/empty | Partially covered | Normalization behavior exists; no focused test for this shape. |
| Dataset with `image_urls_by_color` present but `colors` missing/empty | Partially covered | Merge logic supports adding colors from image map keys; no dedicated edge test. |
| `available_colors` when `stock_by_color` absent and aggregate `stock <= 0` | Not covered | Out-of-stock color exclusion is tested for `stock_by_color`; aggregate-stock-only path is not isolated. |
| `limit` clamping from service sanitization (direct service call with oversized limit) | Not covered | Query parser rejects `>100`, but direct `ProductQuery` service path clamp is not tested. |
| `offset < 0` sanitization in service (direct service call) | Not covered | Parser rejects negative offsets; service-level sanitization path is not directly tested. |
| Popularity rank tie-break determinism (same rank -> name -> id) | Not covered | Popularity sorting is tested, but explicit tie-break rules are not isolated in a dedicated assertion. |
| Invalid popularity data fallback path (`duplicate id`, `rank <= 0`, empty id) | Not covered | Branch exists (logs and continues), but no direct test through query/service path. |
| Cache defensive-copy immutability (`Colors`, `StockByColor`, `ImageURLsByColor` mutation safety) | Not covered | Clone behavior exists; no test mutating returned payload and re-querying cache. |
| `/products` response `Content-Type` contract on both success and errors | Partially covered | Body/status behavior is asserted; explicit header assertions are limited. |
| CORS headers on non-OPTIONS product requests | Not covered | OPTIONS preflight covered; GET/400/500 CORS header presence not explicitly asserted. |
| `/health` endpoint wrapped with middleware stack (`withCORS`, `withLogging`) | Not covered | Health handler logic exists in `main.go`; integration-style route stack test missing. |
| Shutdown path in `main.go` when context is canceled | Not covered | Graceful shutdown wiring is present but untested. |
