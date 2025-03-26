package prol_test

import (
	"testing"

	"github.com/brunokim/prol-go/prol"
)

func TestEval(t *testing.T) {
	tests := []struct {
		name string
		expr prol.Term
		want prol.Term
	}{
		{"Int", int_(10), int_(10)},
		{"Sub", s("-", int_(10), int_(2)), int_(8)},
		{"Sum", s("+", int_(10), int_(2)), int_(12)},
		{"Neg", s("-", int_(10)), int_(-10)},
		{"Pos", s("+", int_(10)), int_(+10)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := prol.Eval(test.expr)
			if err != nil {
				t.Errorf("want nil err, got %v", err)
				return
			}
			if test.want != got {
				t.Errorf("want %v, got %v", test.want, got)
			}
		})
	}
}
