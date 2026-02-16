// Package binance implements the Binance exchange driver.
// WebSocket message types for Binance streams.
// API Documentation: https://binance-docs.github.io/apidocs/spot/en/#websocket-market-streams
// API Version: v3 (verified 2026-02-16)
package binance

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lilwiggy/ex-act/pkg/domain"
)

// WSMessage is the base wrapper for combined stream messages.
// Binance combined streams format: {"stream":"btcusdt@ticker","data":{...}}
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#combined-stream-exports
type WSMessage struct {
	Stream string          `json:"stream"`
	Data   json.RawMessage `json:"data"`
}

// WSTicker represents a ticker update from the ticker stream.
// WebSocket Stream: <symbol>@ticker or <symbol>@ticker@1s
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#individual-symbol-ticker-streams
// All price/quantity fields are strings in Binance JSON - MUST parse to Decimal.
type WSTicker struct {
	EventType string `json:"e"` // Event type (24hrTicker)
	EventTime int64  `json:"E"` // Event time (milliseconds)
	Symbol    string `json:"s"` // Symbol (e.g., "BTCUSDT")

	PriceChange        string `json:"p"` // Price change
	PriceChangePercent string `json:"P"` // Price change percent
	WeightedAvgPrice   string `json:"w"` // Weighted average price
	PrevClosePrice     string `json:"x"` // Previous close price
	LastPrice          string `json:"c"` // Last price
	LastQuantity       string `json:"Q"` // Last quantity
	BidPrice           string `json:"b"` // Best bid price
	BidQuantity        string `json:"B"` // Best bid quantity
	AskPrice           string `json:"a"` // Best ask price
	AskQuantity        string `json:"A"` // Best ask quantity
	OpenPrice          string `json:"o"` // Open price
	HighPrice          string `json:"h"` // High price
	LowPrice           string `json:"l"` // Low price
	Volume             string `json:"v"` // Total traded base asset volume
	QuoteVolume        string `json:"q"` // Total traded quote asset volume
	OpenTime           int64  `json:"O"` // Statistics open time
	CloseTime          int64  `json:"C"` // Statistics close time
	FirstTradeID       int64  `json:"F"` // First trade ID
	LastTradeID        int64  `json:"L"` // Last trade ID
	TradeCount         int64  `json:"T"` // Total number of trades
}

// ToDomain converts WSTicker to domain.Ticker.
func (t *WSTicker) ToDomain(exchange string) (*domain.Ticker, error) {
	symbol := domain.NormalizeSymbol(t.Symbol)

	bidPrice, err := domain.NewDecimal(t.BidPrice)
	if err != nil {
		return nil, fmt.Errorf("parse bid_price: %w", err)
	}

	bidQty, err := domain.NewDecimal(t.BidQuantity)
	if err != nil {
		return nil, fmt.Errorf("parse bid_quantity: %w", err)
	}

	askPrice, err := domain.NewDecimal(t.AskPrice)
	if err != nil {
		return nil, fmt.Errorf("parse ask_price: %w", err)
	}

	askQty, err := domain.NewDecimal(t.AskQuantity)
	if err != nil {
		return nil, fmt.Errorf("parse ask_quantity: %w", err)
	}

	lastPrice, err := domain.NewDecimal(t.LastPrice)
	if err != nil {
		return nil, fmt.Errorf("parse last_price: %w", err)
	}

	highPrice, err := domain.NewDecimal(t.HighPrice)
	if err != nil {
		return nil, fmt.Errorf("parse high_price: %w", err)
	}

	lowPrice, err := domain.NewDecimal(t.LowPrice)
	if err != nil {
		return nil, fmt.Errorf("parse low_price: %w", err)
	}

	volume, err := domain.NewDecimal(t.Volume)
	if err != nil {
		return nil, fmt.Errorf("parse volume: %w", err)
	}

	quoteVolume, err := domain.NewDecimal(t.QuoteVolume)
	if err != nil {
		return nil, fmt.Errorf("parse quote_volume: %w", err)
	}

	priceChange, err := domain.NewDecimal(t.PriceChange)
	if err != nil {
		return nil, fmt.Errorf("parse price_change: %w", err)
	}

	priceChangePercent, err := domain.NewDecimal(t.PriceChangePercent)
	if err != nil {
		return nil, fmt.Errorf("parse price_change_percent: %w", err)
	}

	openPrice, err := domain.NewDecimal(t.OpenPrice)
	if err != nil {
		return nil, fmt.Errorf("parse open_price: %w", err)
	}

	return &domain.Ticker{
		Exchange:           exchange,
		Symbol:             symbol,
		BidPrice:           bidPrice,
		BidQuantity:        bidQty,
		AskPrice:           askPrice,
		AskQuantity:        askQty,
		LastPrice:          lastPrice,
		HighPrice:          highPrice,
		LowPrice:           lowPrice,
		Volume:             volume,
		QuoteVolume:        quoteVolume,
		PriceChange:        priceChange,
		PriceChangePercent: priceChangePercent,
		OpenPrice:          openPrice,
		Timestamp:          time.UnixMilli(t.EventTime),
	}, nil
}

// WSBookTicker represents best price update from the book ticker stream.
// WebSocket Stream: <symbol>@bookTicker
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#diff-depth-stream
type WSBookTicker struct {
	UpdateID int64  `json:"u"` // Update ID
	Symbol   string `json:"s"` // Symbol
	BidPrice string `json:"b"` // Best bid price
	BidQty   string `json:"B"` // Best bid quantity
	AskPrice string `json:"a"` // Best ask price
	AskQty   string `json:"A"` // Best ask quantity
}

// ToDomain converts WSBookTicker to domain.Ticker (limited fields).
func (t *WSBookTicker) ToDomain(exchange string) (*domain.Ticker, error) {
	symbol := domain.NormalizeSymbol(t.Symbol)

	bidPrice, err := domain.NewDecimal(t.BidPrice)
	if err != nil {
		return nil, fmt.Errorf("parse bid_price: %w", err)
	}

	bidQty, err := domain.NewDecimal(t.BidQty)
	if err != nil {
		return nil, fmt.Errorf("parse bid_quantity: %w", err)
	}

	askPrice, err := domain.NewDecimal(t.AskPrice)
	if err != nil {
		return nil, fmt.Errorf("parse ask_price: %w", err)
	}

	askQty, err := domain.NewDecimal(t.AskQty)
	if err != nil {
		return nil, fmt.Errorf("parse ask_quantity: %w", err)
	}

	return &domain.Ticker{
		Exchange:    exchange,
		Symbol:      symbol,
		BidPrice:    bidPrice,
		BidQuantity: bidQty,
		AskPrice:    askPrice,
		AskQuantity: askQty,
		Timestamp:   time.Now(),
	}, nil
}

// WSDepthUpdate represents an order book delta update.
// WebSocket Stream: <symbol>@depth or <symbol>@depth@100ms
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#diff-depth-stream
// IMPORTANT: This is DIFFERENT from WSDepthSnapshot - this is a delta, not full snapshot.
type WSDepthUpdate struct {
	EventType     string     `json:"e"` // Event type (depthUpdate)
	EventTime     int64      `json:"E"` // Event time
	Symbol        string     `json:"s"` // Symbol
	FirstUpdateID int64      `json:"U"` // First update ID in event
	FinalUpdateID int64      `json:"u"` // Final update ID in event
	Bids          [][]string `json:"b"` // Bids [[price, qty], ...]
	Asks          [][]string `json:"a"` // Asks [[price, qty], ...]
}

// ToDomain converts WSDepthUpdate to bid and ask slices.
func (d *WSDepthUpdate) ToDomain() (bids, asks []domain.OrderBookLevel, err error) {
	bids = make([]domain.OrderBookLevel, 0, len(d.Bids))
	for _, bid := range d.Bids {
		if len(bid) < 2 {
			continue
		}
		price, err := domain.NewDecimal(bid[0])
		if err != nil {
			return nil, nil, fmt.Errorf("parse bid price: %w", err)
		}
		qty, err := domain.NewDecimal(bid[1])
		if err != nil {
			return nil, nil, fmt.Errorf("parse bid quantity: %w", err)
		}
		bids = append(bids, domain.OrderBookLevel{Price: price, Quantity: qty})
	}

	asks = make([]domain.OrderBookLevel, 0, len(d.Asks))
	for _, ask := range d.Asks {
		if len(ask) < 2 {
			continue
		}
		price, err := domain.NewDecimal(ask[0])
		if err != nil {
			return nil, nil, fmt.Errorf("parse ask price: %w", err)
		}
		qty, err := domain.NewDecimal(ask[1])
		if err != nil {
			return nil, nil, fmt.Errorf("parse ask quantity: %w", err)
		}
		asks = append(asks, domain.OrderBookLevel{Price: price, Quantity: qty})
	}

	return bids, asks, nil
}

// WSDepthSnapshot represents a full order book snapshot.
// This is the initial snapshot received when subscribing to depth streams.
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#diff-depth-stream
type WSDepthSnapshot struct {
	LastUpdateID int64      `json:"lastUpdateId"` // Last update ID
	Bids         [][]string `json:"bids"`         // Bids [[price, qty], ...]
	Asks         [][]string `json:"asks"`         // Asks [[price, qty], ...]
}

// ToDomain converts WSDepthSnapshot to domain.OrderBook.
func (d *WSDepthSnapshot) ToDomain(exchange, symbol string) (*domain.OrderBook, error) {
	normalizedSymbol := domain.NormalizeSymbol(symbol)

	bids := make([]domain.OrderBookLevel, 0, len(d.Bids))
	for _, bid := range d.Bids {
		if len(bid) < 2 {
			continue
		}
		price, err := domain.NewDecimal(bid[0])
		if err != nil {
			return nil, fmt.Errorf("parse bid price: %w", err)
		}
		qty, err := domain.NewDecimal(bid[1])
		if err != nil {
			return nil, fmt.Errorf("parse bid quantity: %w", err)
		}
		bids = append(bids, domain.OrderBookLevel{Price: price, Quantity: qty})
	}

	asks := make([]domain.OrderBookLevel, 0, len(d.Asks))
	for _, ask := range d.Asks {
		if len(ask) < 2 {
			continue
		}
		price, err := domain.NewDecimal(ask[0])
		if err != nil {
			return nil, fmt.Errorf("parse ask price: %w", err)
		}
		qty, err := domain.NewDecimal(ask[1])
		if err != nil {
			return nil, fmt.Errorf("parse ask quantity: %w", err)
		}
		asks = append(asks, domain.OrderBookLevel{Price: price, Quantity: qty})
	}

	return &domain.OrderBook{
		Exchange:     exchange,
		Symbol:       normalizedSymbol,
		Bids:         bids,
		Asks:         asks,
		LastUpdateID: d.LastUpdateID,
		Timestamp:    time.Now(),
	}, nil
}

// WSTrade represents a trade message from the trade stream.
// WebSocket Stream: <symbol>@trade or <symbol>@aggTrade
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#trade-streams
type WSTrade struct {
	EventType       string `json:"e"` // Event type (trade or aggTrade)
	EventTime       int64  `json:"E"` // Event time
	Symbol          string `json:"s"` // Symbol
	TradeID         int64  `json:"t"` // Trade ID
	Price           string `json:"p"` // Price
	Quantity        string `json:"q"` // Quantity
	BuyerOrderID    int64  `json:"b"` // Buyer order ID
	SellerOrderID   int64  `json:"a"` // Seller order ID
	TradeTime       int64  `json:"T"` // Trade time
	BuyerIsMaker    bool   `json:"m"` // Buyer is maker
	Ignore          bool   `json:"M"` // Ignore
	AggregatedTrade bool   `json:"-"` // True if aggregated trade
}

// ToDomain converts WSTrade to domain.Trade.
func (t *WSTrade) ToDomain(exchange string) (*domain.Trade, error) {
	symbol := domain.NormalizeSymbol(t.Symbol)

	price, err := domain.NewDecimal(t.Price)
	if err != nil {
		return nil, fmt.Errorf("parse price: %w", err)
	}

	qty, err := domain.NewDecimal(t.Quantity)
	if err != nil {
		return nil, fmt.Errorf("parse quantity: %w", err)
	}

	// Calculate quote quantity
	quoteQty := domain.Mul(price, qty)

	side := domain.OrderSideBuy
	if t.BuyerIsMaker {
		side = domain.OrderSideSell // If buyer is maker, seller is taker (sell)
	}

	return &domain.Trade{
		Exchange:      exchange,
		Symbol:        symbol,
		ID:            fmt.Sprintf("%d", t.TradeID),
		Price:         price,
		Quantity:      qty,
		QuoteQuantity: quoteQty,
		Side:          side,
		IsMaker:       t.BuyerIsMaker,
		Timestamp:     time.UnixMilli(t.TradeTime),
	}, nil
}

// WSOrderUpdate represents an order update from the user data stream.
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#payload-order-update
type WSOrderUpdate struct {
	EventType           string `json:"e"` // Event type (executionReport)
	EventTime           int64  `json:"E"` // Event time
	Symbol              string `json:"s"` // Symbol
	ClientOrderID       string `json:"c"` // Client order ID
	Side                string `json:"S"` // Side (BUY or SELL)
	OrderType           string `json:"o"` // Order type
	TimeInForce         string `json:"f"` // Time in force
	OriginalQuantity    string `json:"q"` // Original quantity
	OriginalPrice       string `json:"p"` // Original price
	AveragePrice        string `json:"a"` // Average price
	OrderStatus         string `json:"X"` // Order status
	LastFilledQuantity  string `json:"l"` // Last filled quantity
	CumulativeFilledQty string `json:"z"` // Cumulative filled quantity
	LastFilledPrice     string `json:"L"` // Last filled price
	CommissionAmount    string `json:"n"` // Commission amount
	CommissionAsset     string `json:"N"` // Commission asset
	TradeTime           int64  `json:"T"` // Trade time
	TradeID             int64  `json:"t"` // Trade ID
	Ignore1             int64  `json:"I"` // Ignore
	Ignore2             bool   `json:"w"` // Ignore
	IsMaker             bool   `json:"m"` // Is maker
	Ignore3             bool   `json:"M"` // Ignore
	OrderCreationTime   int64  `json:"O"` // Order creation time
	CumulativeQuoteQty  string `json:"Z"` // Cumulative quote quantity
	LastQuoteQty        string `json:"Y"` // Last quote asset quantity
	OrderID             int64  `json:"i"` // Order ID
	LastOrderRejectTime int64  `json:"C"` // Last order reject time
	PriceProtect        bool   `json:"P"` // Price protect
	OriginalQuoteQty    string `json:"Q"` // Original quote order quantity
	WorkingTime         int64  `json:"W"` // Working time
	SelfTradePrevention string `json:"V"` // Self trade prevention mode
}

// ToDomain converts WSOrderUpdate to domain.Order.
func (o *WSOrderUpdate) ToDomain(exchange string) (*domain.Order, error) {
	symbol := domain.NormalizeSymbol(o.Symbol)

	// Parse side
	var side domain.OrderSide
	switch o.Side {
	case "BUY":
		side = domain.OrderSideBuy
	case "SELL":
		side = domain.OrderSideSell
	default:
		return nil, fmt.Errorf("invalid order side: %s", o.Side)
	}

	// Parse type
	var orderType domain.OrderType
	switch o.OrderType {
	case "LIMIT":
		orderType = domain.OrderTypeLimit
	case "MARKET":
		orderType = domain.OrderTypeMarket
	default:
		orderType = domain.OrderType(o.OrderType)
	}

	// Parse status
	var status domain.OrderStatus
	switch o.OrderStatus {
	case "NEW":
		status = domain.OrderStatusNew
	case "PARTIALLY_FILLED":
		status = domain.OrderStatusPartiallyFilled
	case "FILLED":
		status = domain.OrderStatusFilled
	case "CANCELED":
		status = domain.OrderStatusCanceled
	case "PENDING_CANCEL":
		status = domain.OrderStatusCanceling
	case "REJECTED":
		status = domain.OrderStatusRejected
	case "EXPIRED":
		status = domain.OrderStatusExpired
	default:
		status = domain.OrderStatus(o.OrderStatus)
	}

	// Parse decimals
	price, _ := domain.NewDecimal(o.OriginalPrice)
	if price == nil {
		price = domain.Zero()
	}

	qty, err := domain.NewDecimal(o.OriginalQuantity)
	if err != nil {
		return nil, fmt.Errorf("parse original_quantity: %w", err)
	}

	filledQty, err := domain.NewDecimal(o.CumulativeFilledQty)
	if err != nil {
		return nil, fmt.Errorf("parse cumulative_filled_qty: %w", err)
	}

	quoteQty, _ := domain.NewDecimal(o.CumulativeQuoteQty)
	if quoteQty == nil {
		quoteQty = domain.Zero()
	}

	commission, _ := domain.NewDecimal(o.CommissionAmount)
	if commission == nil {
		commission = domain.Zero()
	}

	return &domain.Order{
		Exchange:        exchange,
		Symbol:          symbol,
		ID:              fmt.Sprintf("%d", o.OrderID),
		ClientOrderID:   o.ClientOrderID,
		Side:            side,
		Type:            orderType,
		Status:          status,
		Price:           price,
		Quantity:        qty,
		FilledQuantity:  filledQty,
		QuoteQuantity:   quoteQty,
		Commission:      commission,
		CommissionAsset: o.CommissionAsset,
		CreatedAt:       time.UnixMilli(o.OrderCreationTime),
		UpdatedAt:       time.UnixMilli(o.EventTime),
		TradeID:         fmt.Sprintf("%d", o.TradeID),
		IsWorking:       o.WorkingTime > 0 && !status.IsFinal(),
	}, nil
}

// WSBalanceUpdate represents a balance update from the user data stream.
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#payload-balance-update
type WSBalanceUpdate struct {
	EventType    string `json:"e"` // Event type (balanceUpdate)
	EventTime    int64  `json:"E"` // Event time
	Asset        string `json:"a"` // Asset
	BalanceDelta string `json:"d"` // Balance delta
	ClearTime    int64  `json:"T"` // Clear time
}

// WSKline represents a kline/candlestick update.
// WebSocket Stream: <symbol>@kline_<interval>
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#kline-candlestick-streams
type WSKline struct {
	EventType string `json:"e"` // Event type (kline)
	EventTime int64  `json:"E"` // Event time
	Symbol    string `json:"s"` // Symbol
	Kline     struct {
		StartTime           int64  `json:"t"` // Kline start time
		EndTime             int64  `json:"T"` // Kline close time
		Symbol              string `json:"s"` // Symbol
		Interval            string `json:"i"` // Interval
		FirstTradeID        int64  `json:"f"` // First trade ID
		LastTradeID         int64  `json:"L"` // Last trade ID
		OpenPrice           string `json:"o"` // Open price
		ClosePrice          string `json:"c"` // Close price
		HighPrice           string `json:"h"` // High price
		LowPrice            string `json:"l"` // Low price
		Volume              string `json:"v"` // Base asset volume
		TradeCount          int64  `json:"n"` // Number of trades
		IsClosed            bool   `json:"x"` // Is kline closed?
		QuoteVolume         string `json:"q"` // Quote asset volume
		TakerBuyBaseVolume  string `json:"V"` // Taker buy base asset volume
		TakerBuyQuoteVolume string `json:"Q"` // Taker buy quote asset volume
		Ignore              string `json:"B"` // Ignore
	} `json:"k"`
}

// ToDomain converts WSKline to domain.Kline.
func (k *WSKline) ToDomain(exchange string) (*domain.Kline, error) {
	symbol := domain.NormalizeSymbol(k.Symbol)

	open, err := domain.NewDecimal(k.Kline.OpenPrice)
	if err != nil {
		return nil, fmt.Errorf("parse open: %w", err)
	}

	close, err := domain.NewDecimal(k.Kline.ClosePrice)
	if err != nil {
		return nil, fmt.Errorf("parse close: %w", err)
	}

	high, err := domain.NewDecimal(k.Kline.HighPrice)
	if err != nil {
		return nil, fmt.Errorf("parse high: %w", err)
	}

	low, err := domain.NewDecimal(k.Kline.LowPrice)
	if err != nil {
		return nil, fmt.Errorf("parse low: %w", err)
	}

	volume, err := domain.NewDecimal(k.Kline.Volume)
	if err != nil {
		return nil, fmt.Errorf("parse volume: %w", err)
	}

	quoteVolume, err := domain.NewDecimal(k.Kline.QuoteVolume)
	if err != nil {
		return nil, fmt.Errorf("parse quote_volume: %w", err)
	}

	takerBuyVolume, err := domain.NewDecimal(k.Kline.TakerBuyBaseVolume)
	if err != nil {
		return nil, fmt.Errorf("parse taker_buy_volume: %w", err)
	}

	takerBuyQuoteVolume, err := domain.NewDecimal(k.Kline.TakerBuyQuoteVolume)
	if err != nil {
		return nil, fmt.Errorf("parse taker_buy_quote_volume: %w", err)
	}

	return &domain.Kline{
		Exchange:            exchange,
		Symbol:              symbol,
		Interval:            k.Kline.Interval,
		OpenTime:            time.UnixMilli(k.Kline.StartTime),
		CloseTime:           time.UnixMilli(k.Kline.EndTime),
		Open:                open,
		High:                high,
		Low:                 low,
		Close:               close,
		Volume:              volume,
		QuoteVolume:         quoteVolume,
		TradeCount:          k.Kline.TradeCount,
		TakerBuyVolume:      takerBuyVolume,
		TakerBuyQuoteVolume: takerBuyQuoteVolume,
		IsClosed:            k.Kline.IsClosed,
	}, nil
}

// WSAggTrade represents an aggregated trade update.
// WebSocket Stream: <symbol>@aggTrade
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#aggregate-trade-streams
type WSAggTrade struct {
	EventType    string `json:"e"` // Event type (aggTrade)
	EventTime    int64  `json:"E"` // Event time
	Symbol       string `json:"s"` // Symbol
	AggTradeID   int64  `json:"a"` // Aggregate trade ID
	Price        string `json:"p"` // Price
	Quantity     string `json:"q"` // Quantity
	FirstTradeID int64  `json:"f"` // First trade ID
	LastTradeID  int64  `json:"l"` // Last trade ID
	TradeTime    int64  `json:"T"` // Trade time
	BuyerIsMaker bool   `json:"m"` // Buyer is maker
}

// ToDomain converts WSAggTrade to domain.Trade.
func (t *WSAggTrade) ToDomain(exchange string) (*domain.Trade, error) {
	symbol := domain.NormalizeSymbol(t.Symbol)

	price, err := domain.NewDecimal(t.Price)
	if err != nil {
		return nil, fmt.Errorf("parse price: %w", err)
	}

	qty, err := domain.NewDecimal(t.Quantity)
	if err != nil {
		return nil, fmt.Errorf("parse quantity: %w", err)
	}

	quoteQty := domain.Mul(price, qty)

	side := domain.OrderSideBuy
	if t.BuyerIsMaker {
		side = domain.OrderSideSell
	}

	return &domain.Trade{
		Exchange:      exchange,
		Symbol:        symbol,
		ID:            fmt.Sprintf("%d", t.AggTradeID),
		Price:         price,
		Quantity:      qty,
		QuoteQuantity: quoteQty,
		Side:          side,
		IsMaker:       t.BuyerIsMaker,
		Timestamp:     time.UnixMilli(t.TradeTime),
	}, nil
}
