package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fredcarle/migtool"
	_ "github.com/lib/pq"
)

const fullUp int = 9999999999
const fullDow int = 0

type options struct {
	init      bool
	new       bool
	name      string
	migrate   bool
	full      bool
	direction string
	version   int
	help      bool
	tables    []string
}

func main() {
	o, err := getArgs(os.Args[1:])
	if err != nil {
		panic(err)
	}
	c, err := migtool.Connect()
	if err != nil {
		panic(err)
	}

	if o.help {
		fmt.Println(`
migtool [option]

    --help: Show this text.

    --init: Initialize migtool in the database

    new {name} [...tables]: Create a new migration with the given name and the optional comma seperated list of tables

    migrate [full] {up/down} {version}: execute the migration with the given parameters
`,
		)
		return
	}

	if o.init {
		fmt.Printf("Initializing migtool... ")
		err = c.Init()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Done!\n")
	}

	if o.new {
		fmt.Printf("Creating new migration file... ")
		err = c.New(o.name, o.tables...)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Done!\n")
		return
	}

	if o.migrate {
		fmt.Printf("Executing migration... ")
		if o.direction == "up" {
			if o.full {
				err = c.MigrateUp(fullUp)
				if err != nil {
					panic(err)
				}
			} else {
				err = c.MigrateUp(o.version)
				if err != nil {
					panic(err)
				}
			}
		} else if o.direction == "down" {
			if o.full {
				err = c.MigrateDown(fullDow)
				if err != nil {
					panic(err)
				}
			} else {
				err = c.MigrateDown(o.version)
				if err != nil {
					panic(err)
				}
			}
		}
		fmt.Printf("Done!\n")
		return
	}
}

func getArgs(args []string) (*options, error) {
	o := options{}
	if len(args) == 0 {
		o.help = true
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--help":
			o.help = true

		case "--init":
			o.init = true

		case "migrate":
			o.migrate = true
			i++
			if i >= len(args) {
				return nil, errors.New("please add migration direction: migtool migrate [full] {up/down} {version}")
			}
			if args[i] == "full" {
				o.full = true
				i++
				if i >= len(args) {
					return nil, errors.New("please add migration direction: migtool migrate [full] {up/down} {version}")
				}
			}
			switch args[i] {
			case "up":
				o.direction = "up"
				if o.full {
					continue
				}
			case "down":
				o.direction = "down"
				if o.full {
					continue
				}
			default:
				return nil, errors.New("please add migration direction: migtool migrate [full] {up/down} {version}")
			}

			i++
			if i >= len(args) {
				return nil, errors.New("please add migration version: migtool migrate [full] {up/down} {version}")
			}
			version, err := strconv.Atoi(args[i])
			if err != nil {
				return nil, err
			}
			o.version = version

		case "new":
			o.new = true
			i++
			if i >= len(args) {
				return nil, errors.New("please add migration name: migtool new {name} [...tables]")
			}
			o.name = args[i]
			i++
			if i >= len(args) {
				break
			}
			o.tables = strings.Split(args[i], ",")
		}
	}
	return &o, nil
}
