package main

import (
	"log"
	"os"
	config "sensord_v2/config"
	db "sensord_v2/db"
	server "sensord_v2/http"
	modbus "sensord_v2/modbus"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:      "Sensord",
		Usage:     "an terminal app to read modbus, store in db and provide http api",
		UsageText: "sensord /path/to/sensord.toml",
		Copyright: "kenzanin@gmail.com",
		Action: func(ctx *cli.Context) error {
			cfg := ctx.Args().Get(0)
			if len(cfg) == 0 {
				log.Fatal("error config not profided. see usage --help")
			}

			c, err := config.ConfigInit(cfg)
			if err != nil {
				log.Fatal("error: ", err)
			}
			m, err := modbus.ModbusInit(c)
			if err != nil {
				log.Fatal("error init modbus: ", err)
			}

			d := db.DbInit(c)
			d.Connect()

			h := server.ServerInit(c)

			d.Loop()
			m.Loop()
			h.Serve()
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
