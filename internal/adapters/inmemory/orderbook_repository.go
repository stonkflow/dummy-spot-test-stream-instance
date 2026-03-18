package inmemory

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"dummy-spot-test-stream-instance/internal/domain"
)

type OrderBookRepository struct {
	mu    sync.RWMutex
	books map[string]*orderBook
}

type orderBook struct {
	lastUpdateID int64
	bids         map[string]string
	asks         map[string]string
}

func NewOrderBookRepository() *OrderBookRepository {
	return &OrderBookRepository{
		books: make(map[string]*orderBook),
	}
}

func (repository *OrderBookRepository) UpsertDepth(_ context.Context, depth domain.DepthEvent) (domain.DepthEvent, error) {
	if depth.Symbol == "" {
		return domain.DepthEvent{}, fmt.Errorf("depth symbol is required")
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()

	book := repository.books[depth.Symbol]
	if book == nil {
		book = &orderBook{
			bids: make(map[string]string),
			asks: make(map[string]string),
		}
		repository.books[depth.Symbol] = book
	}

	applyLevels(book.bids, depth.Bids)
	applyLevels(book.asks, depth.Asks)
	if depth.LastUpdateID > book.lastUpdateID {
		book.lastUpdateID = depth.LastUpdateID
	}

	return domain.DepthEvent{
		RequestID:    depth.RequestID,
		Symbol:       depth.Symbol,
		EventTimeMS:  depth.EventTimeMS,
		LastUpdateID: book.lastUpdateID,
		Bids:         mapToLevels(book.bids, true),
		Asks:         mapToLevels(book.asks, false),
	}, nil
}

func (repository *OrderBookRepository) Remove(_ context.Context, symbol string) error {
	if symbol == "" {
		return nil
	}

	repository.mu.Lock()
	defer repository.mu.Unlock()

	delete(repository.books, symbol)
	return nil
}

func (repository *OrderBookRepository) Close() error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	repository.books = make(map[string]*orderBook)
	return nil
}

func applyLevels(target map[string]string, levels []domain.DepthLevel) {
	for _, level := range levels {
		price := strings.TrimSpace(level.Price)
		quantity := strings.TrimSpace(level.Quantity)
		if price == "" {
			continue
		}
		if quantity == "" || isZero(quantity) {
			delete(target, price)
			continue
		}
		target[price] = quantity
	}
}

type sortableLevel struct {
	price     string
	quantity  string
	numeric   float64
	isNumeric bool
}

func mapToLevels(source map[string]string, desc bool) []domain.DepthLevel {
	items := make([]sortableLevel, 0, len(source))
	for price, quantity := range source {
		numericPrice, err := strconv.ParseFloat(price, 64)
		items = append(items, sortableLevel{
			price:     price,
			quantity:  quantity,
			numeric:   numericPrice,
			isNumeric: err == nil,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		left := items[i]
		right := items[j]

		if left.isNumeric && right.isNumeric {
			if desc {
				return left.numeric > right.numeric
			}
			return left.numeric < right.numeric
		}

		if desc {
			return left.price > right.price
		}
		return left.price < right.price
	})

	levels := make([]domain.DepthLevel, 0, len(items))
	for _, item := range items {
		levels = append(levels, domain.DepthLevel{
			Price:    item.price,
			Quantity: item.quantity,
		})
	}
	return levels
}

func isZero(value string) bool {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		trimmed := strings.Trim(value, "0.")
		return trimmed == ""
	}
	return parsed == 0
}
