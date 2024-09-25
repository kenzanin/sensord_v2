package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	config "sensord_v2/config"
	"strconv"
)

type server struct {
	c *config.Config
}

func ServerInit(c *config.Config) *server {
	return &server{
		c: c,
	}
}

func (s *server) probe_write() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var probe *config.Probe
		var req string
		path := r.URL.Path
		log.Print("URL path: ", path, "?", r.URL.RawQuery)
		switch path {
		case "/write/PH":
			probe = &s.c.PH
		case "/write/COD":
			probe = &s.c.COD
		case "/write/NH3N":
			probe = &s.c.NH3N
		case "/write/TSS":
			probe = &s.c.TSS
		}

		req = r.URL.Query().Get("write_ka")
		if len(req) > 0 {
			ka, err := strconv.ParseFloat(req, 32)
			if err != nil {
				res := []byte(fmt.Sprintf("{\"%s\":%t, \"error\": \"write_ka %s\" }", probe.Name, probe.Enable, err))
				w.Write(res)
			}
			res := []byte(fmt.Sprintf("{\"%s\":%t, \"write_ka\": \"%f\" }", probe.Name, probe.Enable, ka))
			w.Write([]byte(res))
		}

		req = r.URL.Query().Get("write_kb")
		if len(req) > 0 {
			kb, err := strconv.ParseFloat(req, 32)
			if err != nil {
				res := []byte(fmt.Sprintf("{\"%s\":%t, \"error\": \"write_kb %s\" }", probe.Name, probe.Enable, err))
				w.Write(res)
			}
			res := []byte(fmt.Sprintf("{\"%s\":%t, \"write_kb\": \"%f\" }", probe.Name, probe.Enable, kb))
			w.Write([]byte(res))
		}
	}
}

func (s *server) probe_read() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var probe *config.Probe
		path := r.URL.Path
		log.Print("URL path: ", path, "?", r.URL.RawQuery)
		switch path {
		case "/read/PH":
			probe = &s.c.PH
		case "/read/COD":
			probe = &s.c.COD
		case "/read/NH3N":
			probe = &s.c.NH3N
		case "/read/TSS":
			probe = &s.c.TSS
		}

		req := r.URL.Query().Get("read")
		if len(req) > 0 {
			if req == "true" {
				if !probe.Enable {
					probe.Enable = true
				}
				w.Write([]byte(fmt.Sprintf("{\"%s\":\"%t\", \"value\": %f, \"temperature\": %f}", probe.Name, probe.Enable, probe.Value, probe.Temp)))
			} else if req == "false" {
				probe.Enable = false
				w.Write([]byte(fmt.Sprintf("{\"%s\":\"%t\"}", probe.Name, probe.Enable)))
			}
			log.Print(probe.Name, " Status: ", probe.Enable)
			return
		}
	}
}

func (s *server) Serve() {
	http.HandleFunc("/read/PH", s.probe_read())
	http.HandleFunc("/read/COD", s.probe_read())
	http.HandleFunc("/read/NH3N", s.probe_read())
	http.HandleFunc("/read/TSS", s.probe_read())
	http.HandleFunc("/read/FLOW", s.probe_read())

	http.HandleFunc("/write/PH", s.probe_write())
	http.HandleFunc("/write/COD", s.probe_write())
	http.HandleFunc("/write/NH3N", s.probe_write())
	http.HandleFunc("/write/TSS", s.probe_write())
	http.HandleFunc("/write/FLOW", s.probe_write())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.c)
	})
	http.ListenAndServe(s.c.SERVER.Addr+":"+s.c.SERVER.Port, nil)
}
