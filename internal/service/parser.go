package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/requel1/price-tracker/internal/model"
	"github.com/requel1/price-tracker/internal/parser"
	"github.com/requel1/price-tracker/internal/repository"
)

type ParserService struct {
	productRepo *repository.ProductRepo
	priceRepo   *repository.PriceRepo
	workers     int
}

func NewParserService(
	productRepo *repository.ProductRepo,
	priceRepo *repository.PriceRepo,
	workers int,
) *ParserService {
	return &ParserService{
		productRepo: productRepo,
		priceRepo:   priceRepo,
		workers:     workers,
	}
}

// ParseAll берёт все товары, парсит цены параллельно и сохраняет.
func (s *ParserService) ParseAll(ctx context.Context) error {
	products, err := s.productRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("list products: %w", err)
	}

	if len(products) == 0 {
		slog.Info("no products to parse")
		return nil
	}

	slog.Info("starting parse", "products", len(products), "workers", s.workers)

	jobs := make(chan *model.Product, len(products))
	for _, p := range products {
		jobs <- p
	}
	close(jobs)

	var wg sync.WaitGroup
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				s.parseOne(ctx, p)
			}
		}()
	}

	wg.Wait()
	slog.Info("parse finished")
	return nil
}

func (s *ParserService) parseOne(ctx context.Context, p *model.Product) {
	prs, err := parser.GetParser(p.Site)
	if err != nil {
		slog.Warn("parser not found", "site", p.Site, "product_id", p.ID, "err", err)
		return
	}

	price, err := prs.Parse(ctx, p.URL)
	if err != nil {
		slog.Warn("parse failed", "product_id", p.ID, "url", p.URL, "err", err)
		return
	}

	price.ProductID = p.ID
	if err := s.priceRepo.Create(ctx, price); err != nil {
		slog.Error("save price failed", "product_id", p.ID, "err", err)
		return
	}

	slog.Info("price saved", "product_id", p.ID, "price", price.Price, "currency", price.Currency)
}

// RunBySchedule запускает ParseAll по расписанию.
func (s *ParserService) RunBySchedule(ctx context.Context, interval time.Duration) {
	s.ParseAll(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("scheduler stopped")
			return
		case <-ticker.C:
			s.ParseAll(ctx)
		}
	}
}
