package main

import (
	"errors"
	"fmt"
	"github.com/briankohler/helper"
	log "github.com/briankohler/log"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"strings"
)

type SOA_Record struct {
	name string
	value string
	nsname string
        lo string
}

func (c *SOA_Record) GenConsulRecord() Consul_Record {
	cr := Consul_Record{}
	if c.lo == "" {
		c.lo = "GLOBAL"
	}
	cr.key = fmt.Sprintf("%v/SOA/%v/%v", consul_keyspace, c.lo, c.nsname)
	cr.value = fmt.Sprintf("127.0.0.1/%v", c.name)
	cr.record = c.nsname
	cr.rtype = "SOA"
	cr.tiny = fmt.Sprintf("Z%v:127.0.0.1:%v", c.name, c.nsname)
	if c.lo != "GLOBAL" {
		cr.tiny = fmt.Sprintf("::%v", c.lo)
	}
	return cr
}

func (c *SOA_Record) Check() error {
	//check if c.name belongs to right domain
	ret, err := helper.CheckFQDN([]byte(c.name))
	log.Debug("GOT NAME:", c.name, " VALUE:", c.value)
	if err == nil && ret == true {
		fqdn_str := string(c.nsname)
		fqdn_arr := strings.Split(fqdn_str, ".")
		fl := len(fqdn_arr)
		var nsret bool
		var nserr error
		if fl == 1 { //just search for tinydns data format
			nsret = true
			nserr = nil
		} else {
			nsret, nserr = helper.CheckFQDN([]byte(c.nsname))
		}
		if nserr == nil && nsret == true {
			ip := net.ParseIP("127.0.0.1")
			if ip == nil {
				return errors.New("INVALID: IP Address")
			}
		} else {
			log.Error("(SOAVAL)Error = ", fmt.Sprintf("%s", nserr), " and Return = ", nsret)
			return nserr
		}
	} else {
		log.Error("(SOA)Error = ", fmt.Sprintf("%s", err), " and Return = ", ret)
		return err
	}
	return nil
}

func HandleSOA(w http.ResponseWriter, r *http.Request) {
	log.Debug("LOG: ", r.Method, r.URL)
	vars := mux.Vars(r)
	lo := r.FormValue("loc")
	var p SOA_Record
        p = SOA_Record{name: vars["key"], value: "127.0.0.1", nsname: vars["name"]}
	var err error
	switch r.Method {
	case "POST":
		err = AddRecord(&p)
	case "DELETE":
		err = DeleteRecord(&p)
	case "PUT":
		nvals := SOA_Record{name: vars["key"], value: "127.0.0.1",  nsname: vars["nnsame"], lo: lo}
		err = EditRecord(&p, &nvals)
	default:
		err = fmt.Errorf("Unrecognized method: %v", r.Method)

	}
	if err == nil {
		w.Write([]byte("OK"))
	} else {
		http.Error(w, err.Error(), 404)
	}
}
