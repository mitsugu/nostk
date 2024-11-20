package main

import (
	"errors"
	"github.com/nbd-wtf/go-nostr"
	"testing"
)

func TestHasPrefix(t *testing.T) {
	funcs := Tags{}
	tests := []struct {
		tgs  nostr.Tags
		pref string
		ret  bool
	}{
		{
			tgs: nostr.Tags{
				{"p", "npub1czyqt7dafpysfye9w3agpp4rcrsxnt0tr8v0t8kyu66327maxstq5ckh7u"},
			},
			pref: "npub",
			ret:  true,
		}, {
			tgs: nostr.Tags{
				{"p", "c08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416"},
			},
			pref: "npub",
			ret:  false,
		}, {
			tgs: nostr.Tags{
				{"p", "nsec1czyqt7dafpysfye9w3agpp4hogehogehgoe0t8kyu66327maxstq5ckh7u"},
			},
			pref: "npub",
			ret:  false,
		}, {
			tgs: nostr.Tags{
				{"p", "nsec1czyqt7dafpysfye9w3agpp4hogehogehgoe0t8kyu66327maxstq5ckh7u"},
			},
			pref: "nsec",
			ret:  true,
		},
	}
	for _, tc := range tests {
		if ret := funcs.hasPrefix(tc.tgs, tc.pref); ret != tc.ret {
			t.Fatalf("got status: %v, Want status: %v", ret, tc.ret)
		}
	}
}

func TestNewSubCmdKindTbl(t *testing.T) {
	list := NewSubCmdKindTbl()
	t.Logf("【SubCmdKindTbl table】\n %v\n", list)
}

func TestHasSubcommandForSubCmdKindTbl(t *testing.T) {
	list := NewSubCmdKindTbl()
	tests := []struct {
		commandLine []string
		err         error
	}{
		{
			commandLine: []string{"nostk", "pubMessage", "テストコンテント"},
			err:         nil,
		},
		{
			commandLine: []string{"nostk", "pubMessage", "テストコンテント", "nsfwの理由"},
			err:         nil,
		},
		{
			commandLine: []string{"nostk", "pubMessageTo", "テストコンテント", "npub1czyqt7dafpysfye9w3agpp4rcrsxnt0tr8v0t8kyu66327maxstq5ckh7u"},
			err:         nil,
		},
		{
			commandLine: []string{"nostk", "pubMessageTo", "テストコンテント", "c08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416"},
			err:         nil,
		},
		{
			commandLine: []string{"nostk", "pubRaw", "テストコンテント", "c08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416"},
			err:         errors.New("Not supported subcommand pubRaw"),
		},
	}

	for _, tc := range tests {
		err := list.hasSubcommand(tc.commandLine)
		if err != nil && tc.err != nil {
			if err.Error() != tc.err.Error() {
				t.Fatalf("got error: %v, Want error: %v", err, tc.err)
			}
		}
	}
}

func TestGetForSubCmdKindTbl(t *testing.T) {
	list := NewSubCmdKindTbl()
	tests := []struct {
		commandLine []string
		err         error
		pos         int
	}{
		{
			commandLine: []string{"nostk", "pubMessage", "テストコンテント"},
			err:         nil,
			pos:         1,
		},
		{
			commandLine: []string{"nostk", "pubMessage", "テストコンテント", "nsfwの理由"},
			err:         nil,
			pos:         1,
		},
		{
			commandLine: []string{"nostk", "pubMessageTo", "テストコンテント", "npub1czyqt7dafpysfye9w3agpp4rcrsxnt0tr8v0t8kyu66327maxstq5ckh7u"},
			err:         nil,
			pos:         1,
		},
		{
			commandLine: []string{"nostk", "pubMessageTo", "テストコンテント", "c08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416"},
			err:         nil,
			pos:         1,
		},
	}

	for _, tc := range tests {
		ret, err := list.get(tc.commandLine)
		if err != nil && tc.err != nil {
			if tc.err != nil {
				t.Logf("got error code \"%v\" from %v subcommand", err, tc.commandLine[1])
				continue
			}
		}
		if tc.pos != 1 {
			t.Fatalf("got error: %v, Want error: 1", ret)
		}
	}
}

func TestErrGetForSubCmdKindTbl(t *testing.T) {
	list := NewSubCmdKindTbl()
	tests := []struct {
		commandLine []string
		err         error
		pos         int
	}{
		{
			commandLine: []string{"nostk", "pubRaw", "テストコンテント", "c08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416"},
			err:         errors.New("Not supported subcommand pubRaw"),
			pos:         1,
		},
	}

	for _, tc := range tests {
		ret, err := list.get(tc.commandLine)
		if err != nil && tc.err != nil {
			if tc.err == nil {
				t.Logf("got error code \"%v\" from %v subcommand", err, tc.commandLine[1])
				continue
			}
		}
		if tc.pos != 1 {
			t.Fatalf("got error: %v, Want error: 1", ret)
		}
	}
}

func TestNewConvArgsTagsTbl(t *testing.T) {
	list := NewConvArgsTagsTbl()
	t.Logf("【ConvArgsTagsTbl table】\n %v\n", list)
}

func TestHasSubcommand(t *testing.T) {
	tbl := NewConvArgsTagsTbl()
	tests := []struct {
		commandLine []string
		err         error
		pos         int
	}{
		{
			commandLine: []string{"nostk", "pubMessage", "テストコンテント"},
			err:         nil,
		},
		{
			commandLine: []string{"nostk", "pubMessage", "テストコンテント", "nsfwの理由"},
			err:         nil,
		},
		{
			commandLine: []string{"nostk", "pubMessageTo", "テストコンテント", "npub1czyqt7dafpysfye9w3agpp4rcrsxnt0tr8v0t8kyu66327maxstq5ckh7u"},
			err:         nil,
		},
		{
			commandLine: []string{"nostk", "pubMessageTo", "テストコンテント", "c08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416"},
			err:         nil,
		},
	}

	for _, tc := range tests {
		if err := tbl.hasSubcommand(tc.commandLine); err != tc.err {
			t.Fatalf("got error: %v, Want error: %v", err, tc.err)
		}
	}
}

func TestErrHasSubcommand(t *testing.T) {
	tbl := NewConvArgsTagsTbl()
	tests := []struct {
		commandLine []string
		err         error
		pos         int
	}{
		{
			commandLine: []string{"nostk", "pubRaw", "テストコンテント", "c08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416"},
			err:         errors.New("Not supported subcommand pubRaw"),
		},
	}

	for _, tc := range tests {
		if err := tbl.hasSubcommand(tc.commandLine); err == tc.err {
			t.Fatalf("got error: %v, Want error: %v", err, tc.err)
		}
	}
}

func TestBuildJson(t *testing.T) {
	tests := []struct {
		commandLine []string
		err         error
	}{
		{
			commandLine: []string{"nostk", "pubMessage", "テストコンテント"},
			err:         nil,
		},
		{
			commandLine: []string{"nostk", "pubMessage", "テストコンテント", "nsfwの理由"},
			err:         nil,
		},
		{
			commandLine: []string{"nostk", "pubMessageTo", "テストコンテント", "npub1czyqt7dafpysfye9w3agpp4rcrsxnt0tr8v0t8kyu66327maxstq5ckh7u"},
			err:         nil,
		},
		{
			commandLine: []string{"nostk", "pubMessageTo", "テストコンテント", "c08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416"},
			err:         nil,
		},
	}
	for i := range tests {
		strJson, err := buildJson(tests[i].commandLine)
		if err != nil {
			t.Fatalf("commandLine : %v, got error: %v, Want error: nil", tests[i].commandLine, err)
		} else {
			t.Logf("recieve json %v", strJson)
		}
	}
}
