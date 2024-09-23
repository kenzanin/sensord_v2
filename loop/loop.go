package Loop

import (
	"log"
	config "sensord_v2/config"
	db "sensord_v2/db"
	modbus "sensord_v2/modbus"
	"time"
)

type LOOP struct {
	loop_reader uint32
	loop_db     uint32
}

func LOOPInit(c *config.Config) *LOOP {
	return &LOOP{
		loop_reader: c.LOOP.LOOP_READER,
		loop_db:     c.LOOP.LOOP_DB,
	}
}

func (l *LOOP) Loop(c *config.Config, m *modbus.MODBUS, d *db.Db) {
	log.Print("entering loop with duration: ", l.loop_reader, " ms.")
	go func() {
		p := []*config.Probe{&c.PH, &c.COD, &c.TSS, &c.NH3N}
		for {
			start := time.Now()
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
			duration := time.Since(start)
			var loop_delay int64
			log.Print("probe reading duration: ", duration.Milliseconds(), " loop delay: ", (time.Millisecond * time.Duration(time.Duration(l.loop_reader))).Milliseconds())
			if duration.Milliseconds() < (time.Millisecond.Milliseconds() * int64(l.loop_reader)) {
				loop_delay = time.Millisecond.Milliseconds()*int64(l.loop_reader) - duration.Milliseconds()
			}
			log.Print("reading probe sleep for: ", loop_delay)
			time.Sleep(time.Millisecond * time.Duration(loop_delay))
		}
	}()
	go func() {
		var enable bool
		for {
			c.Mutex.Lock()
			enable = c.DB.Enable
			c.Mutex.Unlock()
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
