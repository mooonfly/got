package gop_test

import (
	"encoding/base64"
	"fmt"
	"go/parser"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/ysmood/got"
	"github.com/ysmood/got/lib/gop"
)

func TestTokenize(t *testing.T) {
	g := got.New(t)
	ref := "test"
	timeStamp, _ := time.Parse(time.RFC3339Nano, "2021-08-28T08:36:36.807908+08:00")

	v := []interface{}{
		nil,
		[]interface{}{true, false, uintptr(0x17), float32(100.121111133)},
		true, 10, int8(2), int32(100),
		float64(100.121111133),
		complex64(1 + 2i), complex128(1 + 2i),
		[3]int{1, 2},
		make(chan int),
		make(chan string, 3),
		func(string) int { return 10 },
		map[interface{}]interface{}{
			"test": 10,
			"a":    1,
		},
		unsafe.Pointer(&ref),
		struct {
			Int int
			str string
			M   map[int]int
		}{10, "ok", map[int]int{1: 0x20}},
		[]byte("aa\xe2"),
		[]byte("bytes\n\tbytes"),
		byte('a'),
		byte(1),
		'天',
		"\ntest",
		&ref,
		(*struct{ Int int })(nil),
		&struct{ Int int }{},
		&map[int]int{1: 2, 3: 4},
		&[]int{1, 2},
		&[2]int{1, 2},
		&[]byte{1, 2},
		timeStamp,
		time.Hour,
		`{"a": 1}`,
		[]byte(`{"a": 1}`),
	}

	out := gop.StripColor(gop.F(v))

	expected := `
gop.Arr/* len=33 cap=33 */{
    nil,
    gop.Arr/* len=4 cap=4 */{
        true,
        false,
        uintptr(23),
        float32(100.12111),
    },
    true,
    10,
    int8(2),
    'd',
    float64(100.121111133),
    complex64(1+2i),
    1+2i,
    [3]int{
        1,
        2,
        0,
    },
    make(chan int),
    make(chan string, 3),
    (func(string) int)(nil),
    map[interface {}]interface {}/* len=2 */{
        "a": 1,
        "test": 10,
    },
    unsafe.Pointer(uintptr(` + fmt.Sprintf("%v", &ref) + `)),
    struct { Int int; str string; M map[int]int }/* len=3 */{
        Int: 10,
        str: "ok",
        M: map[int]int{
            1: 32,
        },
    },
    gop.Base64("YWHi")/* len=3 */,
    []byte("" +
        "bytes\n" +
        "\tbytes")/* len=12 */,
    byte('a'),
    byte(0x1),
    '天',
    "" +
        "\n" +
        "test"/* len=5 */,
    gop.Ptr("test").(*string),
    (*struct { Int int })(nil),
    &struct { Int int }{
        Int: 0,
    },
    &map[int]int/* len=2 */{
        1: 2,
        3: 4,
    },
    &[]int/* len=2 cap=2 */{
        1,
        2,
    },
    &[2]int{
        1,
        2,
    },
    gop.Ptr([]byte("\x01\x02")/* len=2 */).(*[]uint8),
    gop.Time("2021-08-28T08:36:36.807908+08:00"),
    gop.Duration("1h0m0s"),
    gop.JSONStr(gop.Obj{
        "a": float64(1),
    }, "{\"a\": 1}"),
    gop.JSONBytes(gop.Obj{
        "a": float64(1),
    }, "{\"a\": 1}"),
}`

	g.Eq(out, expected[1:])

	g.Nil(parser.ParseExpr(out))
}

type A struct {
	Int int
	B   *B
}

type B struct {
	s string
	a *A
}

func TestCircularRef(t *testing.T) {
	g := got.New(t)
	a := A{Int: 10}
	b := B{"test", &a}
	a.B = &b

	g.Eq(gop.StripColor(gop.F(a)), ""+
		"gop_test.A/* len=2 */{\n"+
		"    Int: 10,\n"+
		"    B: &gop_test.B/* len=2 */{\n"+
		"        s: \"test\",\n"+
		"        a: &gop_test.A/* len=2 */{\n"+
		"            Int: 10,\n"+
		"            B: gop.Circular(\"B\").(*gop_test.B),\n"+
		"        },\n"+
		"    },\n"+
		"}")
}

func TestCircularMap(t *testing.T) {
	g := got.New(t)
	a := map[int]interface{}{}
	a[0] = a

	ts := gop.Tokenize(a)

	g.Eq(gop.Format(ts, gop.NoTheme), ""+
		"map[int]interface {}{\n"+
		"    0: gop.Circular().(map[int]interface {}),\n"+
		"}")
}

func TestCircularSlice(t *testing.T) {
	g := got.New(t)
	a := [][]interface{}{{nil}, {nil}}
	a[0][0] = a[1]
	a[1][0] = a[0][0]

	ts := gop.Tokenize(a)

	g.Eq(gop.Format(ts, gop.NoTheme), ""+
		"[][]interface {}/* len=2 cap=2 */{\n"+
		"    gop.Arr/* len=1 cap=1 */{\n"+
		"        gop.Arr/* len=1 cap=1 */{\n"+
		"            gop.Circular(0, 0).(gop.Arr),\n"+
		"        },\n"+
		"    },\n"+
		"    gop.Circular(0, 0).(gop.Arr),\n"+
		"}")
}

func TestPlain(t *testing.T) {
	g := got.New(t)
	g.Eq(gop.Plain(10), "10")
}

func TestP(t *testing.T) {
	gop.Stdout = ioutil.Discard
	_ = gop.P("test")
	gop.Stdout = os.Stdout
}

func TestConvertors(t *testing.T) {
	g := got.New(t)
	g.Nil(gop.Circular(""))

	s := g.RandStr(8)
	g.Eq(gop.Ptr(s).(*string), &s)

	bs := base64.StdEncoding.EncodeToString([]byte(s))

	g.Eq(gop.Base64(bs), []byte(s))
	now := time.Now()
	g.Eq(gop.Time(now.Format(time.RFC3339Nano)), now)
	g.Eq(gop.Duration("10m"), 10*time.Minute)

	g.Eq(gop.JSONStr(nil, "[1, 2]"), "[1, 2]")
	g.Eq(gop.JSONBytes(nil, "[1, 2]"), []byte("[1, 2]"))
}

func TestGetPrivateFieldErr(t *testing.T) {
	g := got.New(t)
	g.Panic(func() {
		gop.GetPrivateField(reflect.ValueOf(1), 0)
	})
}

func TestLab(t *testing.T) {
}
