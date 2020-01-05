package migtool

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

func (c *Client) Init() error {
	tx, err := c.db.Beginx()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
	CREATE TABLE IF NOT EXISTS migtool (
		version INT,
		last_update TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc')
	);`)
	if err != nil {
		tx.Rollback()
		return err
	}

	var t time.Time
	rows, err := tx.Queryx("SELECT last_update FROM migtool;")
	if err != nil {
		return err
	}
	for rows.Next() {
		err := rows.Scan(&t)
		if err != nil {
			return err
		}
	}

	if t.IsZero() {
		_, err = tx.Exec("INSERT INTO migtool (version, last_update) VALUES (0, $1);", c.now)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()

	return nil
}

func (c *Client) GetCurrentVersion() (int, error) {
	var version int
	rows, err := c.db.Queryx("SELECT version FROM migtool;")
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		err := rows.Scan(&version)
		if err != nil {
			return 0, err
		}
	}
	return version, nil
}

type migration struct {
	version   int
	direction string
	fileName  string
}

type migtoolInfo struct {
	currentVersion int
	migrations     []migration
}

func (c *Client) getMigrations() (*migtoolInfo, error) {
	var err error
	mInfo := migtoolInfo{}
	mInfo.currentVersion, err = c.GetCurrentVersion()
	if err != nil {
		return nil, err
	}

	f, err := ioutil.ReadDir("migrations")
	if err != nil {
		return nil, err
	}

	for _, file := range f {
		split := strings.Split(file.Name(), "_")
		direction := strings.Split(split[len(split)-1], ".")
		version, err := strconv.Atoi(split[0])
		if err != nil {
			return nil, err
		}
		mInfo.migrations = append(
			mInfo.migrations,
			migration{
				version:   version,
				direction: direction[0],
				fileName:  file.Name(),
			},
		)
	}
	return &mInfo, nil
}

func (c *Client) MigrateUp(toVersion int) error {
	mInfo, err := c.getMigrations()
	if err != nil {
		return err
	}
	if toVersion > mInfo.currentVersion {
		for _, m := range mInfo.migrations {
			if m.version <= mInfo.currentVersion {
				continue
			}
			if m.version > toVersion {
				continue
			}
			if m.direction != "up" {
				continue
			}
			b, err := ioutil.ReadFile("migrations/" + m.fileName)
			if err != nil {
				return err
			}
			err = c.execute(b)
			if err != nil {
				return err
			}
			err = c.SetVersion(m.version)
			if err != nil {
				return err
			}
		}
	} else {
		return errors.New("version number is not greater than current version")
	}

	return nil
}

func (c *Client) MigrateDown(toVersion int) error {
	mInfo, err := c.getMigrations()
	if err != nil {
		return err
	}
	if toVersion < mInfo.currentVersion {
		for i := len(mInfo.migrations) - 1; i >= 0; i-- {
			m := mInfo.migrations[i]
			if m.version > mInfo.currentVersion {
				continue
			}
			if m.version <= toVersion {
				continue
			}
			if m.direction != "down" {
				continue
			}
			b, err := ioutil.ReadFile("migrations/" + m.fileName)
			if err != nil {
				return err
			}
			err = c.execute(b)
			if err != nil {
				return err
			}
			if i == 0 {
				err = c.SetVersion(0)
				if err != nil {
					return err
				}
			} else {
				err = c.SetVersion(mInfo.migrations[i-1].version)
				if err != nil {
					return err
				}
			}

		}
	} else {
		return errors.New("version number is not smaller than current version")
	}

	return nil
}

func (c *Client) New(name string, tables ...string) error {
	_, err := ioutil.ReadDir("migrations")
	if os.IsNotExist(err) {
		err := os.Mkdir("migrations", os.ModePerm)
		if err != nil {
			return err
		}
		_, err = ioutil.ReadDir("migrations")
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	up := fmt.Sprintf("migrations/%d_%s_up.sql", c.now.Unix(), name)
	down := fmt.Sprintf("migrations/%d_%s_down.sql", c.now.Unix(), name)

	create := createTables(tables)
	drop := dropTables(tables)

	err = ioutil.WriteFile(up, create, os.ModePerm)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(down, drop, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) SetVersion(version int) error {
	_, err := c.db.Exec("UPDATE migtool SET version = $1, last_update = $2", version, c.now)
	return err
}

func createTables(tables []string) []byte {
	var sql string
	for _, table := range tables {
		sql += fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n\n);\n\n", table)
	}
	return []byte(sql)
}

func dropTables(tables []string) []byte {
	var sql string
	for _, table := range tables {
		sql += fmt.Sprintf("DROP TABLE IF EXISTS %s;\n", table)
	}
	return []byte(sql)
}

func (c *Client) execute(sql []byte) error {
	_, err := c.db.Exec(string(sql))
	return err
}
