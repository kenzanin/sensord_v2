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

func ModbusInit(c *config.Config) (*MODBUS, error) {
	h := &MODBUS{
		handler: modbus.NewRTUClientHandler(c.MODBUS.Port),
		mutex:   &c.Mutex,
		c:       c,
	}
	h.handler.Timeout = time.Millisecond * time.Duration(c.MODBUS.Time_Out)
	h.handler.BaudRate = c.MODBUS.Baud
	h.handler.DataBits = c.MODBUS.Databits
	h.handler.Parity = c.MODBUS.Parity
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
	h.SlaveId = c.Id
	client := modbus.NewClient(h)
	var err error
	value := float32(0.0)
	temperature := float32(0.0)
	for i := 0; i < c.Retry; i += 1 {
		res, er := client.ReadHoldingRegisters(c.Value_reg, 2)
		err = er
		if er != nil {
			log.Print("error reading ", c.Name, " value: ", err)
			time.Sleep(time.Millisecond * time.Duration(c.Retry_Delay))
			continue
		}
		tmp01 := binary.LittleEndian.Uint32(res)
		value = math.Float32frombits(tmp01)
		log.Print("Succes reading ", c.Name, " value[]: ", hex.EncodeToString(res), " value: ", value)
		break
	}

	if c.Temp_Reg != 0 {
		for i := 0; i < c.Retry; i += 1 {
			res, er := client.ReadHoldingRegisters(c.Temp_Reg, 2)
			err = er
			if er != nil {
				log.Print("error reading ", c.Name, " temp: ", err)
				time.Sleep(time.Millisecond * time.Duration(c.Retry_Delay))
				continue
			}
			tmp01 := binary.LittleEndian.Uint32(res)
			temperature = math.Float32frombits(tmp01)
			log.Print("Succes reading ", c.Name, " temp[]: ", hex.EncodeToString(res), " temp: ", temperature)
			break
		}
	}
	time.Sleep(time.Duration(c.Retry_Delay) * time.Millisecond)
	return value, temperature, err
}

func (m *MODBUS) ReadFlow(c *config.Probe) (float32, uint32, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	h := m.handler
	h.SlaveId = c.Id
	client := modbus.NewClient(h)
	var err error
	value := float32(0.0)
	total := uint32(0)
	for i := 0; i < c.Retry; i += 1 {
		res, er := client.ReadHoldingRegisters(c.Value_reg, 4)
		err = er
		if er != nil {
			log.Print("error reading ", c.Name, " flow and total: ", err)
			time.Sleep(time.Duration(c.Retry_Delay) * time.Millisecond)
			continue
		}
		tmp01 := binary.LittleEndian.Uint32(res[0:1])
		value = math.Float32frombits(tmp01)
		total = binary.LittleEndian.Uint32(res[2:3])
		log.Print("Succes reading ", c.Name, " flow and total : ", res[0:3], " ", res[4:7])
		break
	}
	time.Sleep(time.Duration(c.Retry_Delay) * time.Millisecond)
	return value, total, err
}

func (m *MODBUS) ReadKAB(c *config.Probe) (float32, float32, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	h := m.handler
	h.SlaveId = c.Id
	client := modbus.NewClient(h)
	var err error
	ka := float32(0)
	kb := float32(0)
	for i := 0; i < c.Retry; i += 1 {
		res, er := client.ReadHoldingRegisters(c.Kab_Reg, 4)
		err = er
		if err != nil {
			log.Print("error reading KAB ", c.Name, ": ", err)
			time.Sleep(time.Duration(c.Retry_Delay) * time.Millisecond)
			continue
		}
		tmp01 := binary.LittleEndian.Uint32(res[0:1])
		ka = math.Float32frombits(tmp01)
		tmp01 = binary.LittleEndian.Uint32(res[2:3])
		kb = math.Float32frombits(tmp01)
		log.Print("Success reading KA, KB ", c.Name, " ka: ", res[0:1], " kb: ", res[1:2])
		break
	}
	time.Sleep(time.Duration(c.Retry_Delay) * time.Millisecond)
	return ka, kb, err
}

func (m *MODBUS) WriteKB(c *config.Probe, ka float32, kb float32) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	h := m.handler
	h.SlaveId = c.Id
	client := modbus.NewClient(h)
	var01 := make([]byte, 4)
	binary.LittleEndian.PutUint32(var01[0:1], math.Float32bits(ka))
	binary.LittleEndian.PutUint32(var01[2:3], math.Float32bits(kb))
	var err error
	var res []byte
	for i := 0; i < c.Retry; i += 1 {
		r, er := client.WriteMultipleRegisters(c.Kab_Reg, 4, var01)
		err = er
		if err != nil {
			log.Print("error writing ka, kb ", c.Name, ": ", err)
			time.Sleep(time.Millisecond * time.Duration(c.Retry_Delay))
			continue
		}
		res = r
		log.Print("sucsess writing ka, kb", c.Name, " ka: ", res[0:1], " kb: ", res[2:3])
		break
	}
	return err
}
