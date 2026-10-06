package parser

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"

	"github.com/requel1/price-tracker/internal/model"
)

type ChitaiGorodParser struct{}

func NewChitaiGorodParser() *ChitaiGorodParser {
	return &ChitaiGorodParser{}
}

func (p *ChitaiGorodParser) Parse(ctx context.Context, url string) (*model.Price, error) {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)
	c.SetRequestTimeout(10 * time.Second)

	var priceStr string

	c.OnHTML(".product-offer-price__actual", func(e *colly.HTMLElement) {
		if priceStr == "" {
			priceStr = e.Text
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("parse error: %v\n", err)
	})

	if err := c.Visit(url); err != nil {
		return nil, fmt.Errorf("visit: %w", err)
	}

	if priceStr == "" {
		return nil, fmt.Errorf("price not found")
	}

	price, err := parsePrice(priceStr)
	if err != nil {
		return nil, fmt.Errorf("parse price %q: %w", priceStr, err)
	}

	return &model.Price{
		Price:    price,
		Currency: "RUB",
		ParsedAt: time.Now(),
	}, nil
}

func parsePrice(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\u00a0", "")
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "₽", "")
	s = strings.ReplaceAll(s, "руб.", "")
	s = strings.ReplaceAll(s, ",", ".")

	return strconv.ParseFloat(s, 64)
}
