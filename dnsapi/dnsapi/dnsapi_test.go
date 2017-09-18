package main

/*
End to End Test of the dnsapio app server
Requires: Running dnsapi server, configured to store values in consul
*/

import "testing"
import "net/http"
import "io/ioutil"
import "encoding/json"
import "os"

import "fmt"

// AUTH_DOMAINS="local.docker"

type TestCases struct {
	m       string
	u       string
	b       bool
	tinystr string
}

var baseurl string

func TestCheckDnsapi(t *testing.T) {
	tests := []TestCases{
		{"LMN", "/LOC/la/1.2.3.4", false, ""},              //Unhandled Method
		{"POST", "/LOC/la/1.2.3.4", true, "%la:1.2.3.4"},   //add a new loc
		{"POST", "/LOC/lx/1.2.3", true, "%lx:1.2.3"},       //add a new loc
		{"POST", "/LOC/la/1.2.3.5", false, ""},             //already exists
		{"POST", "/LOC/lax/1.2.3.5", true, "%lax:1.2.3.5"}, //> 2 ascii chars
		{"POST", "/LOC/lb/a", false, ""},                   // not an IP add/subnet

		{"POST", "/CNAME/cname.local.test1/a.local.docker/3600", true, "Ccname.local.test1:a.local.docker:3600"},            //add a new CNAME record
		{"POST", "/CNAME/cname.local.test1/b.local.docker/3600", true, "Ccname.local.test1:b.local.docker:3600"},            //add another CNAME record in RR
		{"POST", "/CNAME/cname.local.test/b.local.docker/3600", false, ""},                                          //unsupported fqdn
		{"POST", "/CNAME/cname.local.test/b/3600", false, ""},                                                  //Invalid value
		{"POST", "/CNAME/cname.local.test/b.local.docker/x", false, ""},                                             //Invalid ttl
		{"POST", "/CNAME/cname.local.test1/x.local.docker/3600?loc=la", true, "Ccname.local.test1:x.local.docker:3600::la"}, //location specific
		{"POST", "/CNAME/cname.local.test1/x.local.docker/3600?loc=lx", true, "Ccname.local.test1:x.local.docker:3600::lx"}, // same record, different loc
		{"XYZ", "/CNAME/cname.local.test1/x.local.docker/3600?loc=la", false, ""},                                   //Unhandled Method
		{"POST", "/CNAME/cname.local.test1/y.local.docker/3600?loc=lb", false, ""},                                  //non-existent loc

		{"PUT", "/CNAME/cname.local.test1/b.local.docker/3600/b.local.docker/1200", true, "Ccname.local.test1:b.local.docker:1200"},              //Change ttl only
		{"PUT", "/CNAME/cname.local.test1/b.local.docker/1200/c.local.docker/3600", true, "Ccname.local.test1:c.local.docker:3600"},              //Change everything
		{"PUT", "/CNAME/cname.local.test1/c.local.docker/3600/a.local.docker/3600", false, ""},                                           //Point to an already existing record
		{"PUT", "/CNAME/cname.local.test1/c.local.docker/3600/d.local.docker/x", false, ""},                                              //Bad ttl
		{"PUT", "/CNAME/cnamex.local.test1/b.local.docker/45/c.local.docker/55", false, ""},                                              //Edit non-existent record
		{"PUT", "/CNAME/cname.local.test1/x.local.docker/3600/xa.local.docker/1200?loc=lb", false, ""},                                   //loc mismatch
		{"PUT", "/CNAME/cname.local.test1/x.local.docker/3600/xa.local.docker/1200?loc=la", true, "Ccname.local.test1:xa.local.docker:1200::la"}, //location specific
		{"PUT", "/CNAME/cname.local.test1/x.local.docker/3600/xb.local.docker/1200?loc=lx", true, "Ccname.local.test1:xb.local.docker:1200::lx"}, //record for diff location

		{"POST", "/A/a.local.test1/1.2.3.4/3600", true, "=a.local.test1:1.2.3.4:3600"},              //Add new A record
		{"ABC", "/A/a.local.test1/1.2.3.4/3600", false, ""},                                      //Unhandled Method
		{"POST", "/A/a.local.test1/1.2.3.5/3600", true, "=a.local.test1:1.2.3.5:3600"},              //Add new A record
		{"POST", "/A/a1.local.test1/1.2.3/3600", false, ""},                                      //Bad IP
		{"POST", "/A/a1.local.test1/1.2.3/y1", false, ""},                                        //Invalid TTL
		{"POST", "/A/a1.local.test/1.2.3.4/3600", false, ""},                                     //Invalid Domain
		{"POST", "/A/cname.local.test1/1.2.3.4/3600", false, ""},                                 //cname.local.test1 is already a CNAME
		{"POST", "/A/a1.local.test1/1.2.3.4/3600?loc=la", true, "=a1.local.test1:1.2.3.4:3600::la"}, //Add new A record with loc
		{"POST", "/A/a1.local.test1/1.2.3.6/3600?loc=la", true, "=a1.local.test1:1.2.3.6:3600::la"}, //Add new A record with loc
		{"POST", "/A/a1.local.test1/1.2.3.6/3600?loc=lx", true, "=a1.local.test1:1.2.3.6:3600::lx"}, //Same record, different location
		{"POST", "/A/a1.local.test1/1.2.3.4/3600?loc=lb", false, ""},                             // non-existent loc

		{"PUT", "/A/a.local.test1/1.2.3.4/3600/1.2.3.4/1200", true, "=a.local.test1:1.2.3.4:1200"},              //change ttl only
		{"PUT", "/A/a.local.test1/1.2.3.4/1200/1.2.3.6/3600", true, "=a.local.test1:1.2.3.6:3600"},              //change everything
		{"PUT", "/A/a.local.test1/1.2.3.6/3600/1.2.3.6/x", false, ""},                                        //Bad ttl
		{"PUT", "/A/a.local.test1/1.2.3.7/3600/1.2.3.6/1200", false, ""},                                     //Edit Non-existent record
		{"PUT", "/A/a.local.test1/1.2.3.6/3600/1.23.6/1200", false, ""},                                      //Bad IP
		{"PUT", "/A/a1.local.test1/1.2.3.4/3600/1.2.3.5/1200?loc=lb", false, ""},                             //Edit with wrong loc
		{"PUT", "/A/a1.local.test1/1.2.3.4/3600/1.2.3.7/1200?loc=la", true, "=a1.local.test1:1.2.3.7:1200::la"}, //Edit with loc
		{"PUT", "/A/a1.local.test1/1.2.3.6/3600/1.2.3.8/1200?loc=lx", true, "=a1.local.test1:1.2.3.8:1200::lx"}, //Edit with loc

		{"PQR", "/ALIAS/alias.local.test1/1.2.3.4/3600", false, ""},                             //Unhandled Method alias
		{"POST", "/ALIAS/alias.local.test1/1.2.3.4/3600", true, "+alias.local.test1:1.2.3.4:3600"}, //Add alias
		{"POST", "/ALIAS/alias.local.test1/1.2.3.5/3600", true, "+alias.local.test1:1.2.3.5:3600"}, //Add another record in RR
		{"POST", "/ALIAS/alias.local.test/1.2.3.5/3600", false, ""},                             //Invalid domain
		{"POST", "/ALIAS/alias.local.test1/1.2.3.5/z1", false, ""},                              // Invalid ttl
		{"POST", "/ALIAS/alias.local.test1/1.2.5/3600", false, ""},                              // Invalid IP

		{"POST", "/ALIAS/alias1.local.test1/1.2.3.4/3600?loc=la", true, "+alias1.local.test1:1.2.3.4:3600::la"}, //Add new A record with loc

		{"POST", "/ALIAS/alias1.local.test1/1.2.3.4/3600?loc=lx", true, "+alias1.local.test1:1.2.3.4:3600::lx"}, //Same rec. location changes

		{"POST", "/ALIAS/alias1.local.test1/1.2.3.5/1200?loc=lx", true, "+alias1.local.test1:1.2.3.5:1200::lx"}, //Works for Alias

		{"POST", "/ALIAS/alias1.local.test1/1.2.3.4/3600?loc=lb", false, ""}, // non-existent loc

		{"PUT", "/ALIAS/alias1.local.test1/1.2.3.5/1200/3.4.5.6/2400?loc=lx", true, "+alias1.local.test1:3.4.5.6:2400::lx"}, //Works for Alias
		{"PUT", "/ALIAS/alias1.local.test1/1.2.3.4/3600/4.5.6.7/2400?loc=lx", true, "+alias1.local.test1:4.5.6.7:2400::lx"}, //Works for Alias
		{"PUT", "/ALIAS/alias1.local.test1/1.2.3.4/3600/2.3.4.5/1200?loc=la", true, "+alias1.local.test1:2.3.4.5:1200::la"}, //Edit with loc
		{"PUT", "/ALIAS/alias.local.test1/1.2.3.4/3600/1.2.3.4/1200", true, "+alias.local.test1:1.2.3.4:1200"},              //change ttl only
		{"PUT", "/ALIAS/alias.local.test1/1.2.3.4/1200/1.2.3.6/3600", true, "+alias.local.test1:1.2.3.6:3600"},              //change everything
		{"PUT", "/ALIAS/alias.local.test1/1.2.3.6/3600/1.2.3.6/x", false, ""},                                            //Bad ttl
		{"PUT", "/ALIAS/alias.local.test1/1.2.3.7/3600/1.2.3.6/1200", false, ""},                                         //Edit Non-existent record
		{"PUT", "/ALIAS/alias.local.test1/1.2.3.6/3600/1.23.6/1200", false, ""},                                          //Bad IP
		{"PUT", "/ALIAS/alias1.local.test1/1.2.3.4/3600/2.3.4.5?loc=lx", false, ""},                                      //Wrong loc

		{"PUT", "/NS/ns.local.test1", false, ""},                                         // Unhandled method
		{"POST", "/NS/ns.local.test1", true, ".ns.local.test1:127.0.0.1:lo:60"},             // ns to loopback
		{"POST", "/NS/ns.local.testX", false, ""},                                        // not in AUTH_DOMAINS
		{"POST", "/NS/ns1.local.test1/1.2.3.400/a/60", false, ""},                        // bad IP
		{"POST", "/NS/ns1.local.test1/1.2.3.4/a/60", true, ".ns1.local.test1:1.2.3.4:a:60"}, // ns rec
		{"POST", "/NS/ns1.local.test1/3.4.5.6/b/60", true, ".ns1.local.test1:3.4.5.6:b:60"}, // add another

		{"POST", "/RAW/raw.local.test1/Craw.local.test1:a:45", false, ""},                  // format has its own API
		{"POST", "/RAW/raw.local.test1/Xraw.local.test1:a:45", false, ""},                  // X is unknown record type
		{"UNH", "/RAW/raw.local.test1/@raw.local.test1:a:45", false, ""},                   // Unhandled Method
		{"POST", "/RAW/raw.local.test1/@raw.local.test1:a:45", true, "@raw.local.test1:a:45"}, // This is OK

		{"PUT", "/RAW/raw.local.test1/@raw.local.test1:b:45/@raw.local.test1:b:30", false, ""},                  // Non existent record
		{"PUT", "/RAW/raw.local.test1/@raw.local.test1:a:45/Braw.local.test1:b:30", false, ""},                  // Unhandled record type
		{"PUT", "/RAW/raw.local.test1/@raw.local.test1:a:45/@raw.local.test1:b:30", true, "@raw.local.test1:b:30"}, // This is OK

		{"PUT", "/NS/ns1.local.test1/1.2.3.4/a/60/1.2.3.4/a/60", false, ""},                              // noop
		{"PUT", "/NS/ns1.local.test1/1.2.3.4/a/60/1.2.3.4/a/y1", false, ""},                              // bad ttl
		{"PUT", "/NS/ns1.local.test1/1.2.3.4/a/60/1.2.3/a/60", false, ""},                                // bad IP
		{"PUT", "/NS/ns1.local.test1/1.2.3.4/a/60/2.3.4.5/a/1200", true, ".ns1.local.test1:2.3.4.5:a:1200"}, // a few changes

		{"POST", "/CNAME/alias.local.test1/b.local.docker/44", false, ""}, //Alias is already in alias
		{"POST", "/NS/alias1.local.test1/1.2.3.4/a/60", false, ""},   //alias1 is already alias

		{"DELETE", "/ALIAS/alias.local.test1/1.2.3.6/3600", true, "+alias.local.test1:1.2.3.6:3600"},                  //Delete
		{"DELETE", "/ALIAS/alias.local.test1/1.2.3.5/3600", true, "+alias.local.test1:1.2.3.5:3600"},                  //Delete
		{"DELETE", "/CNAME/cname.local.test1/c.local.docker/3600", true, "Ccname.local.test1:c.local.docker:3600"},              //delete record
		{"DELETE", "/CNAME/cname.local.test1/a.local.docker/3600", true, "Ccname.local.test1:a.local.docker:3600"},              //add another CNAME record in RR
		{"DELETE", "/LOC/la/1.2.3.4", true, "%la:1.2.3.4"},                                                      //Delete
		{"DELETE", "/LOC/lx/1.2.3", true, "%lx:1.2.3"},                                                          //Delete
		{"DELETE", "/LOC/lax/1.2.3.5", true, "%lax:1.2.3.5"},                                                    //Delete
		{"DELETE", "/CNAME/cname.local.test1/xa.local.docker/1200?loc=lb", false, ""},                                   //location is wrong
		{"DELETE", "/CNAME/cname.local.test1/xa.local.docker/1200?loc=la", true, "Ccname.local.test1:xa.local.docker:1200::la"}, //location specific
		{"DELETE", "/CNAME/cname.local.test1/xb.local.docker/1200?loc=lx", true, "Ccname.local.test1:xb.local.docker:1200::lx"}, //location specific
		{"DELETE", "/A/a.local.test1/1.2.3.6/3600", true, "=a.local.test1:1.2.3.6:3600"},                              //Delete
		{"DELETE", "/A/a.local.test1/1.2.3.5/3600", true, "=a.local.test1:1.2.3.5:3600"},                              //Delete
		{"DELETE", "/A/a1.local.test1/1.2.3.6/3600?loc=la", true, "=a1.local.test1:1.2.3.6:3600::la"},                 //Non existent Record
		{"DELETE", "/A/a1.local.test1/1.2.3.7/1200?loc=la", true, "=a1.local.test1:1.2.3.7:1200::la"},                 //Delete with loc
		{"DELETE", "/A/a1.local.test1/1.2.3.8/1200?loc=lx", true, "=a1.local.test1:1.2.3.8:1200::lx"},                 //Delete with loc
		{"DELETE", "/ALIAS/alias1.local.test1/2.3.4.5/1200?loc=lm", false, ""},                                     //Wrong loc
		{"DELETE", "/ALIAS/alias1.local.test1/3.4.5.6/2400?loc=lx", true, "+alias1.local.test1:3.4.5.6:2400::lx"},     //Works for Alias
		{"DELETE", "/ALIAS/alias1.local.test1/4.5.6.7/2400?loc=lx", true, "+alias1.local.test1:4.5.6.7:2400::lx"},     //Works for Alias
		{"DELETE", "/ALIAS/alias1.local.test1/2.3.4.5/1200?loc=la", true, "+alias1.local.test1:2.3.4.5:1200::la"},     //Edit with loc
		{"DELETE", "/NS/ns.local.test1", true, ".ns.local.test1:127.0.0.1:lo:60"},                                     // ns to loopback
		{"DELETE", "/NS/ns1.local.test1/3.4.5.6/b/60", true, ".ns1.local.test1:3.4.5.6:b:60"},                         // add another
		{"DELETE", "/NS/ns1.local.test1/2.3.4.5/a/1200", true, ".ns1.local.test1:2.3.4.5:a:1200"},                     // delete another
		{"DELETE", "/RAW/raw.local.test1/@raw.local.test1:b:30", true, "@raw.local.test1:b:30"},                          // This is OK
		{"POST", "/CNAME/alias.local.test1/b1.local.docker/44", true, "Calias.local.test1:b1.local.docker:44"},                  //Try to change a record type, provided there are no more entries for it
		{"DELETE", "/CNAME/alias.local.test1/b1.local.docker/44", true, "Calias.local.test1:b1.local.docker:44"},                //Try to change a record type, provided there are no more entries for it

		//Random unhandled Queries
		{"POST", "/CNAME/alias.local.test1/b1.local.docker/44/a/g", false, ""},
		{"POST", "/A/alias.local.test1/1.2.3.4/44/22/g/x/y", false, ""},
		{"PUT", "/ALIAS/alias.local.test1/1.2.3.4/44/22.2.4.3/100/x/2", false, ""},
	}

	baseurl = os.Getenv("DNS_URL")
	if baseurl == "" {
		baseurl = "http://localhost:9080"
	}
	baseurl = fmt.Sprintf("%v/v2/dnsapi", baseurl)
	for _, c := range tests {
		//t.Logf("\n\n")
		//t.Log(i)
		//t.Log(c)

		if c.b == true {
			err := CheckTiny(c.tinystr)
			if c.m == "DELETE" {
				if err != nil {
					t.Errorf("ERROR:%v:%v:%v", err, c.m, c.tinystr)
					continue
				}
			} else {
				if err == nil {
					t.Errorf("ERROR:%v:%v:%v", err, c.m, c.tinystr)
					continue
				}
			}
		}

		client := &http.Client{}
		url := fmt.Sprintf("%v%v", baseurl, c.u)
		req, err := http.NewRequest(c.m, url, nil)
		resp, err := client.Do(req)
		bx := false
		var bodyg string //global var for body
		if err == nil {
			body, _ := ioutil.ReadAll(resp.Body)
			bodyg = string(body)
			if bodyg == "OK" {
				bx = true
				//t.Logf("GOOD: %v", string(body))
			}
		} else {
			t.Errorf("ERROR: %v", err)
			continue
		}
		//t.Log(bx)
		if bx != c.b {
			t.Log(c)
			t.Errorf("ERROR: %v: %v %v %v bx(%v) != c.b(%v)", err, bodyg, c.m, url, bx, c.b)
			continue
		} else {
			if c.b == true {
				err := CheckTiny(c.tinystr)
				if c.m == "DELETE" {
					if err == nil {
						t.Errorf("ERROR:%v:%v:%v:%v:%v", err, c.m, bodyg, c.tinystr, url)
					} else {
						//t.Logf("GOOD : %v :  NOT FOUND in data even after DELETE", c.tinystr)

					}
				} else {
					if err != nil {
						t.Errorf("ERROR:%v:%v:%v:%v:%v", err, c.m, bodyg, c.tinystr, url)
					} else {
						//t.Logf("GOOD :  %v NOT FOUND in data ", c.tinystr)
					}
				}
			} else {
				//t.Logf("GOOD : ")
			}
		}
	}

}

func CheckTiny(s string) error {
	url := fmt.Sprintf("%v/data?ct=json", baseurl)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var dat []string
	if err := json.Unmarshal(data, &dat); err != nil {
		return err
	}
	for _, each := range dat {
		if string(each) == s {
			return nil
		}
	}
	err = fmt.Errorf("No can find %v", s)
	return err
}
