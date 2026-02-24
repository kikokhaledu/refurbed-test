package main

import (
	"net/url"
	"strings"
	"testing"
)

func TestParseProductQuery_Defaults(t *testing.T) {
	values := url.Values{}

	query, err := ParseProductQuery(values)
	if err != nil {
		t.Fatalf("ParseProductQuery() unexpected error: %v", err)
	}

	if query.Limit != defaultLimit {
		t.Fatalf("expected default limit %d, got %d", defaultLimit, query.Limit)
	}
	if query.Offset != 0 {
		t.Fatalf("expected default offset 0, got %d", query.Offset)
	}
	if query.Bestseller != nil {
		t.Fatalf("expected bestseller to be nil")
	}
	if len(query.Colors) != 0 {
		t.Fatalf("expected no colors, got %v", query.Colors)
	}
}

func TestParseProductQuery_ParsesFilters(t *testing.T) {
	values := url.Values{
		"search":     []string{"  iPhone "},
		"color":      []string{"Blue, red", "green", "blue"},
		"bestseller": []string{"true"},
		"minPrice":   []string{"100"},
		"maxPrice":   []string{"700"},
		"limit":      []string{"8"},
		"offset":     []string{"16"},
	}

	query, err := ParseProductQuery(values)
	if err != nil {
		t.Fatalf("ParseProductQuery() unexpected error: %v", err)
	}

	if query.Search != "iPhone" {
		t.Fatalf("expected search=iPhone, got %q", query.Search)
	}
	if query.Bestseller == nil || !*query.Bestseller {
		t.Fatalf("expected bestseller=true, got %v", query.Bestseller)
	}
	if query.MinPrice == nil || *query.MinPrice != 100 {
		t.Fatalf("expected minPrice=100, got %v", query.MinPrice)
	}
	if query.MaxPrice == nil || *query.MaxPrice != 700 {
		t.Fatalf("expected maxPrice=700, got %v", query.MaxPrice)
	}
	if query.Limit != 8 {
		t.Fatalf("expected limit=8, got %d", query.Limit)
	}
	if query.Offset != 16 {
		t.Fatalf("expected offset=16, got %d", query.Offset)
	}
	wantColors := []string{"blue", "red", "green"}
	if strings.Join(query.Colors, ",") != strings.Join(wantColors, ",") {
		t.Fatalf("expected colors=%v, got %v", wantColors, query.Colors)
	}
}

func TestParseProductQuery_InvalidCases(t *testing.T) {
	tests := []struct {
		name    string
		values  url.Values
		wantErr string
	}{
		{
			name:    "invalid bestseller",
			values:  url.Values{"bestseller": []string{"yes"}},
			wantErr: "invalid bestseller",
		},
		{
			name:    "invalid minPrice",
			values:  url.Values{"minPrice": []string{"abc"}},
			wantErr: "invalid minPrice",
		},
		{
			name:    "invalid maxPrice",
			values:  url.Values{"maxPrice": []string{"-1"}},
			wantErr: "invalid maxPrice",
		},
		{
			name:    "min greater than max",
			values:  url.Values{"minPrice": []string{"500"}, "maxPrice": []string{"100"}},
			wantErr: "minPrice cannot be greater than maxPrice",
		},
		{
			name:    "invalid limit",
			values:  url.Values{"limit": []string{"0"}},
			wantErr: "invalid limit",
		},
		{
			name:    "invalid offset",
			values:  url.Values{"offset": []string{"-1"}},
			wantErr: "invalid offset",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseProductQuery(tc.values)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}
