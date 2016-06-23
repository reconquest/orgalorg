package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"text/template"

	"github.com/seletskiy/hierr"
)

type (
	statusBarPhase string
)

const (
	statusBarPhaseConnecting statusBarPhase = "CONNECTING"
	statusBarPhaseExecuting                 = "EVALUATING"
)

type statusBar struct {
	sync.Mutex

	Phase    statusBarPhase
	Total    int
	Failures int
	Success  int

	format *template.Template

	last string

	lock sync.Locker
}

func newStatusBar(format *template.Template) *statusBar {
	return &statusBar{
		format: format,
	}
}

func (bar *statusBar) SetPhase(phase statusBarPhase) {
	bar.Lock()
	defer bar.Unlock()

	bar.Phase = phase

	bar.Success = 0
}

func (bar *statusBar) SetTotal(total int) {
	bar.Lock()
	defer bar.Unlock()

	bar.Total = total
}

func (bar *statusBar) IncSuccess() {
	bar.Lock()
	defer bar.Unlock()

	bar.Success++
}

func (bar *statusBar) IncFailures() {
	bar.Lock()
	defer bar.Unlock()

	bar.Failures++
}

func (bar *statusBar) SetOutputLock(lock sync.Locker) {
	bar.lock = lock
}

func (bar *statusBar) Clear(writer io.Writer) {
	bar.Lock()
	defer bar.Unlock()

	if bar.lock != nil {
		bar.lock.Lock()
		defer bar.lock.Unlock()
	}

	fmt.Fprint(writer, strings.Repeat(" ", len(bar.last))+"\r")

	bar.last = ""
}

func (bar *statusBar) Draw(writer io.Writer) {
	bar.Lock()
	defer bar.Unlock()

	if bar.lock != nil {
		bar.lock.Lock()
		defer bar.lock.Unlock()
	}

	buffer := &bytes.Buffer{}

	if bar.Phase == "" {
		return
	}

	err := bar.format.Execute(buffer, bar)
	if err != nil {
		errorf("%s", hierr.Errorf(
			err,
			`error during rendering status bar`,
		))
	}

	fmt.Fprintf(buffer, "\r")

	bar.last = trimFormatCodes(buffer.String())

	io.Copy(writer, buffer)
}
