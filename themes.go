package main

import (
	"fmt"

	"github.com/kovetskiy/lorg"
	"github.com/reconquest/colorgful"
	"github.com/reconquest/loreley"
)

const (
	themeDefault = `default`
	themeDark    = `dark`
	themeLight   = `light`
)

var (
	statusBarThemeTemplate = `{bg %d}{fg %d}` +
		`{bold}` +
		`{if eq .Phase "lock"}{bg %d}  LOCK{end}` +
		`{if eq .Phase "connect"}{bg %[3]d}  CONNECT{end}` +
		`{if eq .Phase "exec"}{bg %d}  EXEC{end}` +
		`{if eq .Phase "upload"}{bg %d}  UPLOAD{end}` +
		`{nobold} ` +
		`{from "" %d} ` +
		`{fg %d}{bold}{printf "%%4d" .Success}{nobold}{fg %d}` +
		`/{printf "%%4d" .Total} ` +
		`{if .Fails}{fg %d}✗ {.Fails}{end} ` +
		`{from "" %d}` +
		`{if eq .Phase "upload"}{fg %d} ` +
		`{printf "%%9s/%%s" .Written .Bytes} ` +
		`{end}`

	statusBarThemes = map[string]string{
		themeDark: fmt.Sprintf(
			statusBarThemeTemplate,
			99, 7, 22, 1, 25, 237, 46, 15, 214, -1, 140,
		),

		themeLight: fmt.Sprintf(
			statusBarThemeTemplate,
			99, 7, 22, 1, 64, 254, 106, 16, 9, -1, 140,
		),

		themeDefault: fmt.Sprintf(
			statusBarThemeTemplate,
			234, 255, 22, 1, 19, 245, 85, 255, 160, -1, 140,
		),
	}

	logFormat = `${time} ${level:[%s]:right:true} %s`
)

func getLoggerTheme(theme string) (lorg.Formatter, error) {
	switch theme {
	case "default":
		return colorgful.ApplyDefaultTheme(
			logFormat,
			colorgful.Default,
		)
	case "dark":
		return colorgful.ApplyDefaultTheme(
			logFormat,
			colorgful.Dark,
		)
	case "light":
		return colorgful.ApplyDefaultTheme(
			logFormat,
			colorgful.Light,
		)
	default:
		return colorgful.Format(theme)
	}
}

func getStatusBarTheme(theme string) (*loreley.Style, error) {
	if format, ok := statusBarThemes[theme]; ok {
		theme = format
	}

	style, err := loreley.CompileWithReset(theme, nil)
	if err != nil {
		return nil, err
	}

	return style, nil
}

func parseTheme(target string, args map[string]interface{}) string {
	var (
		theme = args["--"+target+"-format"].(string)
		light = args["--colors-light"].(bool)
		dark  = args["--colors-dark"].(bool)
	)

	switch {
	case light:
		return themeLight

	case dark:
		return themeDark

	default:
		return theme
	}
}
