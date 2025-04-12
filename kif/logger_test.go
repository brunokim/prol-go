package kif_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/brunokim/prol-go/kif"
	"github.com/google/go-cmp/cmp"
)

type buffer struct {
	strings.Builder
}

func (*buffer) Close() error {
	return nil
}

func TestYAMLLogger(t *testing.T) {
	var buf buffer
	logger := kif.NewLogger(&buf)
	logger.Debug(kif.KV{"message", "debug message"})
	logger.Info(kif.KV{"message", "info message"})
	lines := buf.String()
	t.Log(lines)

	tests := []struct {
		parts []string
	}{
		{[]string{"DEBUG", "'debug message'"}},
		{[]string{"INFO", "'info message'"}},
	}

	pattern := `level=(.*) file='logger_test.go' line=[0-9]+ package='github.com/brunokim/prol-go/kif_test' func=TestYAMLLogger message=(.*)`
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("Error compiling regex: %v", err)
	}
	for i, line := range strings.Split(lines, "\n") {
		if line == "" {
			continue
		}
		parts := re.FindStringSubmatch(line)
		if len(parts) == 0 {
			t.Fatalf("Expected pattern:\n\t%s\nline:\n\t%s", pattern, line)
		}
		test := tests[i]
		if diff := cmp.Diff(test.parts, parts[1:]); diff != "" {
			t.Errorf("Expected pattern:\n\t%s\nline:\n\t%s\ndiff:\n\t%s", pattern, line, diff)
		}
	}
}

func TestJSONLogger(t *testing.T) {
	var buf buffer
	logger := kif.NewLogger(&buf)
	logger.SetEncoder(kif.JSONEncoder)
	logger.Debug(kif.KV{"message", "debug message"})
	logger.Info(kif.KV{"message", "info message"})
	lines := buf.String()
	t.Log(lines)

	tests := []struct {
		parts []string
	}{
		{[]string{"DEBUG", "debug message"}},
		{[]string{"INFO", "info message"}},
	}

	pattern := `{"level":"(.*)","file":"logger_test.go","line":"[0-9]+","package":"github.com/brunokim/prol-go/kif_test","func":"TestJSONLogger","message":"(.*)"}`
	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("Error compiling regex: %v", err)
	}
	for i, line := range strings.Split(lines, "\n") {
		if line == "" {
			continue
		}
		parts := re.FindStringSubmatch(line)
		if len(parts) == 0 {
			t.Fatalf("Expected pattern:\n\t%s\nline:\n\t%s", pattern, line)
		}
		test := tests[i]
		if diff := cmp.Diff(test.parts, parts[1:]); diff != "" {
			t.Errorf("Expected pattern:\n\t%s\nline:\n\t%s\ndiff:\n\t%s", pattern, line, diff)
		}
	}
}
