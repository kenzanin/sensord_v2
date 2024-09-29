package db_test

import (
	config "sensord_v2/config"
	"sensord_v2/db"
	"testing"
)

func TestDb(t *testing.T) {
	c, err := config.ConfigInit("../sensord.toml")
	if err != nil {
		t.Fatal("error load config")
	}
	d := db.DbInit(c)
	err = d.Connect()
	if err != nil {
		t.Fatal("error connect db")
	}
	defer d.Close()

	err = d.Insert()
	if err != nil {
		t.Fatal("error insert db")
	}
}
