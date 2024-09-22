package Loop

import (
	"log"
	config "sensord_v2/config"
	db "sensord_v2/db"
	modbus "sensord_v2/modbus"
	"time"
)

type LOOP struct {
}

func LOOPInit() *LOOP {
	return &LOOP{}
}

func (l *LOOP) Loop(c *config.Config, m *modbus.MODBUS, d *db.Db) {
	log.Print("entering loop with duration: ", c.LOOP_DELAY, " ms.")
	go func() {
		p := []*config.Probe{&c.PH, &c.COD, &c.TSS, &c.NH3N}
		for {
			for _, e := range p {
				enable := e.ENABLE
				if enable {
					log.Print("reading probe: ", e.NAME)
					val, tempe, err := m.ReadFloat32(e)
					if err != nil {
						c.Mutex.Lock()
						e.ERROR = true
						c.Mutex.Unlock()
						log.Print("error reading probe slave, ", e.NAME, ": ", err)
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
					log.Print("reading probe disabled: ", e.NAME)
					continue
				}
			}
			log.Print("reading probe sleep for: ", c.LOOP_DELAY, " ms.")
			time.Sleep(time.Millisecond * time.Duration(c.LOOP_DELAY))
		}
	}()
	go func() {
		var enable bool
		for {
			enable = c.DB.Enable
			if enable {
				err := d.Insert()
				if err != nil {
					log.Print("loop insert db error: ", err)
				}
				log.Print("insert db sleep for: ", c.DB.Loop, " ms")
				time.Sleep(time.Millisecond * time.Duration(c.DB.Loop))
			}
		}
	}()
}
