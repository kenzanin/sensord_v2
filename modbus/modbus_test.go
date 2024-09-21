package Modbus_test

import (
	config "sensord_v2/config"
	modbus "sensord_v2/modbus"
	"testing"
)

func TestModbusInit(t *testing.T) {
	c := config.ConfigInit()
	c.Load("../config.toml")
	_, err := modbus.ModbusInit(&c.Mutex, &c.MODBUS)
	if err != nil {
		t.Fatal("error init modbus: ", err)
	}
}

func TestReadingPH(t *testing.T) {
	c := config.ConfigInit()
	c.Load("../config.toml")
	m, err := modbus.ModbusInit(&c.Mutex, &c.MODBUS)
	if err != nil {
		t.Fatal("error init modbus: ", err)
	}

	_, _, err = m.ReadFloat32(&c.PH)
	if err != nil {
		t.Fatal("ReadFloat32 error: ", err)
	}
}
