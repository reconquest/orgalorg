package main

import (
	"fmt"
)

const (
	themeDark  = `dark`
	themeLight = `light`
)

var (
	statusBarThemeTemplate = `{" "}{bg %d}{fg %d}` +
		`{if eq .Phase "` + statusBarPhaseExecuting + `"}` +
		`{bg %d}` +
		`{end}` +
		` {bold}{.Phase}{nobold} ` +
		`{fg %d}{reverse}{noreverse}{fg %d}{bg %d} ` +
		`{fg %d}{bold}{printf "%%4d" .Success}{nobold}{fg %d}` +
		`/{printf "%%4d" .Total} ` +
		`{if .Failures}{fg %d}(failed: {.Failures}){end} {fg %d}{bg %d}`

	statusBarThemes = map[string]string{
		themeDark: fmt.Sprintf(
			statusBarThemeTemplate,
			99, 7, 76, 237, 15, 237, 46, 15, 214, 237, 0,
		),

		themeLight: fmt.Sprintf(
			statusBarThemeTemplate,
			99, 7, 64, 254, 16, 254, 106, 16, 9, 254, 0,
		),
	}

	logThemeTemplate = `{level "error" "{fg %d}"}` +
		`{level "warning" "{fg %d}{bg %d}"}` +
		`{level "debug" "{fg %d}"}` +
		`{level "trace" "{fg %d}"}` +
		`* {log "time"} ` +
		`{level "error" "{bg %d}{bold}"}{log "level:[%%s]:right:true"}` +
		`{level "error" "{bg %d}"}{nobold} %%s`

	logThemes = map[string]string{
		themeDark: fmt.Sprintf(
			logThemeTemplate,
			1, 11, 0, 250, 243, 52, 0,
		),

		themeLight: fmt.Sprintf(
			logThemeTemplate,
			199, 172, 230, 240, 248, 220, 0,
		),
	}
)

func getStatusBarTheme(theme string) string {
	if format, ok := statusBarThemes[theme]; ok {
		return format
	}

	return theme
}

func getLogTheme(theme string) string {
	if format, ok := logThemes[theme]; ok {
		return format
	}

	return theme
}
