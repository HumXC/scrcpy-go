package option

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	args := []string{
		"power_on=false",
		"tunnel_forward=false",
		"ss",
	}
	opt, err := Parse(args)
	if err != nil {
		t.Error(err)
	}
	ar := opt.ToArgs()
	assert.Equal(t, 1, len(ar))
	assert.Equal(t, "power_on=false", ar[0])
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
	setValue(intV, "12")
	// set bool value
	setValue(boolV, "true")
	// set string value
	setValue(stringV, "hello")

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
