package main

type MetadataRecord struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	BasePrice float64 `json:"base_price"`
	ImageURL  string  `json:"image_url"`
}

type DetailsRecord struct {
	ID              string   `json:"id"`
	DiscountPercent int      `json:"discount_percent"`
	Bestseller      bool     `json:"bestseller"`
	Colors          []string `json:"colors"`
	Stock           int      `json:"stock"`
}

type Product struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Price           float64  `json:"price"`
	DiscountPercent int      `json:"discount_percent"`
	Bestseller      bool     `json:"bestseller"`
	Colors          []string `json:"colors"`
	ImageURL        string   `json:"image_url"`
	Stock           int      `json:"stock"`
}

type ProductListResponse struct {
	Items   []Product `json:"items"`
	Total   int       `json:"total"`
	Limit   int       `json:"limit"`
	Offset  int       `json:"offset"`
	HasMore bool      `json:"has_more"`
}

type errorResponse struct {
	Error string `json:"error"`
}
