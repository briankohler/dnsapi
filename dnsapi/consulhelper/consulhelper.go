package consulhelper

// CONSUL DATA FORMAT
//
//  The tinydns data is stored in consul as a json string.
//  A dns query consists of a query for a single record, resulting in one or more answers
//  query record is the key in consul
//
//  the value is a json representation of a two level hash, whereby :
//  	Key is each of the answers
//  	Value is the full tinydns record representation of the specific individual record
//  	Example CNAME Record. Lets say:
//  		a.local.docker  => a1.local.docker,a2.local.docker,a3.local.docker
//  		Consul Key => tinydns/CNAME/a.local.docker
//  		Value => {
//  		         "a1.local.docker": "Ca.local.docker:a1.local.docker:3600",
//  		         "a2.local.docker": "Ca.local.docker:a2.local.docker:3600",
//  		         "a3.local.docker": "Ca.local.docker:a3.local.docker:3600",
//  			 }
//  POST will always end up adding a new record
//  EDIT will modify the value of an existing record
//  DELETE deletes an existing record ( value )
//

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/briankohler/log"
	consulapi "github.com/hashicorp/consul/api"
	"os"
	"strconv"
	"time"
)

type ConsulApiClient *consulapi.Client

const lockbase = "TINYDNSLOCK"
const locktimeout = 10

/*
Get all keys from Consul; this queries the KV store
*/
func ListKeys(c *consulapi.Client, s string, lastnsecs int64) ([]string, error) {

	k := fmt.Sprintf("%v/LASTUPDATETIME", lockbase)
	kv := c.KV()
	kvp, _, _ := kv.Get(k, nil)
	var t int64
	if kvp != nil {
		v, _ := strconv.ParseInt(string(kvp.Value), 10, 64)
		v += lastnsecs
		now := time.Now()
		t = now.Unix()
		//fmt.Printf("Update+durtn=%v,Nowtime=%v,ValueinConsul=%v,lastnsecs=%v\n", v, t, string(kvp.Value), lastnsecs)
		if (t > v) && (lastnsecs > -1) { //negative number to get all the data - an override
			err := fmt.Errorf("NOUPDATES in the last %v secs", lastnsecs)
			return nil, err
		}
	}

	kv = c.KV()
	kvl, _, err := kv.Keys(s, "", &consulapi.QueryOptions{RequireConsistent: true})

	var retv []string
	var th = make(map[string]int)

	if err == nil {
		for _, v := range kvl {
			//fmt.Printf("\n%v\n", v)
			dat, err := GetValueInConsul(c, v)
			if err == nil {
				for _, v1 := range dat {
					//fmt.Printf("\t%v -> %v\n", k, v1)
					if th[v1] != 1 {
						th[v1] = 1
						retv = append(retv, v1)
					}
				}
			}
		}
	} else {
		return nil, nil
	}
	return retv, nil
}

func GetValueInConsul(c *consulapi.Client, s string) (map[string]string, error) { // Value is Key/Val hash in json
	kv := c.KV()
	kvp, _, err := kv.Get(s, nil)
	var dat = make(map[string]string)
	if kvp != nil {
		if err := json.Unmarshal([]byte(kvp.Value), &dat); err != nil {
			return nil, err
		}
	}

	return dat, err

}

func PrintVal(c *consulapi.Client, s string) (string, error) {

	kv := c.KV()
	kvp, _, err := kv.Get(s, nil)
	//fmt.Printf("Fetching values for %v\n", s)
	if kvp != nil {
		//fmt.Println(kvp)
		return string(kvp.Value), err
	}
	return "NIL", nil
}

func unlock(c *consulapi.Client, s string) {
	k := fmt.Sprintf("%v/%v", lockbase, s)
	//fmt.Printf("LOCK=%v\n", k)
	kv := c.KV()
	kvp, _, _ := kv.Get(k, nil)
	if kvp != nil {
		d := &consulapi.KVPair{Key: k, Value: []byte(strconv.Itoa(0))}
		kv.Put(d, nil)
	}

}
func lock(c *consulapi.Client, s string) error {
	k := fmt.Sprintf("%v/%v", lockbase, s)
	kv := c.KV()
	kvp, _, _ := kv.Get(k, nil)
	var t int64
	if kvp != nil {
		v, _ := strconv.ParseInt(string(kvp.Value), 10, 64)
		v += locktimeout
		now := time.Now()
		t = now.Unix()
		if t < v {
			err := fmt.Errorf("LOCK FAIL %v (%v)", v, t)
			return err
		}
	}
	x := fmt.Sprintf("%v", t)
	d := &consulapi.KVPair{Key: k, Value: []byte(x)}
	_, er := kv.Put(d, nil)
	return er
}

func UpdateLastUpdateTime(c *consulapi.Client) error {
	k := fmt.Sprintf("%v/LASTUPDATETIME", lockbase)
	now := time.Now()
	t := now.Unix()
	x := fmt.Sprintf("%v", t)
	kv := c.KV()
	d := &consulapi.KVPair{Key: k, Value: []byte(x)}
	_, er := kv.Put(d, nil)
	return er
}

func DeleteValue(c *consulapi.Client, recordname string, s string, v string, tinystr string) error {

	if err := lock(c, s); err != nil {
		return err
	}
	kvp, err := GetValueInConsul(c, s)
	if err == nil {
		if j := kvp[v]; j != tinystr {
			unlock(c, s)
			err := fmt.Errorf("NON-Existent: Cannot delete Key(%v) consul(%v) ;Input(%v)", v, j, tinystr)
			return err
		}
		delete(kvp, v)
		byt, _ := json.Marshal(kvp)
		kv := c.KV()
		d := &consulapi.KVPair{Key: s, Value: []byte(byt)}
		_, err := kv.Put(d, nil)
		if err == nil {
			_ = UpdateLastUpdateTime(c)
			if len(kvp) == 0 {
				err := DeleteSetRecord(c, recordname)
				if err != nil {
					err1 := fmt.Errorf("%v: Unset Record type FAILED", s)
					unlock(c, s)
					return err1
				}
			}
		}
		unlock(c, s)
		return err
	}
	return err
}

func EditValue(c *consulapi.Client, recordname string, recordtype string, s string, vnow string, vnew string, tinystrnow string, tinystrnew string) error {

	if err := SetRecordType(c, recordname, recordtype); err != nil {
		return err
	}
	if err := lock(c, s); err != nil {
		return err
	}
	kvp, err := GetValueInConsul(c, s)
	//fmt.Printf("VALUES for %v\n%v\ntinystr=%v\nmap=%v\n",s,v,tinystrnow,kvp)
	if err == nil {
		if j := kvp[vnow]; j != tinystrnow {
			unlock(c, s)
			err := fmt.Errorf("Input(%v), Consul(%v): MISMATCH", kvp[vnow], tinystrnow)
			return err
		}
		if tinystrnew == tinystrnow {
			unlock(c, s)
			return errors.New("NowVal=Newval: NOOP")
		}
		if j := kvp[vnew]; j != "" && vnew != vnow { //same string ignored so TTL changes can happen
			unlock(c, s)
			err := fmt.Errorf("New(%v) is already defined as %v", vnew, kvp[vnew])
			return err
		}
		kvp[vnew] = tinystrnew
		if vnow != vnew {
			delete(kvp, vnow)
		}
		byt, _ := json.Marshal(kvp)
		kv := c.KV()
		d := &consulapi.KVPair{Key: s, Value: []byte(byt)}
		_, err := kv.Put(d, nil)
		unlock(c, s)
		if err == nil {
			_ = UpdateLastUpdateTime(c)
		}
		return err
	}
	return err
}

func DeleteSetRecord(c *consulapi.Client, s string) error {
	base := "RECORDLISTBYTYPE"
	s = fmt.Sprintf("%v/%v", base, s)
	if err := lock(c, s); err != nil {
		return err
	}
	kv := c.KV()
	_, err := kv.Delete(s, nil)
	unlock(c, s)
	return err
}

func SetRecordType(c *consulapi.Client, s string, v string) error {
	// Egs: a record cannot have both A & CNAME. Each record can only be of one unique type

	base := "RECORDLISTBYTYPE"
	s = fmt.Sprintf("%v/%v", base, s)
	if err := lock(c, s); err != nil {
		return err
	}
	kv := c.KV()
	kvp, _, err := kv.Get(s, nil)
	if err != nil {
		unlock(c, s)
		return err
	}
	if kvp != nil {
		if string(kvp.Value) == v {
			unlock(c, s)
			return nil
		}
		if string(kvp.Value) != "" {
			err := fmt.Errorf("WORKING on %v as %v whereas it is already %v", s, v, string(kvp.Value))
			unlock(c, s)
			return err
		}
	}
	d := &consulapi.KVPair{Key: s, Value: []byte(v)}
	_, err1 := kv.Put(d, nil)
	unlock(c, s)
	return err1
}

func PostValue(c *consulapi.Client, recordname string, recordtype string, s string, v string, tinystr string) error {

	if err := SetRecordType(c, recordname, recordtype); err != nil {
		return err
	}
	if err := lock(c, s); err != nil {
		return err
	}
	kvp, err := GetValueInConsul(c, s)
	if err == nil {
		if j := kvp[v]; j != "" {
			unlock(c, s)
			err := fmt.Errorf("Input(%v-%v-%v), Exists(%v -> %v)", s, v, tinystr, v, j)
			return err
		}
		kvp[v] = tinystr
		byt, _ := json.Marshal(kvp)
		kv := c.KV()
		d := &consulapi.KVPair{Key: s, Value: []byte(byt)}
		_, err := kv.Put(d, nil)
		unlock(c, s)
		if err == nil {
			_ = UpdateLastUpdateTime(c)
		}
		return err
	}
	return err
}

func Check_exists_in_consul(c *consulapi.Client, s string, v string) (string, error) {

	kvp, err := GetValueInConsul(c, s)
	if err == nil {
		if j := kvp[v]; j == "" {
			err := fmt.Errorf("NON-Existent: Cannot fetch %v -> %v ", s, v)
			return "", err
		}
	}
	return kvp[v], err
}

func Initialize_consul() (*consulapi.Client, error) {
	config := consulapi.DefaultConfig()
	consul_host := os.Getenv("CONSUL_HOST")
	consul_port := os.Getenv("CONSUL_PORT")

	if consul_host == "" {
		consul_host = "localhost"
	}

	if consul_port == "" {
		consul_port = "8500"
	}

	config.Address = fmt.Sprintf("%s:%s", consul_host, consul_port)
	consul, err := consulapi.NewClient(config)
	log.Info("Connecting to consul using ", config.Address)
	return consul, err
}
