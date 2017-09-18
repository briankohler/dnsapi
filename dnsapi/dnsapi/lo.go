package main

/*

L Record is currently only supported for CNAME & ALIAS RECORDs

*/

import (
	"errors"
	"fmt"
	"github.com/briankohler/consulhelper"
	log "github.com/briankohler/log"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"strings"
)

type LOC_Record struct {
	name  string
	value string

	//redundant, but just keeping it
	ttl string
	lo  string
}

func VerifyValidLOC(s string) error {
	key := fmt.Sprintf("%v/LOC", consul_keyspace)
	dat, err := consulhelper.GetValueInConsul(consul, key)
	if err == nil {
		if dat[s] == "" {
			return errors.New("LOC undefined. Need to be defined first")
		}
	}
	return err
}
func (c *LOC_Record) GenConsulRecord() Consul_Record {
	cr := Consul_Record{}
	cr.key = fmt.Sprintf("%v/LOC", consul_keyspace)
	cr.value = c.name
	cr.record = c.name
	cr.rtype = "LOC"
	cr.tiny = fmt.Sprintf("%%%v:%v", c.name, c.value)
	return cr
}

func verify_lo_net(s string) error {
	iparr := strings.Split(s, ".")
	fl := len(iparr)
	if fl > 4 {
		return errors.New("INVALID: IP net range")
	}
	if fl < 1 {
		return errors.New("NOT DEFINED: IP net range")
	}
	for a := range iparr {
		ipb, err := strconv.Atoi(iparr[a])
		if err == nil {
			if !(ipb < 255 && ipb > -1) {
				return errors.New("INVALID: IP net range")
			}
		} else {
			e := fmt.Errorf("%v: Error converting to Int: %v", iparr[a], err)
			return e
		}
	}
	return nil
}

func (c *LOC_Record) Check() error {
	if ip_err := verify_lo_net(c.value); ip_err != nil {
		return ip_err
	}
	//	if len(c.name) > 2 {
	//		return errors.New("lo can be 1 or 2 ascii")
	//	}
	if c.name == "GLOBAL" {
		return errors.New("GLOBAL is a reserved location")
	}
	return nil

}

func HandleLOC(w http.ResponseWriter, r *http.Request) {
	log.Debug("LOG: ", r.Method, r.URL)
	vars := mux.Vars(r)
	p := LOC_Record{name: vars["key"], value: vars["value"]}
	var err error
	switch r.Method {
	case "POST":
		err = AddRecord(&p)
	case "DELETE":
		err = DeleteRecord(&p)
	case "PUT":
		nvals := LOC_Record{name: vars["key"], value: vars["nvalue"]}
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
