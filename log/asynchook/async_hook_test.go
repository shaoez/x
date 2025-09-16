package asynchook

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestAsyncHook(t *testing.T) {
	logger := logrus.New()

	var ok bool
	hook := New(1024, logrus.AllLevels, func(entry *logrus.Entry) error {
		ok = true
		return nil
	})

	logger.AddHook(hook)

	logger.Info("Done")

	hook.Close()

	if !ok {
		t.FailNow()
	}
}

func TestAsyncForHook(t *testing.T) {
	logger := logrus.New()

	testHook := new(test.Hook)
	asyncHook := NewWithHook(1024, testHook)
	logger.AddHook(asyncHook)

	logger.Info("Done")

	asyncHook.Close()

	if testHook.LastEntry().Message != "Done" {
		t.Error(testHook.Entries)
	}
}
