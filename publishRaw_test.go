package main

import (
	"github.com/nbd-wtf/go-nostr"
	"testing"
)

/* コピれ {{{
func TestFunc(t *testing.T) {
	list := NewSubCmdKindTbl()
	tests := []struct{
		commandLine []string
		err			error
	}{
		{
			commandLine:	[]string{"nostk","pubMessage","テストコンテント"},
			err:			nil,
		},
		{
			commandLine:	[]string{"nostk","pubMessage","テストコンテント","nsfwの理由"},
			err:			nil,
		},
		{
			commandLine:	[]string{"nostk","pubMessageTo","テストコンテント","npub1czyqt7dafpysfye9w3agpp4rcrsxnt0tr8v0t8kyu66327maxstq5ckh7u"},
			err:			nil,
		},
		{
			commandLine:	[]string{"nostk","pubMessageTo","テストコンテント","c08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416"},
			err:			nil,
		},
		{
			commandLine:	[]string{"nostk","pubRaw","テストコンテント","c08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416"},
			err:			errors.New("Not supported subcommand pubRaw"),
		},
	}

	for _, tc := range tests {
		strJson, err := buildJson(tests[i].commandLine)
		if err != nil {
			t.Fatalf("commandLine : %v, got error: %v, Want error: nil",tests[i].commandLine,  err)
		} else {
			t.Logf("recieve json %v",strJson)
		}
	}
}
}}} */

func TestReplaceBech32(t *testing.T) {
	tests := []struct {
		Kind int
		Tgs  nostr.Tags
		Err  error
	}{
		{
			Kind: 1,
			Tgs: nostr.Tags{
				{"p", "npub1czyqt7dafpysfye9w3agpp4rcrsxnt0tr8v0t8kyu66327maxstq5ckh7u"},
			},
			Err: nil,
		},
		{
			Kind: 1,
			Tgs: nostr.Tags{
				{"p", "c08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416"},
			},
			Err: nil,
		},
		{
			Kind: 1,
			Tgs: nostr.Tags{
				{"p", "nsec149qult2em55mwvm07m3fyvu0hdkfyc8c5s3dqetsu0jcn2psavwqvyvxwu"},
			},
			Err: nil,
		},
	}

	for _, tc := range tests {
		err := replaceBech32(tc.Kind, tc.Tgs)
		if err != tc.Err {
			t.Logf("test kind: %v, test tags: %v", tc.Kind, tc.Tgs)
			t.Fatalf("got error: %v, Want error: %v", err, tc.Err)
		}
	}
}
