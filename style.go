package main

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/kovetskiy/lorg"
	"github.com/seletskiy/hierr"
	"github.com/seletskiy/tplutil"
)

const (
	styleEscape = "\x1b"

	styleAttrForeground = "3"
	styleAttrBackground = "4"
	styleAttrDefault    = "9"

	styleAttrReset     = "0"
	styleAttrReverse   = "7"
	styleAttrNoReverse = "27"
	styleAttrBold      = "1"
	styleAttrNoBold    = "22"

	styleAttrForeground256 = "38;5"
	styleAttrBackground256 = "48;5"

	styleResetBlock = "{reset}"
)

var (
	styleCodeRegexp = regexp.MustCompile(styleEscape + `[^m]+`)
)

func getResetStyleSequence() string {
	return getStyleSequence(styleAttrReset)
}

func getBackgroundStyleSequence(color int) string {
	if color == 0 {
		return getStyleSequence(
			styleAttrBackground + styleAttrDefault,
		)
	}

	return getStyleSequence(styleAttrBackground256, fmt.Sprint(color))
}

func getForegroundStyleSequence(color int) string {
	if color == 0 {
		return getStyleSequence(
			styleAttrForeground + styleAttrDefault,
		)
	}

	return getStyleSequence(styleAttrForeground256, fmt.Sprint(color))
}

func getBoldStyleSequence() string {
	return getStyleSequence(styleAttrBold)
}

func getNoBoldStyleSequence() string {
	return getStyleSequence(styleAttrNoBold)
}

func getReverseStyleSequence() string {
	return getStyleSequence(styleAttrReverse)
}

func getNoReverseStyleSequence() string {
	return getStyleSequence(styleAttrNoReverse)
}

func getStyleSequence(attr ...string) string {
	if !isColorEnabled {
		return ""
	}

	return fmt.Sprintf("%s[%sm", styleEscape, strings.Join(attr, `;`))
}

func getLogPlaceholder(placeholder string) string {
	return fmt.Sprintf("${%s}", placeholder)
}

func getLogLevelStylePlaceholder(
	level string,
	styleString string,
) (string, error) {
	style, err := compileStyle(styleString)
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't compile specified style string: '%s'`,
			styleString,
		)
	}

	styleCode, err := tplutil.ExecuteToString(style, nil)
	if err != nil {
		return "", hierr.Errorf(
			err,
			`can't execute specified style string: '%s'`,
			styleString,
		)
	}

	return fmt.Sprintf(
		"${color:%s:%s}",
		level,
		strings.Replace(styleCode, ":", "\\:", -1),
	), nil
}

func executeLogColorPlaceholder(level lorg.Level, value string) string {
	var (
		parts       = strings.SplitN(value, ":", 2)
		targetLevel = parts[0]
		style       = ""
	)

	if len(parts) > 1 {
		style = parts[1]
	}

	if targetLevel == strings.ToLower(level.String()) {
		return style
	}

	return ""
}

func compileStyle(style string) (*template.Template, error) {
	functions := map[string]interface{}{
		"bg":        getBackgroundStyleSequence,
		"fg":        getForegroundStyleSequence,
		"bold":      getBoldStyleSequence,
		"nobold":    getNoBoldStyleSequence,
		"reverse":   getReverseStyleSequence,
		"noreverse": getNoReverseStyleSequence,
		"reset":     getResetStyleSequence,

		"log":   getLogPlaceholder,
		"level": getLogLevelStylePlaceholder,
	}

	return template.New("style").Delims("{", "}").Funcs(functions).Parse(
		style,
	)
}

func trimStyleCodes(input string) string {
	return styleCodeRegexp.ReplaceAllLiteralString(input, ``)
}

func getStatusBarStyle(theme string) (*template.Template, error) {
	statusStyle := getStatusBarTheme(theme) + styleResetBlock

	style, err := compileStyle(statusStyle)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't compile status bar style template`,
		)
	}

	tracef("using status bar style: '%s'", statusStyle)

	return style, nil
}

func getLoggerStyle(theme string) (*lorg.Format, error) {
	logStyle := getLogTheme(theme) + styleResetBlock

	style, err := compileStyle(logStyle)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't compile log style template`,
		)
	}

	styleString, err := tplutil.ExecuteToString(style, nil)
	if err != nil {
		return nil, hierr.Errorf(
			err,
			`can't execute log style template`,
		)
	}

	tracef("using log style: '%#s'", logStyle)

	lorgFormat := lorg.NewFormat(styleString)
	lorgFormat.SetPlaceholder("color", executeLogColorPlaceholder)

	return lorgFormat, nil
}

func parseTheme(target string, args map[string]interface{}) string {
	var (
		theme = args["--"+target+"-format"].(string)
		light = args["--light"].(bool)
		dark  = args["--dark"].(bool)
	)

	switch {
	case light:
		return themeLight

	case dark:
		return themeLight

	default:
		return theme
	}
}
