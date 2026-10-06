package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/requel1/price-tracker/internal/model"
)

var ErrNotFound = errors.New("not found")

type ProductRepo struct {
	pool *pgxpool.Pool
}

func NewProductRepo(pool *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{pool: pool}
}

func (r *ProductRepo) Create(ctx context.Context, p *model.Product) (*model.Product, error) {
	query := ` INSERT INTO products(url,name,site)
	VALUES ($1,$2,$3)
	RETURNING id,created_at
	`
	err := r.pool.QueryRow(ctx, query, p.URL, p.Name, p.Site).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProductRepo) GetByID(ctx context.Context, id int64) (*model.Product, error) {
	query := `SELECT id,url,name,site,created_at FROM products WHERE id = $1`
	p := &model.Product{}
	err := r.pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.URL, &p.Name, &p.Site, &p.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProductRepo) List(ctx context.Context) ([]*model.Product, error) {
	query := `SELECT id,url,name,site,created_at FROM products ORDER BY id`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []*model.Product
	for rows.Next() {
		p := &model.Product{}
		if err := rows.Scan(&p.ID, &p.URL, &p.Name, &p.Site, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *ProductRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
