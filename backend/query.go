package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	defaultLimit = 6
	maxLimit     = 100
)

var allowedQueryParams = map[string]struct{}{
	"search":     {},
	"color":      {},
	"category":   {},
	"condition":  {},
	"bestseller": {},
	"inStock":    {},
	"onSale":     {},
	"minStock":   {},
	"minPrice":   {},
	"maxPrice":   {},
	"sort":       {},
	"limit":      {},
	"offset":     {},
}

type ProductQuery struct {
	Search     string
	Colors     []string
	Categories []string
	Conditions []string
	Sort       string
	Bestseller *bool
	InStock    *bool
	OnSale     *bool
	MinPrice   *float64
	MaxPrice   *float64
	MinStock   *int
	Limit      int
	Offset     int
}

func ParseProductQuery(values url.Values) (ProductQuery, error) {
	if err := validateAllowedQueryParams(values); err != nil {
		return ProductQuery{}, err
	}

	query := ProductQuery{
		Search:     strings.TrimSpace(values.Get("search")),
		Colors:     parseTokenList(values, "color"),
		Categories: parseTokenList(values, "category"),
		Conditions: parseTokenList(values, "condition"),
		Sort:       "",
		Limit:      defaultLimit,
		Offset:     0,
	}

	if bestsellerRaw := strings.TrimSpace(values.Get("bestseller")); bestsellerRaw != "" {
		parsed, err := parseBoolStrict(bestsellerRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid bestseller: %w", err)
		}
		query.Bestseller = &parsed
	}

	if inStockRaw := strings.TrimSpace(values.Get("inStock")); inStockRaw != "" {
		parsed, err := parseBoolStrict(inStockRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid inStock: %w", err)
		}
		query.InStock = &parsed
	}
	if onSaleRaw := strings.TrimSpace(values.Get("onSale")); onSaleRaw != "" {
		parsed, err := parseBoolStrict(onSaleRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid onSale: %w", err)
		}
		query.OnSale = &parsed
	}

	if minPriceRaw := strings.TrimSpace(values.Get("minPrice")); minPriceRaw != "" {
		parsed, err := parseNonNegativeFloat(minPriceRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid minPrice: %w", err)
		}
		query.MinPrice = &parsed
	}

	if maxPriceRaw := strings.TrimSpace(values.Get("maxPrice")); maxPriceRaw != "" {
		parsed, err := parseNonNegativeFloat(maxPriceRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid maxPrice: %w", err)
		}
		query.MaxPrice = &parsed
	}

	if minStockRaw := strings.TrimSpace(values.Get("minStock")); minStockRaw != "" {
		parsed, err := parseNonNegativeInt(minStockRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid minStock: %w", err)
		}
		query.MinStock = &parsed
	}

	if query.MinPrice != nil && query.MaxPrice != nil && *query.MinPrice > *query.MaxPrice {
		return ProductQuery{}, fmt.Errorf("minPrice cannot be greater than maxPrice")
	}

	if sortRaw := strings.TrimSpace(values.Get("sort")); sortRaw != "" {
		sortValue := strings.ToLower(sortRaw)
		switch sortValue {
		case "popularity":
			query.Sort = sortValue
		default:
			return ProductQuery{}, fmt.Errorf("invalid sort: must be 'popularity'")
		}
	}

	if limitRaw := strings.TrimSpace(values.Get("limit")); limitRaw != "" {
		parsed, err := strconv.Atoi(limitRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid limit: must be an integer")
		}
		if parsed <= 0 {
			return ProductQuery{}, fmt.Errorf("invalid limit: must be greater than 0")
		}
		if parsed > maxLimit {
			return ProductQuery{}, fmt.Errorf("invalid limit: must be <= %d", maxLimit)
		}
		query.Limit = parsed
	}

	if offsetRaw := strings.TrimSpace(values.Get("offset")); offsetRaw != "" {
		parsed, err := strconv.Atoi(offsetRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid offset: must be an integer")
		}
		if parsed < 0 {
			return ProductQuery{}, fmt.Errorf("invalid offset: must be >= 0")
		}
		query.Offset = parsed
	}

	return query, nil
}

func parseTokenList(values url.Values, key string) []string {
	rawValues := values[key]
	if len(rawValues) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(rawValues))
	items := make([]string, 0, len(rawValues))

	for _, raw := range rawValues {
		parts := strings.Split(raw, ",")
		for _, part := range parts {
			item := normalizeToken(part)
			if item == "" {
				continue
			}
			if _, exists := seen[item]; exists {
				continue
			}
			seen[item] = struct{}{}
			items = append(items, item)
		}
	}

	return items
}

func validateAllowedQueryParams(values url.Values) error {
	for key := range values {
		if _, ok := allowedQueryParams[key]; !ok {
			return fmt.Errorf("unsupported query parameter %q", key)
		}
	}
	return nil
}

func parseBoolStrict(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("must be 'true' or 'false'")
	}
}

func parseNonNegativeFloat(raw string) (float64, error) {
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("must be a number")
	}
	if value < 0 {
		return 0, fmt.Errorf("must be >= 0")
	}
	return value, nil
}

func parseNonNegativeInt(raw string) (int, error) {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("must be an integer")
	}
	if value < 0 {
		return 0, fmt.Errorf("must be >= 0")
	}
	return value, nil
}

func normalizeToken(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}
