package main

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

type ProductService struct {
	source ProductSource
	ttl    time.Duration
	now    func() time.Time

	mu        sync.RWMutex
	cached    []Product
	expiresAt time.Time
}

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

func (s *ProductService) QueryProducts(ctx context.Context, query ProductQuery) (ProductListResponse, error) {
	query = sanitizeQuery(query)

	products, err := s.getAggregatedProducts(ctx)
	if err != nil {
		return ProductListResponse{}, err
	}

	filtered := filterProducts(products, query)
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
		Items:   page,
		Total:   total,
		Limit:   query.Limit,
		Offset:  query.Offset,
		HasMore: end < total,
	}, nil
}

func (s *ProductService) getAggregatedProducts(ctx context.Context) ([]Product, error) {
	now := s.now()

	s.mu.RLock()
	if now.Before(s.expiresAt) && s.cached != nil {
		cachedCopy := cloneProducts(s.cached)
		s.mu.RUnlock()
		return cachedCopy, nil
	}
	s.mu.RUnlock()

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

	s.mu.Lock()
	defer s.mu.Unlock()

	now = s.now()
	if now.Before(s.expiresAt) && s.cached != nil {
		return cloneProducts(s.cached), nil
	}

	s.cached = cloneProducts(merged)
	s.expiresAt = now.Add(s.ttl)

	return cloneProducts(s.cached), nil
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

		products = append(products, Product{
			ID:              id,
			Name:            strings.TrimSpace(meta.Name),
			Price:           discountedPrice(meta.BasePrice, detail.DiscountPercent),
			DiscountPercent: clampPercent(detail.DiscountPercent),
			Bestseller:      detail.Bestseller,
			Colors:          normalizeColors(detail.Colors),
			ImageURL:        strings.TrimSpace(meta.ImageURL),
			Stock:           max(0, detail.Stock),
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

		filtered = append(filtered, product)
	}

	return filtered
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
	if basePrice < 0 {
		basePrice = 0
	}
	discountPercent = clampPercent(discountPercent)
	value := basePrice * (1 - float64(discountPercent)/100)
	return math.Round(value*100) / 100
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

func cloneProducts(products []Product) []Product {
	if len(products) == 0 {
		return nil
	}
	clone := make([]Product, len(products))
	for i, product := range products {
		colors := make([]string, len(product.Colors))
		copy(colors, product.Colors)
		product.Colors = colors
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
