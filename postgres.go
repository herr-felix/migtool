package migtool

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type Client struct {
	db  *sqlx.DB
	now time.Time
}

func newClient() *Client {
	c := &Client{
		now: time.Now(),
	}

	return c
}

func Connect(connectionString string) (*Client, error) {
	c := newClient()
	err := c.open(connectionString)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) open(connectionString string) error {
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return err
	}

	if db == nil {
		return fmt.Errorf("error opening database: %s", err)
	}

	err = db.Ping()
	if err != nil {
		return err
	}
	c.db = db
	return nil
}

func (c *Client) Disconnect() error {
	return c.db.Close()
}
