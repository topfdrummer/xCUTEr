// Copyright (c) 2016 Niklas Wolber
// This file is licensed under the MIT license.
// See the LICENSE file for more information.

package main

import (
	"fmt"
	"testing"
)

func TestBuilderCommand(t *testing.T) {
	s := &stringVisitor{}
	c := s.Command(&command{Command: "first"}).(fmt.Stringer)

	want := "Execute \"first\""

	if got := c.String(); got != want {
		t.Errorf("want: %q, got: %q", want, got)
	}
}

func TestBuilderMultiple(t *testing.T) {
	s := &stringVisitor{}
	g := s.Sequential()
	g.Append(s.Command(&command{Command: "first"}))
	g.Append(s.Command(&command{Command: "second"}))

	want := "Sequential\n" +
		"├─ Execute \"first\"\n" +
		"└─ Execute \"second\""

	if got := g.(fmt.Stringer).String(); got != want {
		t.Errorf("want:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestBuilderEmptyMultiple(t *testing.T) {
	s := &stringVisitor{}
	g := s.Sequential()

	want := "Sequential"

	if got := g.(fmt.Stringer).String(); got != want {
		t.Errorf("want:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestBuilderNested(t *testing.T) {
	s := &stringVisitor{}
	g1 := s.Sequential()
	g2 := s.Parallel()
	g2.Append(s.Command(&command{Command: "first"}))
	g2.Append(s.Command(&command{Command: "third"}))
	g1.Append(g2.Wrap())
	g1.Append(s.Command(&command{Command: "second"}))

	want := "Sequential\n" +
		"├─ Parallel\n" +
		"│  ├─ Execute \"first\"\n" +
		"│  └─ Execute \"third\"\n" +
		"└─ Execute \"second\""

	if got := g1.(fmt.Stringer).String(); got != want {
		t.Errorf("want:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestBuilderNested3(t *testing.T) {
	s := &stringVisitor{}
	g1 := s.Sequential()
	g2 := s.Parallel()
	g2.Append(s.Command(&command{Command: "second"}))
	g2.Append(s.Command(&command{Command: "third"}))
	g1.Append(g2.Wrap())

	want := "Sequential\n" +
		"└─ Parallel\n" +
		"   ├─ Execute \"second\"\n" +
		"   └─ Execute \"third\""

	if got := g1.(fmt.Stringer).String(); got != want {
		t.Errorf("want:\n%s\n\ngot:\n%s", want, got)
	}
}

func TestBuilderNested4(t *testing.T) {
	s := &stringVisitor{}
	g1 := s.Sequential()
	g1.Append(s.Command(&command{Command: "first"}))
	g1.Append(s.Command(&command{Command: "second"}))
	g2 := s.Parallel()
	g2.Append(s.Command(&command{Command: "third"}))
	g1.Append(g2.Wrap())

	want := "Sequential\n" +
		"├─ Execute \"first\"\n" +
		"├─ Execute \"second\"\n" +
		"└─ Parallel\n" +
		"   └─ Execute \"third\""

	if got := g1.(fmt.Stringer).String(); got != want {
		t.Errorf("want:\n%s\n\ngot:\n%s", want, got)
	}
}
