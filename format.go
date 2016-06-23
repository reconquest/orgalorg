package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/kovetskiy/lorg"
	"github.com/seletskiy/hierr"
	"github.com/seletskiy/tplutil"
)

type (
	outputFormat int
)

const (
	outputFormatText outputFormat = iota
	outputFormatJSON
)

const (
	formatEscape = "\x1b"

	formatAttrForeground = "3"
	formatAttrBackground = "4"
	formatAttrDefault    = "9"

	formatAttrReset     = "0"
	formatAttrReverse   = "7"
	formatAttrNoReverse = "27"
	formatAttrBold      = "1"
	formatAttrNoBold    = "22"

	formatAttrForeground256 = "38;5"
	formatAttrBackground256 = "48;5"

	formatResetBlock = "{reset}"
)

var (
	formatCodeRegexp = regexp.MustCompile(formatEscape + `[^m]+`)
)

func getResetFormatSequence() string {
	return getFormatSequence(formatAttrReset)
}

func getBackgroundFormatSequence(color int) string {
	if color == 0 {
		return getFormatSequence(
			formatAttrBackground + formatAttrDefault,
		)
	}

	return getFormatSequence(formatAttrBackground256, fmt.Sprint(color))
}

func getForegroundFormatSequence(color int) string {
	if color == 0 {
		return getFormatSequence(
			formatAttrForeground + formatAttrDefault,
		)
	}

	return getFormatSequence(formatAttrForeground256, fmt.Sprint(color))
}

func getBoldFormatSequence() string {
	return getFormatSequence(formatAttrBold)
}

func getNoBoldFormatSequence() string {
	return getFormatSequence(formatAttrNoBold)
}

func getReverseFormatSequence() string {
	return getFormatSequence(formatAttrReverse)
}

func getNoReverseFormatSequence() string {
	return getFormatSequence(formatAttrNoReverse)
}

func getFormatSequence(attr ...string) string {
	if !isColorEnabled {
		return ""
	}

	return fmt.Sprintf("%s[%sm", formatEscape, strings.Join(attr, `;`))
}

func getLogPlaceholder(placeholder string) string {
	return fmt.Sprintf("${%s}", placeholder)
}

func getLogLevelFormatPlaceholder(
	level string,
	formatString string,
) (string, error) {
	format, err := compileFormat(formatString)
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't compile specified format string: '%s'`,
			formatString,
		)
	}

	formatCode, err := tplutil.ExecuteToString(format, nil)
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't execute specified format string: '%s'`,
			formatString,
		)
	}

	return fmt.Sprintf(
		"${color:%s:%s}",
		level,
		strings.Replace(formatCode, ":", "\\:", -1),
	), nil
}

func executeLogColorPlaceholder(level lorg.Level, value string) string {
	var (
		parts       = strings.SplitN(value, ":", 2)
		targetLevel = parts[0]
		format      = ""
	)

	if len(parts) > 1 {
		format = parts[1]
	}

	if targetLevel == strings.ToLower(level.String()) {
		return format
	}

	return ""
}

func compileFormat(format string) (*template.Template, error) {
	functions := map[string]interface{}{
		"bg":        getBackgroundFormatSequence,
		"fg":        getForegroundFormatSequence,
		"bold":      getBoldFormatSequence,
		"nobold":    getNoBoldFormatSequence,
		"reverse":   getReverseFormatSequence,
		"noreverse": getNoReverseFormatSequence,
		"reset":     getResetFormatSequence,

		"log":   getLogPlaceholder,
		"level": getLogLevelFormatPlaceholder,
	}

	return template.New("format").Delims("{", "}").Funcs(functions).Parse(
		format,
	)
}

func parseOutputFormat(
	args map[string]interface{},
) (outputFormat, bool, bool) {

	format := outputFormatText
	if args["--json"].(bool) {
		format = outputFormatJSON
	}

	isOutputOnTTY := terminal.IsTerminal(int(os.Stderr.Fd()))

	isColorEnabled := isOutputOnTTY

	if format != outputFormatText {
		isColorEnabled = false
	}

	if args["--no-colors"].(bool) {
		isColorEnabled = false
	}

	return format, isOutputOnTTY, isColorEnabled
}

func trimFormatCodes(input string) string {
	return formatCodeRegexp.ReplaceAllLiteralString(input, ``)
}

func parseStatusBarFormat(
	args map[string]interface{},
) (*template.Template, error) {
	var (
		theme = args["--status-format"].(string)
	)

	statusFormat := getStatusBarTheme(theme) + formatResetBlock

	format, err := compileFormat(statusFormat)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't compile status bar format template`,
		)
	}

	tracef("using status bar format: '%s'", statusFormat)

	return format, nil
}

func parseLogFormat(args map[string]interface{}) (*lorg.Format, error) {
	var (
		theme = args["--log-format"].(string)
	)

	logFormat := getLogTheme(theme) + formatResetBlock

	format, err := compileFormat(logFormat)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't compile log format template`,
		)
	}

	formatString, err := tplutil.ExecuteToString(format, nil)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't execute log format template`,
		)
	}

	tracef("using log format: '%#s'", logFormat)

	lorgFormat := lorg.NewFormat(formatString)
	lorgFormat.SetPlaceholder("color", executeLogColorPlaceholder)

	return lorgFormat, nil
}
