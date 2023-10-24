package server_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/HumXC/scrcpy-go/server"
	"github.com/HumXC/scrcpy-go/server/utils"
)

func TestParse(t *testing.T) {
	args := []string{
		"power_on=true",
		"ss",
	}
	opt, err := server.ParseOption(args)
	if err != nil {
		t.Error(err)
	}
	ar := opt.ToArgs()
	fmt.Println(ar)
	if opt.PowerOn != true {
		t.Error(fmt.Errorf("want:true, got: %t", opt.PowerOn))
	}
}

func TestSetValue(t *testing.T) {
	type S struct {
		Int    int
		Bool   bool
		String string
	}
	v := reflect.ValueOf(&S{}).Elem()
	intV := v.FieldByName("Int")
	boolV := v.FieldByName("Bool")
	stringV := v.FieldByName("String")
	// set int value
	utils.SetValue(intV, "12")
	// set bool value
	utils.SetValue(boolV, "true")
	// set string value
	utils.SetValue(stringV, "hello")

	s := v.Interface().(S)
	if s.Int != 12 {
		t.Errorf("intV is not 12, but %d", s.Int)
	}
	if s.Bool != true {
		t.Errorf("boolV is not true, but %t", s.Bool)
	}
	if s.String != "hello" {
		t.Errorf("stringV is not hello, but %s", s.String)
	}
}
