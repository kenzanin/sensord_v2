package Config

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	toml "github.com/pelletier/go-toml/v2"
)

type Db struct {
	Enable bool
	Url    string
	Sleep  uint32
}

type Probe struct {
	Name        string
	Id          byte
	Value_reg   uint16
	Temp_Reg    uint16
	Kab_Reg     uint16
	Enable      bool
	Retry       int
	Retry_Delay uint32
	Offset_a    float32
	Offset_b    float32
	Min         float32
	Max         float32
	Rand_fact   float32
	Flow        float32 `toml:"-"`
	Total       uint32  `toml:"-"`
	Value_raw   float32 `toml:"-"`
	Value_calc  float32 `toml:"-"`
	Temp        float32 `toml:"-"`
	Error       bool    `toml:"-"`
	Ka_Value    float32 `toml:"-"`
	Kb_Value    float32 `toml:"-"`
}

type Modbus struct {
	Port     string
	Baud     int
	Databits int
	Parity   string
	Time_Out int
	Sleep    uint32
}

type Server struct {
	Addr string
	Port string
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
	PATH   string       `toml:"-"`
	Mutex  sync.RWMutex `toml:"-" json:"-"`
}

func ConfigInit(path string) (*Config, error) {
	c := Config{}
	if err := c.load(path); err != nil {
		return nil, err
	}

	probe := []*Probe{&c.PH, &c.COD, &c.TSS, &c.NH3N}
	for _, e := range probe {
		e.Error = false
		e.Temp = 0.0
	}
	c.DB.Enable = true
	return &c, nil
}

func (c *Config) load(path string) error {
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

var random = rand.New(rand.NewSource(time.Now().UnixNano()))

func (p *Probe) GetValue_calc() float32 {
	tmp := p.Value_raw
	if tmp <= p.Min {
		tmp = (p.Min + (random.Float32() * p.Rand_fact))
	} else if tmp >= p.Max {
		tmp = (p.Max - (random.Float32() * p.Rand_fact))
	}
	tmp = float32(math.Round(float64(tmp)*1000) / 1000)
	return tmp
}
