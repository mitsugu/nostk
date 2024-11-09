package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"github.com/yosuke-furukawa/json5/encoding/json5"
)

const (
	Main         = "main.main"
	PubMessageTo = "main.publishMessageTo"
	tagName		 = 0
	reason		 = 1
	person		 = 1
)

/* publishMessage {{{
*/
func publishMessage(args []string, cc confClass) error {
	dataRawArg := RawArg{}
	switch len(args) {
	case 1:
		return errors.New("Not enough arguments")
	case 2:
		// Receive content from standard input
		if tmpContent, err := readStdIn(); err != nil {
			return errors.New("Not set text message")
		} else {
			dataRawArg.Kind = 1
			dataRawArg.Content = tmpContent
		}
	case 3, 4:
		if tmpArgJson, err := buildJson(args); err != nil {
			return err
		} else {
			err = json5.Unmarshal([]byte(tmpArgJson), &dataRawArg)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("Too meny argument")
	}

	// To prevent nsec leakage
	// Preventing hsec leaks is difficult in principle
	tagsFuncs := Tags{}
	if containsNsec1(dataRawArg.Content) {
		return errors.New(fmt.Sprintf("STRONGEST CAUTION!! : POSTS CONTAINING PRIVATE KEYS!! YOUR POST HAS BEEN REJECTED!!"))
	} else if ret := tagsFuncs.hasPrefix(dataRawArg.Tags, "nsec"); ret == true {
		return errors.New("STRONGEST CAUTION!! : POSTS CONTAINING PRIVATE KEYS!! YOUR POST HAS BEEN REJECTED!!")
	}

	// Convert bech32 string in tags to hex string
	if err := replaceBech32(dataRawArg.Kind, dataRawArg.Tags); err != nil {
		return err
	}

	sk, err := cc.load(cc.ConfData.Filename.Hsec)
	if err != nil {
		fmt.Println("Nothing key pair. Make key pair.")
		return err
	}
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return err
	}

	var rl []string
	if err := cc.getRelayList(&rl, writeFlag); err != nil {
		fmt.Println("Nothing relay list. Make a relay list.")
		return err
	}

	if err := cc.setCustomEmoji(dataRawArg.Content, &dataRawArg.Tags); err != nil {
		return err
	}

	// hashtags
	tmpstr, err := excludeHashtagsParsign(dataRawArg.Content)
	if err != nil {
		return err
	}
	setHashTags(tmpstr, &dataRawArg.Tags)

	ev := nostr.Event{
		PubKey:    pk,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindTextNote,
		Tags:      dataRawArg.Tags,
		Content:   dataRawArg.Content,
	}

	// calling Sign sets the event ID field and the event Sig field
	ev.Sign(sk)

	// publish the event to two relays
	ctx := context.Background()
	for _, url := range rl {
		relay, err := nostr.RelayConnect(ctx, url)
		if err != nil {
			fmt.Println(err)
			continue
		}
		err = relay.Publish(ctx, ev)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("published to %s\n", url)
	}

	return nil
}

// }}}

