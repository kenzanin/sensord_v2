package Modbus

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"log"
	"math"
	config "sensord_v2/config"
	"sync"
	"time"

	"github.com/goburrow/modbus"
)

type MODBUS struct {
	handler *modbus.RTUClientHandler
	mutex   *sync.RWMutex
	c       *config.Config
}

const (
	retry_delay = time.Millisecond * 250
)

func ModbusInit(c *config.Config) (*MODBUS, error) {
	h := &MODBUS{
		handler: modbus.NewRTUClientHandler(c.MODBUS.PORT),
		c:       c,
	}
	h.handler.Timeout = time.Millisecond * time.Duration(c.MODBUS.TIMEOUT)
	h.handler.BaudRate = c.MODBUS.BAUD
	h.handler.DataBits = c.MODBUS.DATABITS
	h.handler.Parity = c.MODBUS.PARITY
	err := h.handler.Connect()
	if err != nil {
		log.Fatal("Modbus init error: ", err)
		return nil, errors.New("Modbus init Error")
	}
	return h, nil
}

func (m *MODBUS) ReadFloat32(c *config.Probe) (float32, float32, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	h := m.handler
	h.SlaveId = c.SLAVEID
	client := modbus.NewClient(h)
	var err error
	value := float32(0.0)
	temperature := float32(0.0)
	for i := 0; i < c.READ_RETRY; i += 1 {
		res, er := client.ReadHoldingRegisters(c.VALUE_REG, 2)
		err = er
		if er != nil {
			log.Print("error reading ", c.NAME, " value: ", err)
			time.Sleep(time.Millisecond * time.Duration(m.c.MODBUS.TIMEOUT))
			continue
		}
		tmp01 := binary.LittleEndian.Uint32(res)
		value = math.Float32frombits(tmp01)
		log.Print("Succes reading ", c.NAME, " value[]: ", hex.EncodeToString(res), " value: ", value)
		break
	}

	if c.TEMP_REG != 0 {
		for i := 0; i < c.READ_RETRY; i += 1 {
			res, er := client.ReadHoldingRegisters(c.TEMP_REG, 2)
			err = er
			if er != nil {
				log.Print("error reading ", c.NAME, " temp: ", err)
				time.Sleep(time.Millisecond * time.Duration(m.c.MODBUS.TIMEOUT))
				continue
			}
			tmp01 := binary.LittleEndian.Uint32(res)
			temperature = math.Float32frombits(tmp01)
			log.Print("Succes reading ", c.NAME, " temp[]: ", hex.EncodeToString(res), " temp: ", temperature)
			break
		}
	}
	time.Sleep(retry_delay)
	return value, temperature, err
}

func (m *MODBUS) ReadFlow(c *config.Probe) (float32, uint32, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	h := m.handler
	h.SlaveId = c.SLAVEID
	client := modbus.NewClient(h)
	var err error
	value := float32(0.0)
	total := uint32(0)
	for i := 0; i < c.READ_RETRY; i += 1 {
		res, er := client.ReadHoldingRegisters(c.VALUE_REG, 4)
		err = er
		if er != nil {
			log.Print("error reading ", c.NAME, " flow and total: ", err)
			time.Sleep(retry_delay)
			continue
		}
		tmp01 := binary.LittleEndian.Uint32(res[0:1])
		value = math.Float32frombits(tmp01)
		total = binary.LittleEndian.Uint32(res[2:3])
		log.Print("Succes reading ", c.NAME, " flow and total : ", res[0:3], " ", res[4:7])
		break
	}
	time.Sleep(retry_delay)
	return value, total, err
}

func (m *MODBUS) ReadKAB(c *config.Probe) (float32, float32, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	h := m.handler
	h.SlaveId = c.SLAVEID
	client := modbus.NewClient(h)
	var err error
	ka := float32(0)
	kb := float32(0)
	for i := 0; i < c.READ_RETRY; i += 1 {
		res, er := client.ReadHoldingRegisters(c.KAB_REG, 4)
		err = er
		if err != nil {
			log.Print("error reading KAB ", c.NAME, ": ", err)
			time.Sleep(retry_delay)
			continue
		}
		tmp01 := binary.LittleEndian.Uint32(res[0:1])
		ka = math.Float32frombits(tmp01)
		tmp01 = binary.LittleEndian.Uint32(res[2:3])
		kb = math.Float32frombits(tmp01)
		log.Print("Success reading KA, KB ", c.NAME, " ka: ", res[0:1], " kb: ", res[1:2])
		break
	}
	return ka, kb, err
}

func (m *MODBUS) WriteKB(c *config.Probe, ka float32, kb float32) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	h := m.handler
	h.SlaveId = c.SLAVEID
	client := modbus.NewClient(h)
	var01 := make([]byte, 4)
	binary.LittleEndian.PutUint32(var01[0:1], math.Float32bits(ka))
	binary.LittleEndian.PutUint32(var01[2:3], math.Float32bits(kb))
	var err error
	var res []byte
	for i := 0; i < c.READ_RETRY; i += 1 {
		r, er := client.WriteMultipleRegisters(c.KAB_REG, 4, var01)
		err = er
		if err != nil {
			log.Print("error writing ka, kb ", c.NAME, ": ", err)
			time.Sleep(retry_delay)
			continue
		}
		res = r
		log.Print("sucsess writing ka, kb", c.NAME, " ka: ", res[0:1], " kb: ", res[2:3])
		break
	}
	return err
}
