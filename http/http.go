package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	config "sensord_v2/config"
	"strconv"
	"time"
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

		req = r.URL.Query().Get("write_offa")
		if len(req) > 0 {
			ka, err := strconv.ParseFloat(req, 32)
			if err != nil {
				res := []byte(fmt.Sprintf("{\"%s\":%t, \"error\": \"write_kb %s\" }", probe.Name, probe.Enable, err))
				w.Write(res)
			}
			s.c.Mutex.Lock()
			probe.Offset_a = float32(ka)
			s.c.Mutex.Unlock()
			res := []byte(fmt.Sprintf("{\"%s\":%t, \"write_offa\": \"%f\" }", probe.Name, probe.Enable, ka))
			w.Write([]byte(res))
		}

		req = r.URL.Query().Get("write_offb")
		if len(req) > 0 {
			kb, err := strconv.ParseFloat(req, 32)
			if err != nil {
				res := []byte(fmt.Sprintf("{\"%s\":%t, \"error\": \"write_kb %s\" }", probe.Name, probe.Enable, err))
				w.Write(res)
			}
			s.c.Mutex.Lock()
			probe.Offset_b = float32(kb)
			s.c.Mutex.Unlock()
			res := []byte(fmt.Sprintf("{\"%s\":%t, \"write_offb\": \"%f\" }", probe.Name, probe.Enable, kb))
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
		case "/read/FLOW":
			probe = &s.c.FLOW
		}

		req := r.URL.Query().Get("read")
		if len(req) > 0 {
			if req == "true" {
				if !probe.Enable {
					s.c.Mutex.Lock()
					probe.Enable = true
					s.c.Mutex.Unlock()
				}
				if probe.Name == "FLOW" {
					w.Write([]byte(fmt.Sprintf("{\"%s\":\"%t\", \"value_raw\": %f, \"total:\",%d}", probe.Name, probe.Enable, probe.Flow, probe.Total)))
				} else {
					w.Write([]byte(fmt.Sprintf("{\"%s\":\"%t\", \"value_raw\": %f, \"value_calc\": %f,\"temperature\": %f}", probe.Name, probe.Enable, probe.Value_raw, probe.GetValue_calc(), probe.Temp)))
				}
			} else if req == "false" {
				s.c.Mutex.Lock()
				probe.Enable = false
				s.c.Mutex.Unlock()
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
	http.HandleFunc("/read/PROBES", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log.Print("Url Path: ", r.URL.Path)
		type data struct {
			Time      uint64
			Ph_raw    float32
			Ph_calc   float32
			Temp      float32
			Cod_raw   float32
			Cod_calc  float32
			Tss_raw   float32
			Tss_calc  float32
			Nh3n_raw  float32
			Nh3n_calc float32
			Flow      float32
			Total     uint32
		}
		s.c.Mutex.Lock()
		dat := data{
			Time:      uint64(time.Now().Unix()),
			Ph_raw:    s.c.PH.Value_raw,
			Ph_calc:   s.c.PH.GetValue_calc(),
			Cod_raw:   s.c.COD.Value_raw,
			Cod_calc:  s.c.COD.GetValue_calc(),
			Tss_raw:   s.c.TSS.Value_raw,
			Tss_calc:  s.c.TSS.GetValue_calc(),
			Nh3n_raw:  s.c.NH3N.Value_raw,
			Nh3n_calc: s.c.NH3N.GetValue_calc(),
			Flow:      s.c.FLOW.Flow,
			Total:     s.c.FLOW.Total,
			Temp:      s.c.PH.Temp,
		}
		s.c.Mutex.Unlock()
		log.Printf("%#v", dat)
		json.NewEncoder(w).Encode(dat)
	})

	http.HandleFunc("/write/PH", s.probe_write())
	http.HandleFunc("/write/COD", s.probe_write())
	http.HandleFunc("/write/NH3N", s.probe_write())
	http.HandleFunc("/write/TSS", s.probe_write())
	http.HandleFunc("/write/FLOW", s.probe_write())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.c)
	})
	http.HandleFunc("/save", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		s.c.Save()
		json.NewEncoder(w).Encode(s.c)
	})
	http.ListenAndServe(s.c.SERVER.Addr+":"+s.c.SERVER.Port, nil)
}
