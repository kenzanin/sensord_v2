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
	d.conn, err = pgx.Connect(context.Background(), d.c.DB.Db_Url)
	if err != nil {
		log.Print("Error Connect to DB: ", err)
	}
	return err
}

func (d *Db) Insert() error {
	utime := time.Now().Unix()
	ph := d.c.PH.VALUE
	cod := d.c.COD.VALUE
	tss := d.c.TSS.VALUE
	nh3n := d.c.NH3N.VALUE
	flow := 1.0

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
