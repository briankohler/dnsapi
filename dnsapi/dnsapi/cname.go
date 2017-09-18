package main

import (
	"errors"
	"fmt"
	"github.com/briankohler/helper"
	log "github.com/briankohler/log"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type CNAME_Record struct {
	name  string
	value string
	ttl   string
	lo    string
}

func (c *CNAME_Record) GenConsulRecord() Consul_Record {
	cr := Consul_Record{}
	if c.lo == "" {
		c.lo = "GLOBAL"
	}
	cr.key = fmt.Sprintf("%v/CNAME/%v/%v", consul_keyspace, c.lo, c.name)
	cr.value = c.value
	cr.record = c.name
	cr.rtype = "CNAME"
	cr.tiny = fmt.Sprintf("C%v:%v:%v", c.name, c.value, c.ttl)
	if c.lo != "GLOBAL" {
		cr.tiny = fmt.Sprintf("%v::%v", cr.tiny, c.lo)
	}
	return cr
}

func (c *CNAME_Record) Check() error {
	//check if c.name belongs to right domain
	ret, err := helper.CheckFQDN([]byte(c.name))
	if err == nil && ret == true {
		//TODO check if c.value is OK
		if _, err = strconv.Atoi(c.ttl); err == nil {
			if len(c.value) < 3 {
				return errors.New("INVALID: CNAME value")
			}
			if c.lo != "" {
				err := VerifyValidLOC(c.lo)
				if err != nil {
					return err
				}
			}
			return nil
		} else {
			return errors.New("TTL: should be number")
		}
	} else {
		log.Error(fmt.Sprintf("Error = %v and Return = %v\n", err, ret))
		return err
	}
	return err

}

func HandleCNAME(w http.ResponseWriter, r *http.Request) {
	log.Debug("LOG: ", r.Method, r.URL)
	if err := VerifyURL("CNAME", r.Method, string(r.URL.String())); err != nil {
		http.Error(w, err.Error(), 404)
		return
	}

	//apiRouter.HandleFunc("/CNAME/{key}/{value}/{ttl}", HandleCNAME) //Handle both POST & DELETE here
	//apiRouter.HandleFunc("/CNAME/{key}/{value}/{ttl}/{nvalue}/{nttl}", HandleCNAME)

	vars := mux.Vars(r)
	lo := r.FormValue("loc")
	p := CNAME_Record{name: vars["key"], value: vars["value"], ttl: vars["ttl"], lo: lo}
	var err error
	switch r.Method {
	case "POST":
		err = AddRecord(&p)
	case "DELETE":
		err = DeleteRecord(&p)
	case "PUT":
		nvals := CNAME_Record{name: vars["key"], value: vars["nvalue"], ttl: vars["nttl"], lo: lo}
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
