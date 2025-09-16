package iohook

import (
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

type testWriter func([]byte)

func (w testWriter) Write(data []byte) (n int, err error) {
	w(data)
	return len(data), nil
}

func TestWriterHook(t *testing.T) {
	logger := logrus.New()

	logger.AddHook(New(testWriter(func(data []byte) {
		if !strings.Contains(string(data), "TEST LOG") {
			t.Error()
		}
	}), &logrus.JSONFormatter{}))

	logger.Info("TEST LOG")

}
