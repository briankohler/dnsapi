/* dnsapi, the RESTful API in front of TinyDNS


Available Routes:

  get   '/dnsapi/data'
  get   '/dnsapi/CHECKSUM'
  get   '/dnsapi/HEALTH'
  put   '/dnsapi/NS/:key/:value/:ttl/:newvalue/:newttl'
  put   '/dnsapi/LOC/:key/:value/:ttl/:newvalue/:nttl'
  put   '/dnsapi/ALIAS/:key/:value/:ttl/:newvalue/:newttl'
  put   '/dnsapi/A/:key/:value/:ttl/:newvalue/:newttl'
  put   '/dnsapi/CNAME/:key/:value/:ttl/:newvalue/:newttl'
  post  '/dnsapi/CNAME/:key/:value/:ttl'
  post  '/dnsapi/A/:key/:value/:ttl'
  post  '/dnsapi/ALIAS/:key/:value/:ttl'
  post  '/dnsapi/LOC/:key/:value/:ttl'
  post  '/dnsapi/NS/:key/:value/:ttl'
  post  '/dnsapi/NS/:key  // add NS record with 127.0.0.1 as SOA
  post  '/dnsapi/SOA/:key/:value 
  del   '/dnsapi/:recordtype/:key/:value/:ttl'
  

The following environment variables are available to configure the application:

  AUTH_DOMAINS        a comma separated list of domains that are allowed to enter the zone
  CONSUL_HOST         hostname for consul; default localhost
  CONSUL_PORT         port for consul; default 8500
  CONSUL_KEYSPACE     keyspace under which the dns records are stored
  LOGLEVEL            DEBUG/INFO/WARN/ERROR/FATAL; default INFO
  BINDADDR            IP Address to bind to; default 127.0.0.1
  PORT                Port to bind to; default 9080
*/

// TODO Handle loc
// TODO Tests

package main

import (
	"fmt"
	"github.com/briankohler/consulhelper"
	"github.com/briankohler/log"
	"github.com/briankohler/logmiddleware"
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	//	"github.com/gorilla/mux"
	consulapi "github.com/hashicorp/consul/api"
	"net/http"
	"os"
)

const DNSPREFIX = "/v2/dnsapi"

var consul *consulapi.Client // global vars - its a sign of power
var consul_keyspace string

type Input_Record struct {
	name  string
	value string
	ttl   string
	lo    string
	rtype string
}

func main() {

	// Lets get consul in first
	var err error
	consul, err = consulhelper.Initialize_consul()
	if err != nil {
		fmt.Println(err)
		os.Exit(128)
	}
	consul_keyspace = os.Getenv("CONSUL_KEYSPACE")
	if consul_keyspace == "" {
		consul_keyspace = "tinydns"
	}

	r := Handler()

	bindaddr := os.Getenv("BINDADDR")
	bindport := os.Getenv("PORT")

	if bindaddr == "" {
		bindaddr = "127.0.0.1"
	}
	if bindport == "" {
		bindport = "9080"
	}

        // bring in some middleware
        n := negroni.New()
        switch os.Getenv("LOGLEVEL") {
        default:
                n.Use(logmiddleware.NewCustomMiddleware(logrus.InfoLevel, &logrus.JSONFormatter{}, "web"))
        case "DEBUG":
                n.Use(logmiddleware.NewCustomMiddleware(logrus.DebugLevel, &logrus.JSONFormatter{}, "web"))
        case "INFO":
                n.Use(logmiddleware.NewCustomMiddleware(logrus.InfoLevel, &logrus.JSONFormatter{}, "web"))
        case "WARN":
                n.Use(logmiddleware.NewCustomMiddleware(logrus.WarnLevel, &logrus.JSONFormatter{}, "web"))
        case "ERROR":
                n.Use(logmiddleware.NewCustomMiddleware(logrus.ErrorLevel, &logrus.JSONFormatter{}, "web"))
        case "FATAL":
                n.Use(logmiddleware.NewCustomMiddleware(logrus.FatalLevel, &logrus.JSONFormatter{}, "web"))
        }
	n.UseHandler(r)
	log.Info("Starting dnsapi server on ", fmt.Sprintf(bindaddr+":"+bindport))
	http.ListenAndServe(fmt.Sprintf(bindaddr+":"+bindport), n)
}
