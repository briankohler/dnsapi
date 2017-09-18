package main

import (
	"errors"
	"fmt"
	//"github.com/briankohler/helper"
	log "github.com/briankohler/log"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"strings"
)

type NS_Record struct {
	name   string
	value  string
	ttl    string
	nsname string
	lo     string
}

func (c *NS_Record) GenConsulRecord() Consul_Record {
	cr := Consul_Record{}
	if c.lo == "" {
		c.lo = "GLOBAL"
	}
	cr.key = fmt.Sprintf("%v/NS/%s/%v", consul_keyspace, c.lo, c.nsname)
	cr.value = fmt.Sprintf("127.0.0.1/%v", c.name)
	cr.record = c.name
	cr.rtype = "NS"
	cr.tiny = fmt.Sprintf("&%v:127.0.0.1:%v", c.name, c.nsname)
	if c.lo != "GLOBAL" {
		cr.tiny = fmt.Sprintf("::%v", c.lo)
	}
	return cr
}

func (c *NS_Record) Check() error {
	//check if c.name belongs to right domain
	//ret, err := helper.CheckFQDN([]byte(c.nsname))
        ret := true
        // err := nil
	log.Debug("GOT NAME: ", c.name, " TTL: ", c.ttl)
	if ret == true {
		fqdn_str := string(c.nsname)
		fqdn_arr := strings.Split(fqdn_str, ".")
		fl := len(fqdn_arr)
		var nsret bool
		var nserr error
		if fl == 1 { //just search for tinydns data format
			nsret = true
			nserr = nil
		} // else {
		//	nsret, nserr = helper.CheckFQDN([]byte(c.nsname))
		// }
		if nserr == nil && nsret == true {
			ip := net.ParseIP("127.0.0.1")
			if ip == nil {
				return errors.New("INVALID: IP Address")
			}
		} else {
			log.Error("(NSVAL)Error = ", nserr, " and Return = ", nsret)
			return nserr
		}
	}// else {
	//	log.Error(fmt.Sprintf("(NS)Error = %v and Return = %v\n", err, ret))
	//	return err
	//}
	return nil
}

func HandleNS(w http.ResponseWriter, r *http.Request) {
	log.Debug("LOG: ", r.Method, r.URL)
	vars := mux.Vars(r)
	lo := r.FormValue("loc")
	var p NS_Record

	if (vars["ip"] == "") && (vars["ttl"] == "") && (vars["nsname"] == "") { // Handle NSLO
		p = NS_Record{name: vars["key"], value: "127.0.0.1", nsname: vars["name"]}
	} else {
		p = NS_Record{name: vars["key"], value: vars["ip"], lo: lo, nsname: vars["name"]}
	}
	var err error
	switch r.Method {
	case "POST":
		err = AddRecord(&p)
	case "DELETE":
		err = DeleteRecord(&p)
	case "PUT":
		nvals := NS_Record{name: vars["key"], value: vars["nip"], ttl: vars["nttl"], nsname: vars["nname"], lo: lo}
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
