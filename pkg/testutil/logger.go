package testutil

import (
	"bytes"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

func NewDiscardLogger() *logrus.Logger {
	l := logrus.New()
	l.Out = ioutil.Discard
	return l
}

func NewBufferLogger(buf *bytes.Buffer) *logrus.Logger {
	l := logrus.New()
	l.Out = buf
	return l
}
