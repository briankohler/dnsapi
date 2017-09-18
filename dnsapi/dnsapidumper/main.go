package main

import (
	"bufio"
	"fmt"
	"github.com/briankohler/consulhelper"
	"github.com/briankohler/log"
	"github.com/codegangsta/cli"
	"os"
        "net"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

var lastsecs int64
var tinydatadir string

/*
Env Vars
DNSAPIDUMPER_NEWDATA_CHANGETHRESHOLD : Threshold of acceptable changepct [50
DNSAPIDUMPER_TINYDATADIR : localtion of tinydata file[service/tinydns/root]
DNSAPIDUMPER_CHECK_DURATION: Last n secs to check for changes[600]
AUTH_DOMAINS: for unbound flush_zone
*/

func main() {
	app := cli.NewApp()
	ls := os.Getenv("DNSAPIDUMPER_CHECK_DURATION")
	if ls == "" {
		lastsecs = 600
	} else {
		lastsecs, _ = strconv.ParseInt(string(ls), 10, 64)
	}

	tinydatadir = os.Getenv("DNSAPIDUMPER_TINYDATADIR")
	if tinydatadir == "" {
		tinydatadir = "/service/tinydns/root"
	}
	tinydatafile := fmt.Sprintf("%v/data", tinydatadir)

	app.Name = "dnsapidumper"
	app.Usage = "Dumps out a tinydns formatted file from Consul\nEnv\n\tDNSAPIDUMPER_NEWDATA_CHANGETHRESHOLD\n\tDNSAPIDUMPER_TINYDATADIR\n\tDNSAPIDUMPER_CHECK_DURATION\n\tAUTH_DOMAINS\n\tCONSUL_KEYSPACE"
	app.Version = "0.1.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "out, o",
			Value: tinydatafile,
			Usage: "Filename to write to",
		},
	}

	app.Action = func(c *cli.Context) {
		writeToFile(c)
	}

	app.Run(os.Args)

}

func writeToFile(c *cli.Context) {
	consul, err := consulhelper.Initialize_consul()
	consul_keyspace := os.Getenv("CONSUL_KEYSPACE")
	if consul_keyspace == "" {
		consul_keyspace = "tinydns"
	}

	var str sort.StringSlice
	str, err = consulhelper.ListKeys(consul, consul_keyspace, lastsecs)

	str.Sort()
	if err == nil {
		thresholdpct := os.Getenv("DNSAPIDUMPER_NEWDATA_CHANGETHRESHOLD")
		if thresholdpct == "" {
			thresholdpct = "50"
		}
		thresholdpct_num, _ := strconv.ParseInt(thresholdpct, 10, 64)
		tinydatafile := fmt.Sprintf("%v/data", tinydatadir)
		if err := NewTinyDataOK(str, tinydatafile, int(thresholdpct_num)); err == nil {
                        log.Info("Writing to file ", fmt.Sprintf(c.String("out")))
			f, _ := os.Create(c.String("out"))
			defer f.Close()
			w := bufio.NewWriter(f)
			for k := range str {
				w.WriteString(fmt.Sprintf("%v\n", str[k]))
			}
			w.Flush()
                        log.Info(strconv.Itoa(len(str)), " DNS records written")
			if err := os.Chdir(tinydatadir); err == nil {
				if err := MakeTinyData(); err == nil {
					err := FlushUnboundZones()
                                        if err != nil {
					     log.Error(err.Error())
					}
				} else {
					log.Error("unbound cache flush pending")
					log.Error(err.Error())
				}
			} else {
				log.Error(err.Error())
			}
		} else {
			log.Error(err.Error())
		}
	} else {
		log.Error(err.Error())
	}
}

func FlushUnboundZones() error {
	cmd := "/usr/bin/unbound-control"
	if os.Getenv("AUTH_DOMAINS") == "" {
		err := fmt.Errorf("Error: AUTH_DOMAINS need to be defined")
		return err
	}
	ad := strings.Split(os.Getenv("AUTH_DOMAINS"), ",")
        ad = append(ad, "10.in-addr.arpa")
	log.Debug("AUTH_DOMAINS ", ad, " size = ", strconv.Itoa(len(ad)))

	for a := range ad {
		var args []string
		if os.Getenv("UNBOUND_REMOTE_HOST") == "" {
			args = []string{"flush_zone", ad[a]}
		} else {
                        ip4, ip6 := net.LookupHost(os.Getenv("UNBOUND_REMOTE_HOST"))
                        log.Debug("resolved ", ip4, " and ", ip6)
			args = []string{"-s", strings.Join(ip4, ""), "flush_zone", ad[a]}
		}
		out, err := exec.Command(cmd, args...).CombinedOutput() 
                if err == nil {
			log.Info("flush ", ad[a], " - SUCCESS, output - ", fmt.Sprintf("%s", out))
		} else {
			err := fmt.Errorf("Error: ", fmt.Sprintf("%v", cmd), " ", fmt.Sprintf("%v", args), " output - ", fmt.Sprintf("%s", out))
			return err
		}
	}
	return nil
}

func MakeTinyData() error {
        var args []string
        args = []string{"-s"}
        cmd := exec.Command("/usr/bin/make", args...)
        makefilepath := strings.Split(tinydatadir, "/")
        makefilepath = makefilepath[:len(makefilepath)-1]
        cmd.Dir = strings.Join(makefilepath, "/")
	out, err := cmd.CombinedOutput()
        if err == nil { 
		log.Info("make - SUCCESS, output - ", fmt.Sprintf("%s", out))
		return nil
	} else {
		err1 := fmt.Sprintf("make FAIL: %v: %v", err, os.Stderr)
		err2 := fmt.Errorf("make FAIL: %v: %v", err, os.Stderr)
		log.Error(err1)
		return err2
	}
}

func NewTinyDataOK(s []string, tinydatafile string, thresholdpct int) error {
	if thresholdpct < 0 { // Once we are confident this just works
		return nil
	}
	if _, err := os.Stat(tinydatafile); os.IsNotExist(err) {
		f, errC := os.Create(tinydatafile)
		if errC != nil {
			return errC
		}
		f.Close()
	}
	if file, err := os.Open(tinydatafile); err == nil {

		// make sure it gets closed
		defer file.Close()

		// create a new scanner and read the file line by line
		scanner := bufio.NewScanner(file)
		fileh := make(map[string]int)
		for scanner.Scan() {
			fileh[scanner.Text()] = 1
		}
		// check for errors
		if err := scanner.Err(); err != nil {
			return err
		}
		count := 0
		origcount := len(fileh)
		if origcount == 0 {
			if os.Getenv("CREATE_TINYDATA_FILE") == "1" {
				return nil
			} else {
				err := fmt.Errorf("Tinydata is empty but Env CREATE_TINYDATA_FILE is not set to 1, Bailing out")
				return err
			}
		}
		for k := range s {
			if fileh[s[k]] == 1 {
				delete(fileh, s[k])
			} else { // New line inserted to data
				count += 1
			}
		}
		if (count == 0) && (len(fileh) == 0) {
			err := fmt.Errorf("New data exactly same as tinydata on disk .. ignoring")
			return err
		}
		if changepct := (count + len(fileh)) / origcount; changepct > thresholdpct {
			err := fmt.Errorf("Change pct (%v) > Thresholdpct(%v)", changepct, thresholdpct)
			return err
		}
	} else {
		return err
	}
	return nil

}
