package gop

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Style type
type Style struct {
	Set   string
	Unset string
}

var (
	// Bold style
	Bold = addStyle(1, 22)
	// Faint style
	Faint = addStyle(2, 22)
	// Italic style
	Italic = addStyle(3, 23)
	// Underline style
	Underline = addStyle(4, 24)
	// Blink style
	Blink = addStyle(5, 25)
	// RapidBlink style
	RapidBlink = addStyle(6, 26)
	// Invert style
	Invert = addStyle(7, 27)
	// Hide style
	Hide = addStyle(8, 28)
	// Strike style
	Strike = addStyle(9, 29)

	// Black color
	Black = addStyle(30, 39)
	// Red color
	Red = addStyle(31, 39)
	// Green color
	Green = addStyle(32, 39)
	// Yellow color
	Yellow = addStyle(33, 39)
	// Blue color
	Blue = addStyle(34, 39)
	// Magenta color
	Magenta = addStyle(35, 39)
	// Cyan color
	Cyan = addStyle(36, 39)
	// White color
	White = addStyle(37, 39)

	// BgBlack color
	BgBlack = addStyle(40, 49)
	// BgRed color
	BgRed = addStyle(41, 49)
	// BgGreen color
	BgGreen = addStyle(42, 49)
	// BgYellow color
	BgYellow = addStyle(43, 49)
	// BgBlue color
	BgBlue = addStyle(44, 49)
	// BgMagenta color
	BgMagenta = addStyle(45, 49)
	// BgCyan color
	BgCyan = addStyle(46, 49)
	// BgWhite color
	BgWhite = addStyle(47, 49)

	// None type
	None = Style{}
)

var regNewline = regexp.MustCompile(`\r?\n`)

// S is the shortcut for Stylize
func S(str string, styles ...Style) string {
	return Stylize(str, styles)
}

// Stylize string
func Stylize(str string, styles []Style) string {
	for _, s := range styles {
		str = stylize(s, str)
	}
	return str
}

func stylize(s Style, str string) string {
	if NoStyle || s == None {
		return str
	}

	newline := regNewline.FindString(str)

	lines := regNewline.Split(str, -1)
	out := []string{}

	for _, l := range lines {
		out = append(out, s.Set+l+s.Unset)
	}

	return strings.Join(out, newline)
}

// NoStyle respects https://no-color.org/ and "tput colors"
var NoStyle = func() bool {
	_, noColor := os.LookupEnv("NO_COLOR")

	b, _ := exec.Command("tput", "colors").CombinedOutput()
	n, _ := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 32)
	return noColor || n == 0
}()

// RegANSI token
var RegANSI = regexp.MustCompile(`\x1b\[\d+m`)

// StripANSI tokens
func StripANSI(str string) string {
	return RegANSI.ReplaceAllString(str, "")
}

// VisualizeANSI tokens
func VisualizeANSI(str string) string {
	reg := regexp.MustCompile(`\d+`)
	return RegANSI.ReplaceAllStringFunc(str, func(s string) string {
		return "<" + reg.FindString(s) + ">"
	})
}

// FixNestedStyle like
//     <red>1<blue>2<cyan>3</cyan>4</blue>5</red>
// into
//     <red>1</red><blue>2</blue><cyan>3</cyan><blue>4</blue><red>5</red>
func FixNestedStyle(s string) string {
	out := ""
	stack := []string{}
	i := 0
	l := 0
	r := 0

	for i < len(s) {
		loc := RegANSI.FindStringIndex(s[i:])
		if loc == nil {
			break
		}

		l, r = i+loc[0], i+loc[1]
		token := s[l:r]

		out += s[i:l]

		if len(stack) == 0 {
			stack = append(stack, token)
			out += token
		} else if token == styleMap[stack[len(stack)-1]].Unset {
			out += token
			stack = stack[:len(stack)-1]
			if len(stack) > 0 {
				out += stack[len(stack)-1]
			}
		} else {
			out += styleMap[stack[len(stack)-1]].Unset
			stack = append(stack, token)
			out += token
		}

		i = r
	}

	return out + s[i:]
}

var styleSetMap = map[string]Style{}

func addStyle(set, unset int) Style {
	s := Style{
		fmt.Sprintf("\x1b[%dm", set),
		fmt.Sprintf("\x1b[%dm", unset),
	}
	styleSetMap[s.Set] = s
	return s
}
