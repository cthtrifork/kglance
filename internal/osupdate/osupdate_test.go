package osupdate

import (
	"context"
	"errors"
	"testing"
)

const staged = `{"deployments":[
 {"booted":false,"staged":true,"version":"latest.20261111"},
 {"booted":true,"staged":false,"version":"latest.20261108"}]}`

func TestParse(t *testing.T) {
	cases := map[string]State{
		staged: {Staged: true, Version: "latest.20261111"},
		`{"deployments":[{"booted":true,"staged":false,"version":"x"}]}`: {},
		`{"deployments":[]}`: {},
		`{}`:                 {},
	}
	for in, want := range cases {
		got, err := parse([]byte(in))
		if err != nil || got != want {
			t.Errorf("parse(%s) = %+v, %v; want %+v", in, got, err, want)
		}
	}
}

func TestParseRejectsGarbage(t *testing.T) {
	if _, err := parse([]byte("not json")); err == nil {
		t.Fatal("want an error")
	}
}

func TestServiceReturnsZeroStateOnError(t *testing.T) {
	s := NewService(func(context.Context) (State, error) { return State{Staged: true}, errors.New("boom") })
	if got := s.State(); got != (State{}) {
		t.Fatalf("State() = %+v, want the zero state", got)
	}
	ok := NewService(func(context.Context) (State, error) { return State{Staged: true, Version: "v"}, nil })
	if got := ok.State(); !got.Staged || got.Version != "v" {
		t.Fatalf("State() = %+v", got)
	}
}
