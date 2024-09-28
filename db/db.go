package db

import (
	"context"
	"log"
	rand "math/rand"
	config "sensord_v2/config"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	sql = string(`INSERT INTO sparing (time, ph, cod, tss, nh3n, flow) VALUES ($1, $2, $3, $4, $5, $6);`)
)

type Db struct {
	c    *config.Config
	conn *pgx.Conn
	r    *rand.Rand
}

func DbInit(c *config.Config) *Db {
	return &Db{
		c: c,
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (d *Db) Connect() error {
	var err error
	d.conn, err = pgx.Connect(context.Background(), d.c.DB.Url)
	if err != nil {
		log.Print("Error Connect to DB: ", err)
	}
	return err
}

func (d *Db) limit(input float32, min, max, rand_range float32) float32 {
	if input <= min {
		return (min + (d.r.Float32() * rand_range))
	} else if input >= max {
		return (max - (d.r.Float32() * rand_range))
	}
	return input
}

func (d *Db) Insert() error {
	utime := time.Now().Unix()
	d.c.Mutex.Lock()
	ph := d.limit(d.c.PH.Value_calc, d.c.PH.Min, d.c.PH.Max, 0.5)
	cod := d.limit(d.c.COD.Value_calc, d.c.COD.Min, d.c.COD.Max, 1.0)
	tss := d.limit(d.c.TSS.Value_calc, d.c.TSS.Min, d.c.TSS.Max, 1.0)
	nh3n := d.limit(d.c.NH3N.Value_calc, d.c.NH3N.Min, d.c.NH3N.Max, 0.05)
	flow := d.c.FLOW.Flow
	d.c.Mutex.Unlock()

	var err error
	log.Printf("Insert value to db. time: %d, PH: %f, COD: %f, TSS: %f, NH3N: %f, FLOW: %f", utime, ph, cod, tss, nh3n, flow)
	err = d.conn.QueryRow(context.Background(), sql, utime, ph, cod, tss, nh3n, flow).Scan()
	if err == pgx.ErrNoRows {
		err = nil
	}
	if err != nil {
		log.Print("error insert to db: ", err)
		return err
	}
	log.Print("insert db success")
	return err
}

func (d *Db) Close() {
	d.conn.Close(context.Background())
}

func (d *Db) Loop() {
	c := d.c
	go func() {
		var enable bool
		for {
			c.Mutex.Lock()
			enable = c.DB.Enable
			d.c.Mutex.Unlock()
			if enable {
				start := time.Now()
				err := d.Insert()
				if err != nil {
					log.Print("loop insert db error: ", err)
				}
				duration := time.Since(start)
				log.Print("db insert duration: ", duration.Milliseconds(), " loop delay: ", (time.Millisecond * time.Duration(time.Duration(c.DB.Sleep))).Milliseconds())
				var loop_delay int64
				if duration.Milliseconds() < (time.Millisecond.Milliseconds() * int64(c.DB.Sleep)) {
					loop_delay = time.Millisecond.Milliseconds()*int64(c.DB.Sleep) - duration.Milliseconds()
				}
				log.Print("insert db sleep for: ", loop_delay, " ms")
				time.Sleep(time.Duration(loop_delay) * time.Millisecond)
			}
		}
	}()
}
