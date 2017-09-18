package main

/* NOTE:

The A Record also automatically generates a PTR record. If you want an A
record in Round Robin mode, you should be using ALIAS instead of A

A record should only be used where you need a PTR record too

*/

import (
	"errors"
	"fmt"
	"github.com/briankohler/helper"
	log "github.com/briankohler/log"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"strconv"
)

type A_Record struct {
	name  string
	value string
	ttl   string
	lo    string
}

func (c *A_Record) GenConsulRecord() Consul_Record {
	cr := Consul_Record{}
	if c.lo == "" {
		c.lo = "GLOBAL"
	}
	cr.key = fmt.Sprintf("%v/A/%v/%v", consul_keyspace, c.lo, c.name)
	fmt.Println(cr.key)
	cr.value = c.value
	cr.record = c.name
	cr.rtype = "A"
	cr.tiny = fmt.Sprintf("=%v:%v:%v", c.name, c.value, c.ttl)
	if c.lo != "GLOBAL" {
		cr.tiny = fmt.Sprintf("%v::%v", cr.tiny, c.lo)
	}
	return cr
}

func (c *A_Record) Check() error {
	//check if c.name belongs to right domain
	ret, err := helper.CheckFQDN([]byte(c.name))
	if err == nil && ret == true {
		ip := net.ParseIP(c.value)
		if ip == nil {
			return errors.New("INVALID: IP Address")
		}
		if _, err = strconv.Atoi(c.ttl); err != nil {
			return errors.New("TTL: should be number")
		}
		if c.lo != "" {
			err := VerifyValidLOC(c.lo)
			if err != nil {
				return err
			}
		}
	} else {
		log.Error("Error = ", err, " and Return = ", ret)
		return err
	}
	return nil
}

func HandleA(w http.ResponseWriter, r *http.Request) {
	log.Debug("LOG: ", r.Method, r.URL)
	vars := mux.Vars(r)
	lo := r.FormValue("loc")
	p := A_Record{name: vars["key"], value: vars["value"], ttl: vars["ttl"], lo: lo}
	var err error
	switch r.Method {
	case "POST":
		err = AddRecord(&p)
	case "DELETE":
		err = DeleteRecord(&p)
	case "PUT":
		nvals := A_Record{name: vars["key"], value: vars["nvalue"], ttl: vars["nttl"], lo: lo}
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
