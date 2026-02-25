package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"slices"
	"strings"
	"sync"
	"time"
)

type ProductService struct {
	source           ProductSource
	popularitySource PopularitySource
	ttl              time.Duration
	now              func() time.Time

	mu        sync.Mutex
	cached    []Product
	expiresAt time.Time
	loading   bool
	loadDone  chan struct{}
}

const staleRetryWindow = 2 * time.Second

func NewProductService(source ProductSource, ttl time.Duration) *ProductService {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	return &ProductService{
		source: source,
		ttl:    ttl,
		now:    time.Now,
	}
}

func (s *ProductService) WithPopularitySource(source PopularitySource) *ProductService {
	s.popularitySource = source
	return s
}

func (s *ProductService) QueryProducts(ctx context.Context, query ProductQuery) (ProductListResponse, error) {
	query = sanitizeQuery(query)

	products, err := s.getAggregatedProducts(ctx)
	if err != nil {
		return ProductListResponse{}, err
	}
	availableColors := listAvailableColors(products)
	priceMin, priceMax := listAvailablePriceBounds(products)

	filtered := filterProducts(products, query)
	sortProducts(filtered, query.Sort)
	total := len(filtered)

	start := query.Offset
	if start > total {
		start = total
	}

	end := start + query.Limit
	if end > total {
		end = total
	}

	page := filtered[start:end]
	if page == nil {
		page = []Product{}
	}

	return ProductListResponse{
		Items:           page,
		Total:           total,
		Limit:           query.Limit,
		Offset:          query.Offset,
		HasMore:         end < total,
		AvailableColors: availableColors,
		PriceMin:        priceMin,
		PriceMax:        priceMax,
	}, nil
}

func (s *ProductService) getAggregatedProducts(ctx context.Context) ([]Product, error) {
	for {
		now := s.now()

		s.mu.Lock()
		if now.Before(s.expiresAt) && s.cached != nil {
			cachedCopy := cloneProducts(s.cached)
			s.mu.Unlock()
			return cachedCopy, nil
		}

		if s.loading {
			loadDone := s.loadDone
			s.mu.Unlock()

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-loadDone:
				continue
			}
		}

		s.loading = true
		s.loadDone = make(chan struct{})
		loadDone := s.loadDone
		s.mu.Unlock()

		merged, err := s.loadAndMerge(context.WithoutCancel(ctx))

		var staleFallback []Product
		s.mu.Lock()
		if err == nil {
			s.cached = cloneProducts(merged)
			s.expiresAt = s.now().Add(s.ttl)
		} else if s.cached != nil {
			staleFallback = cloneProducts(s.cached)
			retryAfter := staleRetryWindow
			if s.ttl > 0 && s.ttl < retryAfter {
				retryAfter = s.ttl
			}
			if retryAfter <= 0 {
				retryAfter = time.Second
			}
			s.expiresAt = s.now().Add(retryAfter)
		}
		s.loading = false
		close(loadDone)
		s.loadDone = nil
		if err == nil {
			cachedCopy := cloneProducts(s.cached)
			s.mu.Unlock()
			return cachedCopy, nil
		}
		s.mu.Unlock()

		if staleFallback != nil {
			log.Printf("products refresh failed, serving stale cache: %v", err)
			return staleFallback, nil
		}

		return nil, err
	}
}

func (s *ProductService) loadAndMerge(ctx context.Context) ([]Product, error) {
	metadata, err := s.source.LoadMetadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	details, err := s.source.LoadDetails(ctx)
	if err != nil {
		return nil, fmt.Errorf("load details: %w", err)
	}

	merged, err := mergeProducts(metadata, details)
	if err != nil {
		return nil, fmt.Errorf("merge products: %w", err)
	}

	applyPopularityRanks(merged, nil)
	if s.popularitySource != nil {
		popularity, popErr := s.popularitySource.LoadPopularity(ctx)
		if popErr != nil {
			log.Printf("popularity source load failed, continuing without popularity sort data: %v", popErr)
			return merged, nil
		}
		rankings, rankErr := normalizePopularityRankings(popularity)
		if rankErr != nil {
			log.Printf("popularity source data invalid, continuing without popularity sort data: %v", rankErr)
			return merged, nil
		}
		applyPopularityRanks(merged, rankings)
	}

	return merged, nil
}

func sanitizeQuery(query ProductQuery) ProductQuery {
	if query.Limit <= 0 {
		query.Limit = defaultLimit
	}
	if query.Limit > maxLimit {
		query.Limit = maxLimit
	}
	if query.Offset < 0 {
		query.Offset = 0
	}
	query.Search = strings.TrimSpace(query.Search)
	return query
}

func mergeProducts(metadata []MetadataRecord, details []DetailsRecord) ([]Product, error) {
	detailsByID := make(map[string]DetailsRecord, len(details))
	for _, record := range details {
		id := strings.TrimSpace(record.ID)
		if id == "" {
			return nil, fmt.Errorf("details contains empty id")
		}
		if _, exists := detailsByID[id]; exists {
			return nil, fmt.Errorf("details contains duplicate id %q", id)
		}
		detailsByID[id] = record
	}

	seenMetadataIDs := make(map[string]struct{}, len(metadata))
	products := make([]Product, 0, len(metadata))

	for _, meta := range metadata {
		id := strings.TrimSpace(meta.ID)
		if id == "" {
			return nil, fmt.Errorf("metadata contains empty id")
		}
		if _, exists := seenMetadataIDs[id]; exists {
			return nil, fmt.Errorf("metadata contains duplicate id %q", id)
		}
		seenMetadataIDs[id] = struct{}{}

		detail, ok := detailsByID[id]
		if !ok {
			continue
		}

		baseColors := normalizeColors(detail.Colors)
		stockByColor := normalizeStockByColor(detail.StockByColor, baseColors)
		imageURLsByColor := normalizeImageURLsByColor(detail.ImageURLsByColor)
		colors := mergeColorListWithStockKeys(baseColors, stockByColor)
		colors = mergeColorListWithImageKeys(colors, imageURLsByColor)

		stock := max(0, detail.Stock)
		if len(stockByColor) > 0 {
			stock = sumStockByColor(stockByColor)
		}

		products = append(products, Product{
			ID:               id,
			Name:             strings.TrimSpace(meta.Name),
			Price:            discountedPrice(meta.BasePrice, detail.DiscountPercent),
			DiscountPercent:  clampPercent(detail.DiscountPercent),
			Bestseller:       detail.Bestseller,
			Colors:           colors,
			ImageURLsByColor: imageURLsByColor,
			StockByColor:     stockByColor,
			ImageURL:         strings.TrimSpace(meta.ImageURL),
			Stock:            stock,
			Category:         normalizeToken(meta.Category),
			Brand:            normalizeToken(meta.Brand),
			Condition:        normalizeToken(detail.Condition),
		})
	}

	return products, nil
}

func filterProducts(products []Product, query ProductQuery) []Product {
	if len(products) == 0 {
		return nil
	}

	search := strings.ToLower(strings.TrimSpace(query.Search))
	colorFilter := make(map[string]struct{}, len(query.Colors))
	for _, color := range query.Colors {
		normalized := normalizeToken(color)
		if normalized == "" {
			continue
		}
		colorFilter[normalized] = struct{}{}
	}
	categoryFilter := make(map[string]struct{}, len(query.Categories))
	for _, category := range query.Categories {
		normalized := normalizeToken(category)
		if normalized == "" {
			continue
		}
		categoryFilter[normalized] = struct{}{}
	}
	conditionFilter := make(map[string]struct{}, len(query.Conditions))
	for _, condition := range query.Conditions {
		normalized := normalizeToken(condition)
		if normalized == "" {
			continue
		}
		conditionFilter[normalized] = struct{}{}
	}

	filtered := make([]Product, 0, len(products))

	for _, product := range products {
		if search != "" && !strings.Contains(strings.ToLower(product.Name), search) {
			continue
		}
		if query.Bestseller != nil && product.Bestseller != *query.Bestseller {
			continue
		}
		if query.MinPrice != nil && product.Price < *query.MinPrice {
			continue
		}
		if query.MaxPrice != nil && product.Price > *query.MaxPrice {
			continue
		}
		if len(colorFilter) > 0 && !matchesAnyColor(product.Colors, colorFilter) {
			continue
		}
		if len(categoryFilter) > 0 {
			if _, ok := categoryFilter[normalizeToken(product.Category)]; !ok {
				continue
			}
		}
		if len(conditionFilter) > 0 {
			if _, ok := conditionFilter[normalizeToken(product.Condition)]; !ok {
				continue
			}
		}
		effectiveStock := effectiveStockForQuery(product, colorFilter)
		if query.InStock != nil && (effectiveStock > 0) != *query.InStock {
			continue
		}
		if query.OnSale != nil && (product.DiscountPercent > 0) != *query.OnSale {
			continue
		}
		if query.MinStock != nil && effectiveStock < *query.MinStock {
			continue
		}

		filtered = append(filtered, product)
	}

	return filtered
}

func sortProducts(products []Product, sortMode string) {
	if len(products) <= 1 {
		return
	}

	switch strings.ToLower(strings.TrimSpace(sortMode)) {
	case "popularity":
		slices.SortStableFunc(products, func(a, b Product) int {
			aRank := popularityRankOrFallback(a.PopularityRank)
			bRank := popularityRankOrFallback(b.PopularityRank)
			if aRank < bRank {
				return -1
			}
			if aRank > bRank {
				return 1
			}

			aName := strings.ToLower(strings.TrimSpace(a.Name))
			bName := strings.ToLower(strings.TrimSpace(b.Name))
			if aName < bName {
				return -1
			}
			if aName > bName {
				return 1
			}
			if a.ID < b.ID {
				return -1
			}
			if a.ID > b.ID {
				return 1
			}
			return 0
		})
	}
}

func popularityRankOrFallback(rank int) int {
	if rank > 0 {
		return rank
	}
	return maxInt
}

func matchesAnyColor(productColors []string, filter map[string]struct{}) bool {
	for _, productColor := range productColors {
		if _, ok := filter[normalizeToken(productColor)]; ok {
			return true
		}
	}
	return false
}

func discountedPrice(basePrice float64, discountPercent int) float64 {
	return float64(discountedPriceCents(basePrice, discountPercent)) / 100
}

func discountedPriceCents(basePrice float64, discountPercent int) int64 {
	if basePrice < 0 {
		basePrice = 0
	}
	baseCents := int64(math.Round(basePrice * 100))

	discountPercent = clampPercent(discountPercent)
	factor := int64(100 - discountPercent)
	return (baseCents*factor + 50) / 100
}

func clampPercent(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func normalizeColors(colors []string) []string {
	seen := make(map[string]struct{}, len(colors))
	out := make([]string, 0, len(colors))
	for _, raw := range colors {
		color := normalizeToken(raw)
		if color == "" {
			continue
		}
		if _, ok := seen[color]; ok {
			continue
		}
		seen[color] = struct{}{}
		out = append(out, color)
	}
	return out
}

func normalizeStockByColor(stockByColor map[string]int, colors []string) map[string]int {
	if len(stockByColor) == 0 {
		return nil
	}

	normalized := make(map[string]int, len(stockByColor)+len(colors))
	for _, rawColor := range colors {
		color := normalizeToken(rawColor)
		if color == "" {
			continue
		}
		normalized[color] = 0
	}

	for rawColor, rawStock := range stockByColor {
		color := normalizeToken(rawColor)
		if color == "" {
			continue
		}
		normalized[color] = max(0, rawStock)
	}

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

func normalizeImageURLsByColor(imageURLsByColor map[string]string) map[string]string {
	if len(imageURLsByColor) == 0 {
		return nil
	}

	normalized := make(map[string]string, len(imageURLsByColor))
	for rawColor, rawURL := range imageURLsByColor {
		color := normalizeToken(rawColor)
		imageURL := strings.TrimSpace(rawURL)
		if color == "" || imageURL == "" {
			continue
		}
		normalized[color] = imageURL
	}

	if len(normalized) == 0 {
		return nil
	}

	return normalized
}

func mergeColorListWithStockKeys(colors []string, stockByColor map[string]int) []string {
	if len(stockByColor) == 0 {
		return colors
	}

	seen := make(map[string]struct{}, len(colors))
	out := make([]string, 0, len(colors)+len(stockByColor))
	for _, color := range colors {
		normalized := normalizeToken(color)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	missing := make([]string, 0, len(stockByColor))
	for color := range stockByColor {
		normalized := normalizeToken(color)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		missing = append(missing, normalized)
	}
	slices.Sort(missing)
	out = append(out, missing...)

	return out
}

func mergeColorListWithImageKeys(colors []string, imageURLsByColor map[string]string) []string {
	if len(imageURLsByColor) == 0 {
		return colors
	}

	seen := make(map[string]struct{}, len(colors))
	out := make([]string, 0, len(colors)+len(imageURLsByColor))
	for _, color := range colors {
		normalized := normalizeToken(color)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		out = append(out, normalized)
	}

	missing := make([]string, 0, len(imageURLsByColor))
	for color := range imageURLsByColor {
		normalized := normalizeToken(color)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		missing = append(missing, normalized)
	}
	slices.Sort(missing)
	out = append(out, missing...)

	return out
}

func sumStockByColor(stockByColor map[string]int) int {
	if len(stockByColor) == 0 {
		return 0
	}

	total := 0
	for _, stock := range stockByColor {
		total += max(0, stock)
	}
	return total
}

func effectiveStockForQuery(product Product, colorFilter map[string]struct{}) int {
	if len(product.StockByColor) == 0 {
		return max(0, product.Stock)
	}

	if len(colorFilter) == 0 {
		return sumStockByColor(product.StockByColor)
	}

	total := 0
	matched := false
	for color, stock := range product.StockByColor {
		normalized := normalizeToken(color)
		if _, ok := colorFilter[normalized]; !ok {
			continue
		}
		matched = true
		total += max(0, stock)
	}

	if matched {
		return total
	}

	return 0
}

func listAvailableColors(products []Product) []string {
	if len(products) == 0 {
		return []string{}
	}

	seen := make(map[string]struct{})
	out := make([]string, 0)

	for _, product := range products {
		for _, color := range product.Colors {
			normalized := normalizeToken(color)
			if normalized == "" {
				continue
			}
			if !isColorInStockForProduct(product, normalized) {
				continue
			}
			if _, ok := seen[normalized]; ok {
				continue
			}
			seen[normalized] = struct{}{}
			out = append(out, normalized)
		}
	}

	if len(out) == 0 {
		return []string{}
	}

	slices.Sort(out)
	return out
}

func listAvailablePriceBounds(products []Product) (minPrice float64, maxPrice float64) {
	if len(products) == 0 {
		return 0, 0
	}

	minPrice = products[0].Price
	maxPrice = products[0].Price
	for _, product := range products[1:] {
		if product.Price < minPrice {
			minPrice = product.Price
		}
		if product.Price > maxPrice {
			maxPrice = product.Price
		}
	}

	return minPrice, maxPrice
}

func normalizePopularityRankings(records []PopularityRecord) (map[string]int, error) {
	if len(records) == 0 {
		return nil, nil
	}

	rankings := make(map[string]int, len(records))
	for _, record := range records {
		id := strings.TrimSpace(record.ID)
		if id == "" {
			return nil, fmt.Errorf("popularity contains empty id")
		}
		if record.Rank <= 0 {
			return nil, fmt.Errorf("popularity rank for %q must be > 0", id)
		}
		if _, exists := rankings[id]; exists {
			return nil, fmt.Errorf("popularity contains duplicate id %q", id)
		}
		rankings[id] = record.Rank
	}

	return rankings, nil
}

func applyPopularityRanks(products []Product, rankings map[string]int) {
	for i := range products {
		products[i].PopularityRank = 0
		if len(rankings) == 0 {
			continue
		}
		if rank, ok := rankings[products[i].ID]; ok {
			products[i].PopularityRank = rank
		}
	}
}

func isColorInStockForProduct(product Product, color string) bool {
	if len(product.StockByColor) == 0 {
		return max(0, product.Stock) > 0
	}

	stock, ok := product.StockByColor[normalizeToken(color)]
	if !ok {
		return false
	}

	return max(0, stock) > 0
}

func cloneProducts(products []Product) []Product {
	if len(products) == 0 {
		return nil
	}
	clone := make([]Product, len(products))
	for i, product := range products {
		colors := make([]string, len(product.Colors))
		copy(colors, product.Colors)
		product.Colors = colors
		if len(product.StockByColor) > 0 {
			stockByColor := make(map[string]int, len(product.StockByColor))
			for color, stock := range product.StockByColor {
				stockByColor[color] = stock
			}
			product.StockByColor = stockByColor
		}
		if len(product.ImageURLsByColor) > 0 {
			imageURLsByColor := make(map[string]string, len(product.ImageURLsByColor))
			for color, imageURL := range product.ImageURLsByColor {
				imageURLsByColor[color] = imageURL
			}
			product.ImageURLsByColor = imageURLsByColor
		}
		clone[i] = product
	}
	return clone
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

const maxInt = int(^uint(0) >> 1)
