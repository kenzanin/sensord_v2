package Modbus

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"log"
	"math"
	config "sensord_v2/config"
	"time"

	"github.com/goburrow/modbus"
)

type MODBUS struct {
	handler *modbus.RTUClientHandler
	c       *config.Config
}

func ModbusInit(c *config.Config) (*MODBUS, error) {
	h := &MODBUS{
		handler: modbus.NewRTUClientHandler(c.MODBUS.Port),
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

func (m *MODBUS) ReadFloat32(p *config.Probe) (float32, float32, error) {
	m.c.Mutex.Lock()
	defer m.c.Mutex.Unlock()

	h := m.handler
	h.SlaveId = p.Id
	client := modbus.NewClient(h)
	var err error
	value := float32(0.0)
	temperature := float32(0.0)
	for i := 0; i < p.Retry; i += 1 {
		res, er := client.ReadHoldingRegisters(p.Value_reg, 2)
		err = er
		if er != nil {
			log.Print("error reading value ", p.Name, " error: ", err)
			time.Sleep(time.Millisecond * time.Duration(p.Retry_Delay))
			continue
		}
		tmp01 := binary.LittleEndian.Uint32(res)
		value = math.Float32frombits(tmp01)
		log.Print("Succes reading ", p.Name, " value(hex): ", hex.EncodeToString(res), " value_raw: ", value, " value_calc: ", p.GetValue_calc())
		break
	}

	if p.Temp_Reg != 0 {
		for i := 0; i < p.Retry; i += 1 {
			res, er := client.ReadHoldingRegisters(p.Temp_Reg, 2)
			err = er
			if er != nil {
				log.Print("error reading temperature ", p.Name, " error: ", err)
				time.Sleep(time.Millisecond * time.Duration(p.Retry_Delay))
				continue
			}
			tmp01 := binary.LittleEndian.Uint32(res)
			temperature = math.Float32frombits(tmp01)
			log.Print("Succes reading ", p.Name, " temp(hex): ", hex.EncodeToString(res), " temp: ", temperature)
			break
		}
	}
	time.Sleep(time.Duration(p.Retry_Delay) * time.Millisecond)
	return value, temperature, err
}

func (m *MODBUS) ReadFlow(p *config.Probe) (float32, uint32, error) {
	m.c.Mutex.Lock()
	defer m.c.Mutex.Unlock()

	h := m.handler
	h.SlaveId = p.Id
	client := modbus.NewClient(h)
	var err error
	value := float32(0.0)
	total := uint32(0)
	for i := 0; i < p.Retry; i += 1 {
		res, er := client.ReadHoldingRegisters(p.Value_reg, 4)
		err = er
		if er != nil {
			log.Print("error reading flow and total: ", p.Name, " error: ", err)
			time.Sleep(time.Duration(p.Retry_Delay) * time.Millisecond)
			continue
		}
		tmp01 := binary.LittleEndian.Uint32(res[0:4])
		value = math.Float32frombits(tmp01)
		total = binary.LittleEndian.Uint32(res[4:8])
		log.Print("Succes reading ", p.Name, " flow (hex): ", hex.EncodeToString(res[0:4]), " flow: ", value, " total (hex): ", hex.EncodeToString(res[4:7]), " total: ", total)
		break
	}
	time.Sleep(time.Duration(p.Retry_Delay) * time.Millisecond)
	return value, total, err
}

func (m *MODBUS) ReadKAB(p *config.Probe) (float32, float32, error) {
	m.c.Mutex.Lock()
	defer m.c.Mutex.Unlock()

	h := m.handler
	h.SlaveId = p.Id
	client := modbus.NewClient(h)
	var err error
	ka := float32(0)
	kb := float32(0)
	for i := 0; i < p.Retry; i += 1 {
		res, er := client.ReadHoldingRegisters(p.Kab_Reg, 4)
		err = er
		if err != nil {
			log.Print("error reading KAB ", p.Name, ": ", err)
			time.Sleep(time.Duration(p.Retry_Delay) * time.Millisecond)
			continue
		}
		tmp01 := binary.LittleEndian.Uint32(res[0:1])
		ka = math.Float32frombits(tmp01)
		tmp01 = binary.LittleEndian.Uint32(res[2:3])
		kb = math.Float32frombits(tmp01)
		log.Print("Success reading KA, KB ", p.Name, " ka: ", res[0:1], " kb: ", res[1:2])
		break
	}
	time.Sleep(time.Duration(p.Retry_Delay) * time.Millisecond)
	return ka, kb, err
}

func (m *MODBUS) WriteKB(p *config.Probe, ka float32, kb float32) error {
	m.c.Mutex.Lock()
	defer m.c.Mutex.Unlock()

	h := m.handler
	h.SlaveId = p.Id
	client := modbus.NewClient(h)
	var01 := make([]byte, 4)
	binary.LittleEndian.PutUint32(var01[0:1], math.Float32bits(ka))
	binary.LittleEndian.PutUint32(var01[2:3], math.Float32bits(kb))
	var err error
	var res []byte
	for i := 0; i < p.Retry; i += 1 {
		r, er := client.WriteMultipleRegisters(p.Kab_Reg, 4, var01)
		err = er
		if err != nil {
			log.Print("error writing ka, kb ", p.Name, ": ", err)
			time.Sleep(time.Millisecond * time.Duration(p.Retry_Delay))
			continue
		}
		res = r
		log.Print("sucsess writing ka, kb", p.Name, " ka: ", res[0:1], " kb: ", res[2:3])
		break
	}
	return err
}

func (m *MODBUS) Loop() {
	c := m.c
	log.Print("entering loop with duration: ", c.MODBUS.Sleep, " ms.")
	go func() {
		p := []*config.Probe{&c.PH, &c.COD, &c.TSS, &c.NH3N}
		for {
			start := time.Now()
			for _, e := range p {
				c.Mutex.Lock()
				enable := e.Enable
				c.Mutex.Unlock()
				if enable {
					log.Print("reading probe: ", e.Name)
					val, tempe, err := m.ReadFloat32(e)
					c.Mutex.Lock()
					if err != nil {
						e.Error = true
						log.Print("error reading probe slave, ", e.Name, ": ", err)
					} else {
						e.Error = false
						e.Value_raw = val
						e.Temp = tempe
					}
					c.Mutex.Unlock()
				} else {
					log.Print("reading probe disabled: ", e.Name)
					continue
				}
			}

			// reading flow meter should goes here
			c.Mutex.Lock()
			enable := c.FLOW.Enable
			c.Mutex.Unlock()
			if enable {
				log.Print("Reading probe: ", c.FLOW.Name)
				val, tot, err := m.ReadFlow(&c.FLOW)
				c.Mutex.Lock()
				if err != nil {
					c.FLOW.Error = true
					log.Print("error reading probe ", c.FLOW.Name, ": ", err)
				} else {
					c.FLOW.Error = false
					c.FLOW.Flow = val
					c.FLOW.Total = tot
				}
				c.Mutex.Unlock()
			} else {
				log.Print("reading probe disabled: ", c.FLOW.Name)
			}

			duration := time.Since(start)
			var loop_delay int64
			log.Print("probe reading duration: ", duration.Milliseconds(), " loop delay: ", (time.Millisecond * time.Duration(time.Duration(c.MODBUS.Sleep))).Milliseconds())
			if duration.Milliseconds() < (time.Millisecond.Milliseconds() * int64(c.MODBUS.Sleep)) {
				loop_delay = time.Millisecond.Milliseconds()*int64(c.MODBUS.Sleep) - duration.Milliseconds()
			}
			log.Print("reading probe sleep for: ", loop_delay)
			time.Sleep(time.Millisecond * time.Duration(loop_delay))
		}
	}()
}
