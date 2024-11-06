package main

import (
	"encoding/json"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	//"log"
)

const (
	subCommandName=1
)

/*
subcommand's kind getter
*/
type SubCmdKindTbl map[string]int
func NewSubCmdKindTbl() SubCmdKindTbl {
	return SubCmdKindTbl{
		"pubMessage": 1,
		"pubMessageTo": 1,
		//"emojiReaction": 6,
	}
}

func (r SubCmdKindTbl) hasSubcommand(args []string) error {
	if subCommandName >= len(args) {
		return fmt.Errorf("index %d out of range in args", subCommandName)
	}

	cmd := args[subCommandName]
	if _, exists := r[cmd]; !exists {
		return fmt.Errorf("Not supported subcommand %s", cmd)
	}
	return nil
}
func (r SubCmdKindTbl) get(args []string) (int, error) {
	if err := r.hasSubcommand(args); err != nil {
		return 0, err
	}
	return r[args[subCommandName]], nil
}

/*
subcommand's json builder
*/
type ConvArgsTagsTbl map[string]map[int][]string
func NewConvArgsTagsTbl() ConvArgsTagsTbl {
	return ConvArgsTagsTbl{
		"pubMessage":{
			3:{"content-warning",},
		},
		"pubMessageTo":{
			3:{"p"},
		},
		/*
		"emojiReaction":{
			6:{"e","p","k","emoji"},
		},
		*/
	}
}

func (r ConvArgsTagsTbl) hasSubcommand(args []string) error {
	if subCommandName >= len(args) {
		return fmt.Errorf("index %d out of range in args", subCommandName)
	}

	cmd := args[subCommandName]
	if _, exists := r[cmd]; !exists {
		return fmt.Errorf("Not supported subcommand %s", cmd)
	}
	return nil
}
type RawArg struct {
	Kind	int			`json:"kind"`
	Content	string		`json:"content"`
	Tags	nostr.Tags	`json:"tags"`
}
func buildJson(args []string) (string, error) {
	var ret	RawArg
	list := NewConvArgsTagsTbl()
	kindList := NewSubCmdKindTbl()
	if err := list.hasSubcommand(args); err != nil {
		return "", err
	}
	for i := range args {
		switch i {
			case 0:		// nostk
				continue
			case 1:		// subcommand name
				if tmpKind, err := kindList.get(args); err != nil {
					return "", err
				} else {
					ret.Kind = tmpKind
				}
			case 2:		// content
				ret.Content = args[i]
			default:	// tags
				if err := addTags(args, i, &ret.Tags); err != nil {
					return "", err
				}
		}
	}

	strJson, err := json.Marshal(ret)
	if err != nil {
		return "",err
	}
	return string(strJson), nil
}

func addTags(args []string, index int, tgs *nostr.Tags) error {
	t := []string{}
	list := NewConvArgsTagsTbl()

	if err := list.hasSubcommand(args); err != nil {
		return err
	}
	t = append(t, list[args[subCommandName]][index][0])
	t = append(t, args[index])
	*tgs = append(*tgs, t)
	return nil
}

