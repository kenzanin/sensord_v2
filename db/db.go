package db

import (
	"context"
	"log"
	config "sensord_v2/config"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	sql = `INSERT INTO sparingdb (time, ph, cod, tss, nh3n) VALUES ($1, $2, $3, $4 $5)`
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
	defer d.conn.Close(context.Background())
	return err
}

func (d *Db) Insert() error {
	var err error
	utime := time.Now().Unix()
	log.Printf("Insert value to db. time: %d, PH: %f, COD: %f, TSS: %f, NH3N: %f", utime, d.c.PH.VALUE, d.c.COD.VALUE, d.c.TSS.VALUE, d.c.NH3N.VALUE)
	err = d.conn.QueryRow(context.Background(), sql, utime, d.c.PH.VALUE, d.c.COD.VALUE, d.c.TSS.VALUE, d.c.NH3N.VALUE).Scan()
	if err != nil {
		log.Print("error insert to db: ", err)
		return err
	}
	log.Print("insert db success")
	return err
}
