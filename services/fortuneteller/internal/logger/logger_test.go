package logger

import (
	"expvar"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHook(t *testing.T) {
	Log.SetLevel(logrus.DebugLevel)
	Log.Debug("debug")
	Log.Info("info")
	Log.Warn("warn")
	Log.Error("error")

	v := expvar.Get("logs")
	require.NotNil(t, v)
	lines := v.String()
	assert.Contains(t, lines, "debug")
	assert.Contains(t, lines, "info")
	assert.Contains(t, lines, "warn")
	assert.Contains(t, lines, "error")
}
