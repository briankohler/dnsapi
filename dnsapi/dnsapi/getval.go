package main

import (
	"encoding/json"
	"github.com/briankohler/consulhelper"
	log "github.com/briankohler/log"
	"net/http"
)

//curl   localhost:9080/v2/dnsapi/GETVALFORSTR?str='tinydns/A'

func PrintGetValForStr(w http.ResponseWriter, r *http.Request) {
	log.Debug("LOG: ", r.Method, r.URL)
	p := r.FormValue("str")
	s, err := consulhelper.PrintVal(consul, p)
	if err == nil {
		var dat map[string]interface{}
		if err := json.Unmarshal([]byte(s), &dat); err == nil {
			b, err := json.MarshalIndent(dat, "", "  ")
			if err == nil {
				b2 := append(b, '\n')
				w.Write([]byte(b2))
				return
			}
		}
	} else {
		http.Error(w, err.Error(), 404)
	}
	return
}
