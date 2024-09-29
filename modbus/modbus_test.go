package Modbus_test

import (
	config "sensord_v2/config"
	modbus "sensord_v2/modbus"
	"testing"
)

func TestModbusInit(t *testing.T) {
	c, err := config.ConfigInit("../sensord.toml")
	if err != nil {
		t.Error("error", err)
	}

	m, err := modbus.ModbusInit(c)
	if err != nil {
		t.Fatal("error init modbus: ", err)
	}
	_ = m
}

func TestReadingPH(t *testing.T) {
	c, err := config.ConfigInit("../sensord.toml")
	if err != nil {
		t.Error("error: ", err)
	}
	m, err := modbus.ModbusInit(c)
	if err != nil {
		t.Fatal("error init modbus: ", err)
	}

	_, _, err = m.ReadFloat32(&c.PH)
	if err != nil {
		t.Fatal("ReadFloat32 error: ", err)
	}
}
