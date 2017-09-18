package main

/* NOTE:
A,ALIAS,NS and CNAME are only supported by REST API. for the rest(MX,TXT etc.) you can add tinydata in
raw format directly
*/

import (
	"fmt"
	log "github.com/briankohler/log"
	"github.com/gorilla/mux"
	"net/http"
)

type RAW_Record struct {
	name    string
	tinystr string
}

func (c *RAW_Record) GenConsulRecord() Consul_Record {
	cr := Consul_Record{}
	cr.key = fmt.Sprintf("%v/RAW", consul_keyspace)
	cr.value = c.name
	cr.record = c.name
	cr.rtype = "RAW"
	cr.tiny = c.tinystr
	return cr
}

func (c *RAW_Record) Check() error {
	// Formats that are handled, you have to go through the API
	tinystrb := []byte(c.tinystr)
	rt := string(tinystrb[0])
	fmt.Printf("First byte: %v\n", rt)
	var handled = make(map[string]int)
	handled["."] = 1
	handled["+"] = 1
	handled["="] = 1
	handled["C"] = 1
	handled["%"] = 1
	fmt.Printf("First byte: %v MAP=%v\n", rt, handled[rt])

	if handled[rt] == 1 {
		err := fmt.Errorf("Use API: TYPE already handled via API for %v", string(c.tinystr))
		return err
	}
	// Error for unknown
	handled["^"] = 1
	handled["'"] = 1
	handled["@"] = 1
	handled["#"] = 1
	handled["&"] = 1
	if handled[rt] != 1 {
		err := fmt.Errorf("Unknown TYPE %v", string(c.tinystr))
		return err
	}

	// Verify if its one of the unhandled formats and blindly allow
	return nil
}

func HandleRAW(w http.ResponseWriter, r *http.Request) {
	log.Debug("LOG: ", r.Method, r.URL)
	vars := mux.Vars(r)
	p := RAW_Record{name: vars["key"], tinystr: vars["tinystr"]}
	var err error
	switch r.Method {
	case "POST":
		err = AddRecord(&p)
	case "DELETE":
		err = DeleteRecord(&p)
	case "PUT":
		nvals := RAW_Record{name: vars["key"], tinystr: vars["newtinystr"]}
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
