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

type statusBar struct {
	sync.Locker

	format *template.Template
	last   string

	data interface{}
}

func newStatusBar(format *template.Template) *statusBar {
	return &statusBar{
		format: format,
	}
}

func (bar *statusBar) Lock() {
	if bar.Locker != nil {
		bar.Locker.Lock()
	}
}

func (bar *statusBar) Unlock() {
	if bar.Locker != nil {
		bar.Locker.Unlock()
	}
}

func (bar *statusBar) Set(data interface{}) {
	bar.Lock()
	defer bar.Unlock()

	bar.data = data
}

func (bar *statusBar) SetLock(lock sync.Locker) {
	bar.Locker = lock
}

func (bar *statusBar) Clear(writer io.Writer) {
	bar.Lock()
	defer bar.Unlock()

	fmt.Fprint(writer, strings.Repeat(" ", len(bar.last))+"\r")

	bar.last = ""
}

func (bar *statusBar) Draw(writer io.Writer) {
	bar.Lock()
	defer bar.Unlock()

	buffer := &bytes.Buffer{}

	if bar.data == nil {
		return
	}

	err := bar.format.Execute(buffer, bar.data)
	if err != nil {
		errorf("%s", hierr.Errorf(
			err,
			`error during rendering status bar`,
		))
	}

	fmt.Fprintf(buffer, "\r")

	bar.last = trimStyleCodes(buffer.String())

	io.Copy(writer, buffer)
}
