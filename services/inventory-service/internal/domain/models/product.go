package models

type Product struct {
	ID           string
	Name         string
	Description  string
	PriceMinor   int64
	Currency     string
	CategoryID   string
	CategoryName string
	ImageURL     string
	IsActive     bool
}
