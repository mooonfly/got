package got

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// Testable interface. Usually, you use *testing.T as it.
type Testable interface {
	Name() string                            // same as testing.common.Name
	Skipped() bool                           // same as testing.common.Skipped
	Failed() bool                            // same as testing.common.Failed
	Cleanup(func())                          // same as testing.common.Cleanup
	FailNow()                                // same as testing.common.FailNow
	Fail()                                   // same as testing.common.Fail
	Helper()                                 // same as testing.common.Helper
	Logf(format string, args ...interface{}) // same as testing.common.Logf
	SkipNow()                                // same as testing.common.Skip
}

// G is the helper context, it hold some useful helpers to write tests
type G struct {
	Testable
	Assertions
	Helpers
}

// Options for Assertion
type Options struct {
	// Dump a value to human readable string
	Dump func(interface{}) string

	// Format keywords in the assertion message.
	// Such as color it for CLI output.
	Keyword func(string) string
}

// Defaults for Options
func Defaults() Options {
	return Options{
		func(v interface{}) string {
			if v == nil {
				return "nil"
			}

			s := fmt.Sprintf("%v", v)

			json := func() {
				buf := bytes.NewBuffer(nil)
				enc := json.NewEncoder(buf)
				enc.SetEscapeHTML(false)
				if enc.Encode(v) == nil {
					b, _ := json.Marshal(v)
					s = string(b)
				}
			}

			t := ""
			switch v.(type) {
			case string:
				json()
			case int:
				json()
			case bool:
				json()
			default:
				t = fmt.Sprintf(" <%v>", reflect.TypeOf(v))
			}

			return s + t
		},
		func(s string) string {
			return "⦗" + s + "⦘"
		},
	}
}

// New G
func New(t Testable) G {
	return NewWith(t, Defaults())
}

// NewWith G with options
func NewWith(t Testable, opts Options) G {
	return G{t, Assertions{t, opts.Dump, opts.Keyword}, Helpers{t}}
}
