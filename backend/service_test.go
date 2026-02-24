package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeSource struct {
	metadata []MetadataRecord
	details  []DetailsRecord
	err      error

	metadataCalls int
	detailsCalls  int
}

func (f *fakeSource) LoadMetadata(_ context.Context) ([]MetadataRecord, error) {
	f.metadataCalls++
	if f.err != nil {
		return nil, f.err
	}
	return append([]MetadataRecord(nil), f.metadata...), nil
}

func (f *fakeSource) LoadDetails(_ context.Context) ([]DetailsRecord, error) {
	f.detailsCalls++
	if f.err != nil {
		return nil, f.err
	}
	return append([]DetailsRecord(nil), f.details...), nil
}

func TestMergeProducts_BasicAndPriceCalculation(t *testing.T) {
	metadata := []MetadataRecord{
		{ID: "p1", Name: "Phone", BasePrice: 1000, ImageURL: "img"},
		{ID: "p2", Name: "Watch", BasePrice: 200, ImageURL: "img2"},
		{ID: "p3", Name: "NoDetails", BasePrice: 100},
	}
	details := []DetailsRecord{
		{ID: "p1", DiscountPercent: 25, Bestseller: true, Colors: []string{"Blue", "blue", "RED"}, Stock: 10},
		{ID: "p2", DiscountPercent: 0, Bestseller: false, Colors: []string{"black"}, Stock: -4},
	}

	got, err := mergeProducts(metadata, details)
	if err != nil {
		t.Fatalf("mergeProducts() unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 merged products, got %d", len(got))
	}

	if got[0].Price != 750 {
		t.Fatalf("expected discounted price 750, got %v", got[0].Price)
	}
	if strings.Join(got[0].Colors, ",") != "blue,red" {
		t.Fatalf("expected normalized colors [blue red], got %v", got[0].Colors)
	}
	if got[1].Stock != 0 {
		t.Fatalf("expected stock to be clamped to 0, got %d", got[1].Stock)
	}
}

func TestMergeProducts_DuplicateIDs(t *testing.T) {
	_, err := mergeProducts(
		[]MetadataRecord{{ID: "p1"}, {ID: "p1"}},
		[]DetailsRecord{{ID: "p1"}},
	)
	if err == nil || !strings.Contains(err.Error(), "duplicate id") {
		t.Fatalf("expected duplicate metadata id error, got %v", err)
	}

	_, err = mergeProducts(
		[]MetadataRecord{{ID: "p1"}},
		[]DetailsRecord{{ID: "p1"}, {ID: "p1"}},
	)
	if err == nil || !strings.Contains(err.Error(), "duplicate id") {
		t.Fatalf("expected duplicate details id error, got %v", err)
	}
}

func TestProductService_QueryProducts_FilterAndPagination(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{
			{ID: "p1", Name: "iPhone 12", BasePrice: 400},
			{ID: "p2", Name: "Galaxy S23", BasePrice: 500},
			{ID: "p3", Name: "iPhone 13", BasePrice: 700},
		},
		details: []DetailsRecord{
			{ID: "p1", DiscountPercent: 10, Bestseller: true, Colors: []string{"blue"}, Stock: 1},
			{ID: "p2", DiscountPercent: 0, Bestseller: false, Colors: []string{"green"}, Stock: 1},
			{ID: "p3", DiscountPercent: 20, Bestseller: true, Colors: []string{"red"}, Stock: 1},
		},
	}
	service := NewProductService(source, 30*time.Second)

	minPrice := 500.0
	limit := 1
	response, err := service.QueryProducts(context.Background(), ProductQuery{
		Search:     "iphone",
		Bestseller: boolPtr(true),
		MinPrice:   &minPrice,
		Limit:      limit,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}

	if response.Total != 1 {
		t.Fatalf("expected total=1, got %d", response.Total)
	}
	if len(response.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(response.Items))
	}
	if response.Items[0].ID != "p3" {
		t.Fatalf("expected p3, got %s", response.Items[0].ID)
	}
	if response.HasMore {
		t.Fatalf("expected has_more=false")
	}
}

func TestProductService_CachesWithinTTL(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 0}},
	}

	now := time.Date(2026, 2, 24, 19, 0, 0, 0, time.UTC)
	service := NewProductService(source, 30*time.Second)
	service.now = func() time.Time { return now }

	if _, err := service.QueryProducts(context.Background(), ProductQuery{}); err != nil {
		t.Fatalf("first query unexpected error: %v", err)
	}
	if _, err := service.QueryProducts(context.Background(), ProductQuery{}); err != nil {
		t.Fatalf("second query unexpected error: %v", err)
	}

	if source.metadataCalls != 1 || source.detailsCalls != 1 {
		t.Fatalf("expected exactly one source load in TTL window, metadataCalls=%d detailsCalls=%d", source.metadataCalls, source.detailsCalls)
	}
}

func TestProductService_RefreshesAfterTTL(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 0}},
	}

	now := time.Date(2026, 2, 24, 19, 0, 0, 0, time.UTC)
	service := NewProductService(source, 30*time.Second)
	service.now = func() time.Time { return now }

	if _, err := service.QueryProducts(context.Background(), ProductQuery{}); err != nil {
		t.Fatalf("first query unexpected error: %v", err)
	}
	now = now.Add(31 * time.Second)
	if _, err := service.QueryProducts(context.Background(), ProductQuery{}); err != nil {
		t.Fatalf("second query unexpected error: %v", err)
	}

	if source.metadataCalls != 2 || source.detailsCalls != 2 {
		t.Fatalf("expected cache refresh after TTL, metadataCalls=%d detailsCalls=%d", source.metadataCalls, source.detailsCalls)
	}
}

func TestProductService_PropagatesSourceError(t *testing.T) {
	source := &fakeSource{err: errors.New("boom")}
	service := NewProductService(source, 30*time.Second)

	_, err := service.QueryProducts(context.Background(), ProductQuery{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "load metadata") {
		t.Fatalf("expected wrapped load metadata error, got %q", err.Error())
	}
}

func TestProductService_EmptyResultsReturnsEmptySlice(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 0, Colors: []string{"blue"}}},
	}
	service := NewProductService(source, 30*time.Second)

	response, err := service.QueryProducts(context.Background(), ProductQuery{
		Colors: []string{"red"},
	})
	if err != nil {
		t.Fatalf("QueryProducts() unexpected error: %v", err)
	}
	if response.Items == nil {
		t.Fatalf("expected empty items slice, got nil")
	}
	if len(response.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(response.Items))
	}
}

func boolPtr(v bool) *bool {
	return &v
}
