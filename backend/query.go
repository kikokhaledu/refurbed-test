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

type ProductQuery struct {
	Search     string
	Colors     []string
	Bestseller *bool
	MinPrice   *float64
	MaxPrice   *float64
	Limit      int
	Offset     int
}

func ParseProductQuery(values url.Values) (ProductQuery, error) {
	query := ProductQuery{
		Search: strings.TrimSpace(values.Get("search")),
		Colors: parseColors(values),
		Limit:  defaultLimit,
		Offset: 0,
	}

	if bestsellerRaw := strings.TrimSpace(values.Get("bestseller")); bestsellerRaw != "" {
		parsed, err := parseBoolStrict(bestsellerRaw)
		if err != nil {
			return ProductQuery{}, fmt.Errorf("invalid bestseller: %w", err)
		}
		query.Bestseller = &parsed
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

	if query.MinPrice != nil && query.MaxPrice != nil && *query.MinPrice > *query.MaxPrice {
		return ProductQuery{}, fmt.Errorf("minPrice cannot be greater than maxPrice")
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

func parseColors(values url.Values) []string {
	rawColors := values["color"]
	if len(rawColors) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(rawColors))
	colors := make([]string, 0, len(rawColors))

	for _, raw := range rawColors {
		parts := strings.Split(raw, ",")
		for _, part := range parts {
			color := normalizeToken(part)
			if color == "" {
				continue
			}
			if _, exists := seen[color]; exists {
				continue
			}
			seen[color] = struct{}{}
			colors = append(colors, color)
		}
	}

	return colors
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

func normalizeToken(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}
