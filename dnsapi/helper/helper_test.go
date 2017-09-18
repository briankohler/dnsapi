package helper_test

import "testing"
import "os"
import "fmt"
import "github.com/briankohler/helper"

//import "errors"

type TestCases struct {
	Env  string
	fqdn string
	ret  bool
	//err errors
}

func TestCheckFQDN(t *testing.T) {
	tests := []TestCases{
		{"test.lab1,test.lab2", "a.local.docker", false},
		{"test.lab1,test.lab2", "test.lab1", true},
		{"test.lab1,test.lab2", "test.lab1", true},
		{"", "", false},
	}
	for c, i := range tests {

		t.Log("Verifying CheckFQDN")
		fmt.Println(c)
		fmt.Println(i)
		os.Setenv("AUTH_DOMAINS", i.Env)
		ret, err := helper.CheckFQDN([]byte(i.fqdn))
		if ret != i.ret {
			t.Errorf("%v: Expected %v for %v, but instead got %v : %v\n", c, i.ret, i.fqdn, ret, err)
		}
	}
}
