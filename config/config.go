package Config

import (
	"errors"
	"log"
	"os"
	"sync"

	toml "github.com/pelletier/go-toml/v2"
)

type Loop struct {
	LOOP_READER uint32
	LOOP_DB     uint32
}

type Db struct {
	Enable bool
	Db_Url string
	Loop   uint32
}

type Probe struct {
	NAME       string
	SLAVEID    byte
	VALUE_REG  uint16
	TEMP_REG   uint16
	KAB_REG    uint16
	ENABLE     bool
	READ_RETRY int
	FLOW       float32 `toml:"-"`
	TOTAL      uint32  `toml:"-"`
	VALUE      float32 `toml:"-"`
	TEMP       float32 `toml:"-"`
	ERROR      bool    `toml:"-"`
	KA_VALUE   float32 `toml:"-"`
	KB_VALUE   float32 `toml:"-"`
}

type Modbus struct {
	PORT     string
	BAUD     int
	DATABITS int
	PARITY   string
	TIMEOUT  int
}

type Server struct {
	SERVER_ADDR string
	SERVER_PORT string
}

type Config struct {
	MODBUS Modbus
	PH     Probe
	COD    Probe
	NH3N   Probe
	TSS    Probe
	FLOW   Probe
	SERVER Server
	DB     Db
	LOOP   Loop
	PATH   string       `toml:"-"`
	Mutex  sync.RWMutex `toml:"-" json:"-"`
}

func ConfigInit() *Config {
	c := Config{}
	probe := []*Probe{&c.PH, &c.COD, &c.TSS, &c.NH3N}
	for _, e := range probe {
		e.ENABLE = true
		e.ERROR = false
		e.TEMP = 0.0
		e.VALUE = 0.0
	}
	c.DB.Enable = true
	return &c
}

func (c *Config) Load(path string) error {
	// read file
	toml_file, err := os.ReadFile(path)
	if err != nil {
		log.Fatal("error oppening config: ", err)
		return errors.New("error oppenning config.")
	}

	// unmarshal
	err = toml.Unmarshal(toml_file, &c)
	if err != nil {
		log.Fatal("error unmarshal json: ", err)
		return errors.New("error unmarshal json config. ")
	}
	c.PATH = path

	log.Printf("config content: %#v\n", c)
	return nil
}

func (c *Config) Save() error {
	toml_file, err := toml.Marshal(c)
	if err != nil {
		log.Fatal("error marshal json: ", err)
	}
	err = os.WriteFile(c.PATH, toml_file, 0644)
	if err != nil {
		log.Fatal("error writing config.json", err)
	}
	return nil
}
