package main

import (
	"github.com/gorilla/mux"
	//"net/http"
)

func Handler() *mux.Router {

	r := mux.NewRouter()
	r.HandleFunc("/health", HealthCheck).Methods("GET")
	apiRouter := r.PathPrefix(DNSPREFIX).Subrouter()
	apiRouter.HandleFunc("/HEALTH", HealthCheck).Methods("GET")

	//CNAME
	apiRouter.HandleFunc("/CNAME/{key}/{value}/{ttl}", HandleCNAME) //Handle both POST & DELETE here
	apiRouter.HandleFunc("/CNAME/{key}/{value}/{ttl}/{nvalue}/{nttl}", HandleCNAME)

	//A
	apiRouter.HandleFunc("/A/{key}/{value}/{ttl}", HandleA) //Handle both POST & DELETE here
	apiRouter.HandleFunc("/A/{key}/{value}/{ttl}/{nvalue}/{nttl}", HandleA)

	//ALIAS
	apiRouter.HandleFunc("/ALIAS/{key}/{value}/{ttl}", HandleALIAS) //Handle both POST & DELETE here
	apiRouter.HandleFunc("/ALIAS/{key}/{value}/{ttl}/{nvalue}/{nttl}", HandleALIAS)

	//LOC - location
	apiRouter.HandleFunc("/LOC/{key}/{value}", HandleLOC) //Handle both POST & DELETE here
	apiRouter.HandleFunc("/LOC/{key}/{value}/{nvalue}", HandleLOC)

	//NSLO - A simplified call for tinydns on localhost. Is a subset of the NS call.
	apiRouter.HandleFunc("/NS/{key}", HandleNS).Methods("POST")
	apiRouter.HandleFunc("/NS/{key}", HandleNS).Methods("DELETE")

	//NS
	apiRouter.HandleFunc("/NS/{key}/{ip}/{name}/{ttl}", HandleNS)
	apiRouter.HandleFunc("/NS/{key}/{ip}/{name}/{ttl}/{nip}/{nname}/{nttl}", HandleNS)

	//RAW
	apiRouter.HandleFunc("/RAW/{key}/{tinystr}", HandleRAW)
	apiRouter.HandleFunc("/RAW/{key}/{tinystr}/{newtinystr}", HandleRAW)

        //SOA
        apiRouter.HandleFunc("/SOA/{key}/{name}", HandleSOA)

	//Tinydata
	apiRouter.HandleFunc("/data", GetTinyData).Methods("GET")

	apiRouter.HandleFunc("/GETVALFORSTR", PrintGetValForStr).Methods("GET")
	return r

}
