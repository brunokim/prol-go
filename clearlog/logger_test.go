package clearlog_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/brunokim/prol-go/clearlog"
)

type buffer struct {
	strings.Builder
}

func (*buffer) Close() error {
	return nil
}

type kv = clearlog.KV

const (
	contextPattern = `file="logger_test.go" line=[0-9]+ package="github.com/brunokim/prol-go/clearlog_test"`
)

func TestLogger(t *testing.T) {
	var buf buffer
	logger := clearlog.NewLogger(&buf)
	logger.Debug(kv{"message", "debug"})

	got := buf.String()
	t.Log(got)
	pattern := fmt.Sprintf(`level=10 %s func=TestLogger message=debug`, contextPattern)
	ok, err := regexp.MatchString(pattern, got)
	if err != nil {
		t.Fatalf("Error matching regex: %v", err)
	}
	if !ok {
		t.Errorf("Expected pattern:\n\t%s\ngot:\n\t%s", pattern, got)
	}
}
