//go:build tools

package mocks

//go:generate go run go.uber.org/mock/mockgen@latest -source=../../internal/clients/inventory_client.go -destination=mock_inventory_client.go -package=mocks
