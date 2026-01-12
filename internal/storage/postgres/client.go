package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/baseplate/baseplate/config"
)

type Client struct {
	DB *sql.DB
}

func NewClient(cfg *config.DatabaseConfig) (*Client, error) {
	db, err := sql.Open("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &Client{DB: db}, nil
}

func (c *Client) Close() error {
	return c.DB.Close()
}
