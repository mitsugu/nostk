package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
)

/*
publishMessage {{{
*/
func publishMessage(args []string, cc confClass) error {
	var s string
	var err error
	if len(args) < 3 {
		s, err = readStdIn()
		if err != nil {
			return errors.New("Not set text message")
		}
	} else {
		s = args[2]
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

	tgs := nostr.Tags{}
	if err := cc.setCustomEmoji(s, &tgs); err != nil {
		return err
	}

	ev := nostr.Event{
		PubKey:    pk,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindTextNote,
		Tags:      tgs,
		Content:   s,
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
