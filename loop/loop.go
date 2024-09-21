package Loop

import (
	"log"
	config "sensord_v2/config"
	modbus "sensord_v2/modbus"
	"time"
)

type LOOP struct {
}

func LOOPInit() *LOOP {
	return &LOOP{}
}

func (l *LOOP) Loop(c *config.Config, m *modbus.MODBUS) {
	log.Print("entering loop with duration: ", c.LOOP_DELAY, " ms.")
	go func() {
		p := []*config.Probe{&c.PH, &c.COD, &c.TSS, &c.NH3N}
		for {
			for _, e := range p {
				c.Mutex.Lock()
				enable := e.ENABLE
				c.Mutex.Unlock()
				if enable {
					log.Print("reading ", e.NAME)
					val, tempe, err := m.ReadFloat32(e)
					if err != nil {
						c.Mutex.Lock()
						e.ERROR = true
						c.Mutex.Unlock()
						log.Print("error reading slave, ", e.NAME, ": ", err)
					} else {
						c.Mutex.Lock()
						e.ERROR = false
						c.Mutex.Unlock()
					}

					c.Mutex.Lock()
					e.VALUE = val
					e.TEMP = tempe
					c.Mutex.Unlock()

				} else {
					log.Print("reading disabled: ", e.NAME)
					continue
				}
			}
			log.Print("sleep for: ", c.LOOP_DELAY, " ms.")
			time.Sleep(time.Millisecond * time.Duration(c.LOOP_DELAY))
		}
	}()
}
