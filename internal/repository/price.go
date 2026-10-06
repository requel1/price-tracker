package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/requel1/price-tracker/internal/model"
)

type PriceRepo struct {
	pool *pgxpool.Pool
}

func NewPriceRepo(pool *pgxpool.Pool) *PriceRepo {
	return &PriceRepo{pool: pool}
}

func (r *PriceRepo) Create(ctx context.Context, p *model.Price) error {
	query := `
	INSERT INTO prices(product_id,price,currency,parsed_at)
	VALUES($1,$2,$3,$4)
	`
	_, err := r.pool.Exec(ctx, query, p.ProductID, p.Price, p.Currency, p.ParsedAt)
	return err
}

func (r *PriceRepo) GetByProductID(ctx context.Context, productID int64) ([]*model.Price, error) {
	query := `
	SELECT id,product_id,price,currency,parsed_at
	FROM prices 
	WHERE product_id = $1
	ORDER BY parsed_at DESC
	`
	rows, err := r.pool.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []*model.Price
	for rows.Next() {
		p := &model.Price{}
		if err := rows.Scan(&p.ID, &p.ProductID, &p.Price, &p.Currency, &p.ParsedAt); err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}
	return prices, rows.Err()

}
