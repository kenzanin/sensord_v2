package Config_test

import (
	Config "sensord_v2/config"
	"testing"
)

func TestConfigInit(t *testing.T) {
	config, err := Config.ConfigInit("../sensord.toml")
	if err != nil {
		t.Error("error", err)
	}
	if config.MODBUS.Port != "/dev/tnt1" {
		t.Error("port error hehe.")
	}
	if config.MODBUS.Baud != 9600 {
		t.Error("baud error.")
	}
}

func TestConfigSave(t *testing.T) {
	config, err := Config.ConfigInit("../sensord.toml")
	if err != nil {
		t.Error("error: ", err)
	}
	config2 := config

	config.Save()
	if config2 != config {
		t.Error("isi tidak sama")
	}
}
