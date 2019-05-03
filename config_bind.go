// Copyright (c) 2018 Dean Jackson <deanishe@deanishe.net>
// MIT Licence - http://opensource.org/licenses/MIT

package aw

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// To populates (tagged) struct v with values from the environment.
func (cfg *Config) To(v interface{}) error {

	binds, err := extract(v)
	if err != nil {
		return err
	}

	for _, bind := range binds {
		if err := bind.Import(cfg); err != nil {
			return err
		}
	}

	return nil
}

// From saves the fields of (tagged) struct v to the workflow's settings in Alfred.
//
// All supported and unignored fields are saved, although empty variables
// (i.e. "") are not overwritten with Go zero values, e.g. "0" or "false".
func (cfg *Config) From(v interface{}) error {

	variables, err := cfg.bindVars(v)
	if err != nil {
		return err
	}

	return cfg.setMulti(variables, false)
}

// extract binding values as {ENVVAR: value} map.
func (cfg *Config) bindVars(v interface{}) (map[string]string, error) {

	variables := map[string]string{}

	binds, err := extract(v)
	if err != nil {
		return nil, err
	}

	for _, bind := range binds {
		if k, v, ok := bind.GetVar(cfg); ok {
			variables[k] = v
		}
	}

	return variables, nil
}

// setMulti batches the saving of multiple variables.
func (cfg *Config) setMulti(variables map[string]string, export bool) error {

	for k, v := range variables {
		cfg.Set(k, v, export)
	}

	return cfg.Do()
}

// binding links an environment variable to the field of a struct.
type binding struct {
	Name     string
	EnvVar   string
	FieldNum int
	Target   interface{}
	Kind     reflect.Kind
}

type bindSource interface {
	GetBool(key string, fallback ...bool) bool
	GetInt(key string, fallback ...int) int
	GetFloat(key string, fallback ...float64) float64
	GetString(key string, fallback ...string) string
}

type bindDest interface {
	GetString(key string, fallback ...string) string
	// SetConfig(key, value string, export bool, bundleID ...string) *Config
	setMulti(variables map[string]string, export bool) error
}

// Import populates the target struct from src.
func (bind *binding) Import(src bindSource) error {

	rv := reflect.Indirect(reflect.ValueOf(bind.Target))

	if bind.FieldNum > rv.NumField() {
		return fmt.Errorf("invalid FieldNum (%d) for %s (%v)", bind.FieldNum, bind.Name, rv)
	}

	value := rv.Field(bind.FieldNum)

	// Ignore empty/unset variables
	if src.GetString(bind.EnvVar) == "" {
		return nil
	}

	return bind.setValue(&value, src)
}

// GetVar populates dst from target struct.
func (bind *binding) GetVar(dst bindDest) (key, value string, ok bool) {

	rv := reflect.Indirect(reflect.ValueOf(bind.Target))

	if bind.FieldNum > rv.NumField() {
		return
	}

	var (
		val     = rv.Field(bind.FieldNum)
		cur     = dst.GetString(bind.EnvVar)
		curZero = isZeroString(cur, val.Kind())
		newZero = isZeroValue(val)
	)

	// field key & value
	key = bind.EnvVar
	value = fmt.Sprintf("%v", val)

	// Don't pull zero-value fields into empty variables.
	if curZero && newZero {
		// log.Printf("[bind] %s: both empty", field.Name)
		return
	}

	ok = true

	return
}

func (bind *binding) setValue(rv *reflect.Value, src bindSource) error {

	switch bind.Kind {

	case reflect.Bool:
		b := src.GetBool(bind.EnvVar)
		reflect.Indirect(*rv).SetBool(b)
		// log.Printf("[%s] value=%v", bind.Name, b)

	case reflect.String:

		s := src.GetString(bind.EnvVar)
		reflect.Indirect(*rv).SetString(s)
		// log.Printf("[%s] value=%s", bind.Name, s)

	// Special-case int64, as it may also be a duration.
	case reflect.Int64:

		// Try to parse value as an int, and if that fails, try
		// to parse it as a duration.
		s := src.GetString(bind.EnvVar)

		if _, err := strconv.ParseInt(s, 10, 64); err == nil {

			i := src.GetInt(bind.EnvVar)
			reflect.Indirect(*rv).SetInt(int64(i))

		} else {

			if d, err := time.ParseDuration(s); err == nil {
				reflect.Indirect(*rv).SetInt(int64(d))
			}

		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:

		i := src.GetInt(bind.EnvVar)
		reflect.Indirect(*rv).SetInt(int64(i))
		// log.Printf("[%s] value=%d", bind.Name, i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

		i := src.GetInt(bind.EnvVar)
		reflect.Indirect(*rv).SetUint(uint64(i))
		// log.Printf("[%s] value=%d", bind.Name, i)

	case reflect.Float32, reflect.Float64:

		n := src.GetFloat(bind.EnvVar)
		reflect.Indirect(*rv).SetFloat(n)
		// log.Printf("[%s] value=%f", bind.Name, n)

	default:
		return fmt.Errorf("unsupported type: %v", bind.Kind.String())
	}

	return nil

}

func extract(v interface{}) ([]*binding, error) {

	var binds []*binding

	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("need struct, not %s: %v", rv.Kind(), v)
	}

	typ := rv.Type()

	for i := 0; i < rv.NumField(); i++ {

		var (
			ok      bool
			field   reflect.StructField
			name    string
			tag     string
			varname string
		)

		field = typ.Field(i)
		name = field.Name

		if tag, ok = field.Tag.Lookup("env"); ok {
			if tag == "-" { // Ignore this field
				continue
			}
		}

		if !isBindable(field.Type.Kind()) {
			continue
		}

		if tag != "" {
			varname = tag
		} else {
			varname = EnvVarForField(name)
		}

		bind := &binding{
			Name:     name,
			EnvVar:   varname,
			FieldNum: i,
			Target:   v,
			Kind:     field.Type.Kind(),
		}
		binds = append(binds, bind)

	}

	return binds, nil
}

var bindableKinds = map[reflect.Kind]bool{
	reflect.String:  true,
	reflect.Bool:    true,
	reflect.Int:     true,
	reflect.Int8:    true,
	reflect.Int16:   true,
	reflect.Int32:   true,
	reflect.Int64:   true,
	reflect.Uint:    true,
	reflect.Uint8:   true,
	reflect.Uint16:  true,
	reflect.Uint32:  true,
	reflect.Uint64:  true,
	reflect.Float32: true,
	reflect.Float64: true,
}

func isBindable(kind reflect.Kind) bool {

	_, ok := bindableKinds[kind]
	return ok
}

func isZeroValue(rv reflect.Value) bool {

	if !rv.IsValid() {
		return true
	}

	typ := rv.Type()
	kind := typ.Kind()

	switch kind {

	case reflect.Ptr:
		return rv.IsNil()

	case reflect.String:
		return rv.String() == ""

	case reflect.Bool:
		return rv.Bool() == false

	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0.0

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0

	default:
		log.Printf("[pull] unknown kind: %v", kind)

	}

	return false
}

func isZeroString(s string, kind reflect.Kind) bool {

	if s == "" {
		return true
	}

	switch kind {
	case reflect.Bool:

		v, err := strconv.ParseBool(s)
		if err != nil {
			log.Printf("couldn't convert %s to bool: %v", s, err)
			return false
		}
		return v == false

	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			log.Printf("couldn't convert %s to float: %v", s, err)
			return false
		}
		return v == 0.0

	case reflect.Int64: // special-case int64: may also be duration

		// Try to parse as int64 first
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			return v == 0
		}

		// try to parse as a duration
		d, err := time.ParseDuration(s)
		if err != nil {
			log.Printf("couldn't convert %s to int64/duration: %s", s, err)
			return false
		}
		return d.Nanoseconds() == 0

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		v, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			log.Println(err)
			return false
		}
		return v == 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			log.Println(err)
			return false
		}
		return v == 0

	}

	return false
}

// EnvVarForField generates an environment variable name from a struct field name.
// This is documented to show how the automatic names are generated.
func EnvVarForField(name string) string {
	if !isCamelCase(name) {
		return strings.ToUpper(name)
	}
	return splitCamelCase(name)
}

func isCamelCase(s string) bool {
	if ok, _ := regexp.MatchString(".*[a-z]+[0-9_]*[A-Z]+.*", s); ok {
		return true
	}
	if ok, _ := regexp.MatchString("[A-Z][A-Z][A-Z]+[0-9_]*[a-z]+.*", s); ok {
		return true
	}
	return false
}

func splitCamelCase(name string) string {

	var (
		i     int
		re    *regexp.Regexp
		rest  string
		words []string
	)

	rest = name

	// Start with 3 or more capital letters.
	re = regexp.MustCompile("[A-Z]+([A-Z])[0-9]*[a-z]")
	for {
		if idx := re.FindStringSubmatchIndex(rest); idx != nil {
			i = idx[2]
			s := rest[:i]
			rest = rest[i:]
			words = append(words, s)
		} else {
			break
		}
	}

	re = regexp.MustCompile("[a-z][0-9_]*([A-Z])")
	for {

		if idx := re.FindStringSubmatchIndex(rest); idx != nil {
			i = idx[2]
			s := rest[:i]
			rest = rest[i:]
			words = append(words, s)
		} else {
			break
		}
	}

	if rest != "" {
		words = append(words, rest)
	}

	if len(words) > 0 {
		s := strings.ToUpper(strings.Join(words, "_"))
		return s
	}

	return ""
}
