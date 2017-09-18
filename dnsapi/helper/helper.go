package helper

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func CheckFQDN(fqdn []byte) (bool, error) {
	auth_d := os.Getenv("AUTH_DOMAINS")
	if len(auth_d) == 0 {
		return false, errors.New("AUTH_DOMAINS: Env var is not defined. Contains comma separated fqdn")
	}
        auth_domains := []string{auth_d, "10.in-addr.arpa"}
	// auth_domains := strings.Split(authd, ",")
	fqdn_str := string(fqdn)
	fqdn_arr := strings.Split(fqdn_str, ".")
	fl := len(fqdn_arr)
	for a := range auth_domains {
		b := strings.Split(auth_domains[a], ".")
		bl := len(b)
		if bl > fl {
			continue
		}

		outer := -1
		for i := 1; i <= bl; i++ {
			outer = i
			if strings.TrimSpace(b[bl-i]) != fqdn_arr[fl-i] {
				break
			}
		}
		if outer == bl {
			return true, nil
		}
	}
	return false, fmt.Errorf("fqdn (%v) not in AUTH_DOMAINS", string(fqdn))
}
