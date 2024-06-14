package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
)

/*
	emojiRreaction {{{
		[infomation for develop]
		usage:
			nostk emojiReaction <event_id> <public_key> <content>
				note:
					event_id: hex
					public_key: hex

		kind: 7
		content: emoji (include custom emoji short code)
		tags [
			"e": event id (hex)
			"p": pubkey (hex)
			"emoji": short_code, image_url (optional)
		]
*/
func emojiReaction(args []string, cc confClass) error {
	var event_id string
	var public_key string
	var content string

	if len(args) < 5 {
		return errors.New("Wrong number of parameters")
	}
	for i := range args {
		if i < 2 {
			continue
		}
		switch i {
		case 2: // event_id
			event_id = args[i]
		case 3: // public_key
			public_key = args[i]
		case 4: // content
			content = args[i]
		}
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
	if err := cc.getRelayList(&rl, readWriteFlag); err != nil {
		fmt.Println("Nothing relay list. Make a relay list.")
		return err
	}

	var t []string
	tgs := nostr.Tags{}
	if err := cc.setCustomEmoji(content, &tgs); err != nil {
		return err
	}
	t = nil
	t = append(t, "e")
	t = append(t, event_id)
	tgs = append(tgs, t)
	t = nil
	t = append(t, "p")
	t = append(t, public_key)
	tgs = append(tgs, t)
	t = nil
	t = append(t, "k")
	t = append(t, "1")
	tgs = append(tgs, t)

	ev := nostr.Event{
		PubKey:    pk,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindReaction,
		Tags:      tgs,
		Content:   content,
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
