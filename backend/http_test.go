package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProductHandler_MethodNotAllowed(t *testing.T) {
	service := NewProductService(&fakeSource{}, 30*time.Second)
	handler := NewProductHandler(service)

	request := httptest.NewRequest(http.MethodPost, "/products", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", recorder.Code)
	}
	if allow := recorder.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("expected Allow=GET, got %q", allow)
	}
}

func TestProductHandler_InvalidQuery(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 0}},
	}
	handler := NewProductHandler(NewProductService(source, 30*time.Second))

	request := httptest.NewRequest(http.MethodGet, "/products?minPrice=abc", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
}

func TestProductHandler_Success(t *testing.T) {
	source := &fakeSource{
		metadata: []MetadataRecord{{ID: "p1", Name: "Phone", BasePrice: 100}},
		details:  []DetailsRecord{{ID: "p1", DiscountPercent: 20, Bestseller: true, Colors: []string{"blue"}, Stock: 5}},
	}
	handler := NewProductHandler(NewProductService(source, 30*time.Second))

	request := httptest.NewRequest(http.MethodGet, "/products?limit=1&offset=0", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var response ProductListResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Total != 1 || len(response.Items) != 1 {
		t.Fatalf("expected one product in response, got total=%d len=%d", response.Total, len(response.Items))
	}
	if response.Items[0].Price != 80 {
		t.Fatalf("expected discounted price 80, got %v", response.Items[0].Price)
	}
}

func TestCORSMiddleware_Options(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/products", NewProductHandler(NewProductService(&fakeSource{}, 30*time.Second)))
	handler := withCORS(mux)

	request := httptest.NewRequest(http.MethodOptions, "/products", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", recorder.Code)
	}
	if recorder.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin=*")
	}
}
