package Loop_test

import (
	config "sensord_v2/config"
	"sensord_v2/db"
	loop "sensord_v2/loop"
	modbus "sensord_v2/modbus"
	"testing"
	"time"
)

func TestLoop(t *testing.T) {
	c := config.ConfigInit()
	err := c.Load("../config.toml")
	if err != nil {
		t.Fatal("error load config: ", err)
	}
	m, err := modbus.ModbusInit(c)
	if err != nil {
		t.Fatal("error ModbusInit: ", err)
	}

	d := db.DbInit(c)
	err = d.Connect()
	if err != nil {
		t.Fatal("error db connect")
	}

	l := loop.LOOPInit(c)
	l.Loop(c, m, d)

	p := []*config.Probe{&c.PH, &c.COD, &c.TSS, &c.NH3N}
	for _, e := range p {
		e.ENABLE = true
	}

	time.Sleep(time.Second * 3)

	for _, e := range p {
		t.Log("value ", e.NAME, ": ", e.VALUE)
	}
}
