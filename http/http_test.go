package http_test

import (
	config "sensord_v2/config"
	server "sensord_v2/http"
	loop "sensord_v2/loop"
	modbus "sensord_v2/modbus"
	"testing"
)

func TestServer(t *testing.T) {
	c := config.ConfigInit()
	c.Load("../config.toml")
	mod, _ := modbus.ModbusInit(&c.Mutex, &c.MODBUS)
	l := loop.LOOPInit()
	l.Loop(c, mod)
	p := []*config.Probe{&c.PH, &c.COD, &c.TSS, &c.NH3N}
	for _, e := range p {
		e.ENABLE = true
	}
	s := server.ServerInit(c)
	s.Server()
}
