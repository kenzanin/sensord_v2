package http_test

import (
	config "sensord_v2/config"
	"sensord_v2/db"
	server "sensord_v2/http"
	modbus "sensord_v2/modbus"
	"testing"
)

func TestServer(t *testing.T) {
	c, err := config.ConfigInit("../sensord.toml")
	if err != nil {
		t.Error("error", err)
	}
	m, _ := modbus.ModbusInit(c)
	d := db.DbInit(c)
	err = d.Connect()
	if err != nil {
		t.Fatal("error db connect")
	}

	m.Loop()
	d.Loop()
	s := server.ServerInit(c)
	s.Serve()
}
