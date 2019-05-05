package migtool

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type config struct {
	User     string `json:"user"`
	Password string `json:"password"`
	DB       string `json:"database"`
	Port     string `json:"port"`
	Host     string `json:"host"`
}

type postgres struct {
	Postgres config `json:"postgres"`
}

// LoadConfig loads the config.json data in the Config global configuration variable
func loadConfig() *config {
	p := postgres{}
	raw, err := ioutil.ReadFile("./config.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if err = json.Unmarshal(raw, &p); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	if !configOK(&p.Postgres) {
		fmt.Println("check database config")
		os.Exit(1)
	}
	return &p.Postgres
}

func configOK(c *config) bool {
	switch {
	case c.Host == "":
		return false
	case c.Port == "":
		return false
	case c.User == "":
		return false
	case c.Password == "":
		return false
	case c.DB == "":
		return false
	}
	return true
}
