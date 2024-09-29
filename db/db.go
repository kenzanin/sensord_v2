package db

import (
	"context"
	"log"
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
}

func DbInit(c *config.Config) *Db {
	return &Db{
		c: c,
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

func (d *Db) Insert() error {
	utime := time.Now().Unix()
	d.c.Mutex.Lock()
	ph := d.c.PH.GetValue_calc()
	cod := d.c.COD.GetValue_calc()
	tss := d.c.TSS.GetValue_calc()
	nh3n := d.c.NH3N.GetValue_calc()
	flow := d.c.FLOW.Flow
	d.c.Mutex.Unlock()

	var err error
	log.Printf("Insert value to db. time: %d, PH: %.2f, COD: %.2f, TSS: %.2f, NH3N: %.2f, FLOW: %.2f", utime, ph, cod, tss, nh3n, flow)
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
		time.Sleep(time.Second * 30)
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
