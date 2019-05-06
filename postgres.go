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

func Connect() (*Client, error) {
	c := newClient()
	err := c.open()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) open() error {
	config := loadConfig()
	// Get DB config from environment variables
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DB,
	)
	db, err := sqlx.Connect("postgres", psqlInfo)
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
