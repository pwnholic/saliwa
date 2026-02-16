// Package domain provides core domain types for the exchange connector.
package domain

import (
	"fmt"
	"strings"
)

// SymbolInfo contains metadata about a trading symbol.
type SymbolInfo struct {
	// Exchange is the name of the exchange
	Exchange string `json:"exchange"`

	// Symbol is the normalized symbol (e.g., "BTC/USDT")
	Symbol string `json:"symbol"`

	// BaseAsset is the base asset (e.g., "BTC")
	BaseAsset string `json:"base_asset"`

	// QuoteAsset is the quote asset (e.g., "USDT")
	QuoteAsset string `json:"quote_asset"`

	// ExchangeSymbol is the exchange-specific symbol format (e.g., "BTCUSDT")
	ExchangeSymbol string `json:"exchange_symbol"`

	// Status indicates if the symbol is trading
	Status string `json:"status"`

	// BaseAssetPrecision is the precision for base asset
	BaseAssetPrecision int `json:"base_asset_precision"`

	// QuoteAssetPrecision is the precision for quote asset
	QuoteAssetPrecision int `json:"quote_asset_precision"`

	// MinQuantity is the minimum order quantity
	MinQuantity Decimal `json:"min_quantity,omitempty"`

	// MaxQuantity is the maximum order quantity
	MaxQuantity Decimal `json:"max_quantity,omitempty"`

	// QuantityStep is the quantity increment
	QuantityStep Decimal `json:"quantity_step,omitempty"`

	// MinPrice is the minimum price
	MinPrice Decimal `json:"min_price,omitempty"`

	// MaxPrice is the maximum price
	MaxPrice Decimal `json:"max_price,omitempty"`

	// PriceStep is the price increment
	PriceStep Decimal `json:"price_step,omitempty"`

	// MinNotional is the minimum order value (price * quantity)
	MinNotional Decimal `json:"min_notional,omitempty"`
}

// NormalizeSymbol converts an exchange-specific symbol to normalized format.
// Exchange formats:
//   - Binance: "BTCUSDT" -> "BTC/USDT"
//   - Bybit: "BTCUSDT" -> "BTC/USDT"
//
// The function attempts to find common quote currencies to split the symbol.
func NormalizeSymbol(exchangeSymbol string) string {
	// If already normalized (contains "/"), return as-is
	if strings.Contains(exchangeSymbol, "/") {
		return strings.ToUpper(exchangeSymbol)
	}

	symbol := strings.ToUpper(exchangeSymbol)

	// Common quote currencies in order of length (longest first)
	quoteCurrencies := []string{
		"USDC", "USDT", "USDS", "BUSD", "TUSD", "USDK", "USD",
		"EUR", "GBP", "JPY", "AUD", "CAD", "CHF",
		"BTC", "ETH", "BNB", "SOL", "XRP",
		"TRY", "BRL", "RUB", "ZAR", "UAH", "NGN",
		"VAI", "DAI", "IDRT", "BKRW", "BVND",
	}

	for _, quote := range quoteCurrencies {
		if before, ok := strings.CutSuffix(symbol, quote); ok {
			base := before
			if base != "" {
				return base + "/" + quote
			}
		}
	}

	// If no common quote currency found, return original
	return symbol
}

// ExchangeSymbol converts a normalized symbol to exchange-specific format.
// For example: "BTC/USDT" -> "BTCUSDT"
func ExchangeSymbol(normalizedSymbol string) string {
	return strings.ToUpper(strings.ReplaceAll(normalizedSymbol, "/", ""))
}

// ParseSymbol parses a symbol into base and quote assets.
// Accepts both normalized ("BTC/USDT") and exchange ("BTCUSDT") formats.
func ParseSymbol(symbol string) (base, quote string, err error) {
	// Try normalized format first
	if strings.Contains(symbol, "/") {
		parts := strings.Split(symbol, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid symbol format: %s", symbol)
		}
		return strings.ToUpper(parts[0]), strings.ToUpper(parts[1]), nil
	}

	// Try to normalize and parse
	normalized := NormalizeSymbol(symbol)
	if strings.Contains(normalized, "/") {
		parts := strings.Split(normalized, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid symbol format: %s", symbol)
		}
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("cannot parse symbol: %s", symbol)
}

// IsSymbolValid checks if a symbol string is valid.
func IsSymbolValid(symbol string) bool {
	if symbol == "" {
		return false
	}

	// Check for invalid characters
	for _, c := range symbol {
		if !isValidSymbolChar(c) {
			return false
		}
	}

	return true
}

// isValidSymbolChar returns true if the character is valid in a symbol.
func isValidSymbolChar(c rune) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '/' ||
		c == '-' ||
		c == '_'
}

// FormatSymbol formats a base and quote asset into normalized symbol format.
func FormatSymbol(base, quote string) string {
	return strings.ToUpper(base) + "/" + strings.ToUpper(quote)
}

// SymbolsEqual checks if two symbols are equal (case-insensitive, format-insensitive).
func SymbolsEqual(a, b string) bool {
	// Normalize both symbols
	normalA := NormalizeSymbol(a)
	normalB := NormalizeSymbol(b)
	return strings.EqualFold(normalA, normalB)
}
