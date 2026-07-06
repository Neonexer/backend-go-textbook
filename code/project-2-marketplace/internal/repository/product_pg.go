package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-course/project-2-marketplace/internal/model"
)

// ProductRepo — репозиторий товаров на pgx.
type ProductRepo struct {
	pool *pgxpool.Pool
}

func NewProduct(pool *pgxpool.Pool) *ProductRepo {
	return &ProductRepo{pool: pool}
}

func (r *ProductRepo) FindAll(ctx context.Context) ([]model.Product, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, description, price, seller_id, created_at, updated_at
		 FROM products ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.Title, &p.Description, &p.Price,
			&p.SellerID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}
	return products, rows.Err()
}

func (r *ProductRepo) FindByID(ctx context.Context, id int) (model.Product, error) {
	var p model.Product
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, description, price, seller_id, created_at, updated_at
		 FROM products WHERE id = $1`, id,
	).Scan(&p.ID, &p.Title, &p.Description, &p.Price,
		&p.SellerID, &p.CreatedAt, &p.UpdatedAt)

	if err == pgx.ErrNoRows {
		return model.Product{}, fmt.Errorf("product not found")
	}
	return p, err
}

func (r *ProductRepo) Create(ctx context.Context, p model.Product) (model.Product, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO products (title, description, price, seller_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		p.Title, p.Description, p.Price, p.SellerID,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)

	return p, err
}

func (r *ProductRepo) Update(ctx context.Context, p model.Product) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE products SET title=$1, description=$2, price=$3, updated_at=NOW()
		 WHERE id=$4 AND seller_id=$5`,
		p.Title, p.Description, p.Price, p.ID, p.SellerID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product not found or not owned")
	}
	return nil
}

func (r *ProductRepo) Delete(ctx context.Context, id, sellerID int) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM products WHERE id=$1 AND seller_id=$2`,
		id, sellerID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}
