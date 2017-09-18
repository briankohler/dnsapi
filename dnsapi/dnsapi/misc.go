package main

import (
	"encoding/json"
	"fmt"
	"github.com/briankohler/consulhelper"
	"net/http"
	"strconv"
	"strings"
)

/*
  GET /health
  GET /v2/dnsapi/HEALTH
*/

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

/*
  GET /v2/dnsapi/data
*/
func GetTinyData(w http.ResponseWriter, r *http.Request) {
	tv := r.FormValue("t")
	//fmt.Printf("Formvalue of t=%v\n", tv)
	if tv == "" {
		tv = "600" //default

	}
	t, err := strconv.ParseInt(tv, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), 404)
	}
	ct := r.FormValue("ct")
	str, err := consulhelper.ListKeys(consul, consul_keyspace, t)
	if err != nil {
		http.Error(w, err.Error(), 404)
	} else {
		if ct == "json" {
			byt, err := json.Marshal(str)
			if err == nil {
				w.Write([]byte(byt))
			} else {
				http.Error(w, err.Error(), 404)
			}
		} else {
			w.Write([]byte(fmt.Sprintf("%v\n", strings.Join(str, "\n"))))
		}
	}
}

func AddRecord(d DataHandler) error {

	if err := d.Check(); err == nil {
		cg := ConsulRecGenerator(d)
		ct := cg.GenConsulRecord()
		err := consulhelper.PostValue(consul, ct.record, ct.rtype, ct.key, ct.value, ct.tiny)
		return err
	} else {
		return err
	}
}

func EditRecord(now DataHandler, new DataHandler) error {

	// OK to proceed with the edit
	err := new.Check()
	if err == nil {
		cg := ConsulRecGenerator(now)
		ct := cg.GenConsulRecord()

		ncg := ConsulRecGenerator(new)
		nct := ncg.GenConsulRecord()
		err := consulhelper.EditValue(consul, nct.record, nct.rtype, nct.key, ct.value, nct.value, ct.tiny, nct.tiny)
		return err
	}
	return err
}

func DeleteRecord(d DataHandler) error {

	cg := ConsulRecGenerator(d)
	ct := cg.GenConsulRecord()

	err := consulhelper.DeleteValue(consul, ct.record, ct.key, ct.value, ct.tiny)
	return err
}

func VerifyURL(rt string, method string, url string) error {
	fmt.Printf("Record type: %v, Method: %v, Url: %v\n", rt, method, url)
	s := strings.Split(url, "/")
	type RecordTypeMethod struct {
		method, rt string
	}
	fc := make(map[RecordTypeMethod]int)

	fc[RecordTypeMethod{"POST", "ALIAS"}] = 7
	fc[RecordTypeMethod{"DELETE", "ALIAS"}] = 7
	fc[RecordTypeMethod{"PUT", "ALIAS"}] = 9
	fc[RecordTypeMethod{"POST", "CNAME"}] = 7
	fc[RecordTypeMethod{"DELETE", "CNAME"}] = 7
	fc[RecordTypeMethod{"PUT", "CNAME"}] = 9
	fc[RecordTypeMethod{"POST", "A"}] = 7
	fc[RecordTypeMethod{"DELETE", "A"}] = 7
	fc[RecordTypeMethod{"PUT", "A"}] = 9
	if len(s) == fc[RecordTypeMethod{method, rt}] {
		return nil
	}
	err := fmt.Errorf("Count Mismatch: Expected: %v, Got: %v(%v)", fc[RecordTypeMethod{method, rt}], len(s), url)
	return err
}
