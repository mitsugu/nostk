package main

import (
	"github.com/nbd-wtf/go-nostr"
	"testing"
)

func TestSetHashTags(t *testing.T) {
	tests := []struct {
		content string
		tgs     nostr.Tags
		length  int
	}{
		{
			content: "このテストではハッシュタグは入っていません",
			tgs: nostr.Tags{
				{"", ""},
			},
			length: 0,
		},
		{
			content: "このテストではハッシュタグは入っていません",
			tgs: nostr.Tags{
				{"t", "result1-1"},
			},
			length: 0,
		},
		{
			content: "このテストではハッシュタグは入っていません",
			tgs: nostr.Tags{
				{"t", "result1-1"},
				{"t", "result1-2"},
			},
			length: 0,
		},
		{
			content: "#このテストでは ハッシュタグを入れてみた",
			tgs: nostr.Tags{
				{"t", "このテストでは"},
			},
			length: 1,
		},
		{
			content: "このテストでは #ハッシュタグ を入れてみた",
			tgs: nostr.Tags{
				{"t", "ハッシュタグ"},
			},
			length: 1,
		},
	}
	for _, tc := range tests {
		tgs := nostr.Tags{}
		setHashTags(tc.content, &tgs)
		if tc.length != len(tgs) {
			t.Fatalf("conent : %v, length: %v, tgs : %#v", tc.content, len(tgs), tgs)
		}
	}
}

func TestContainsNsec1(t *testing.T) {
	tests := []struct {
		str    string
		result bool
	}{
		{
			str:    "私はみつぐ",
			result: false,
		},
		{
			str:    "私はnpub1czyqt7dafpysfye9w3agpp4rcrsxnt0tr8v0t8kyu66327maxstq5ckh7u",
			result: false,
		},
		{
			str:    "私はc08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416",
			result: false,
		},
		{
			str:    "私はnsec1czyqt7dafpysfye9w3agpp4rcrsxnt0tr8v0t8kyu66327maxstq5ckh7u",
			result: true,
		},
		{
			str:    "私はほげやまほげことの息子のnsec1czyqt7dafpysfye9w3agpp4rcrsxnt0tr8v0t8kyu66327maxstq5ckh7u",
			result: true,
		},
		{
			str:    "私はほげやまほげことの息子 https://nostrcheck.me/media/c08805f9bd4849049325747a8086a3c0e069adeb19d8f59ec4e6b5157b7d3416/c4186ee638cf8500313952779b14495288cea51b4c56d6892a99b0c2ffb38fa8.webp",
			result: false,
		},
	}
	for _, tc := range tests {
		res := containsNsec1(tc.str)
		if res != tc.result {
			t.Fatalf("str: %v, result: %v", tc.str, res)
		}
	}
}
