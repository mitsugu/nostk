package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
)

/*
	removeEvent {{{
		[infomation for develop]
		usage:
			nostk removeEvent <event_id>
		kind: 5
		content: reason text (must)
		tags [
			"e": event id (hex)
		]
*/
func removeEvent(args []string, cc confClass) error {
	var event_id string
	var kind string
	content := ""

	if len(args) < 4 || 5 < len(args) {
		return errors.New("Wrong number of parameters")
	}
	for i := range args {
		if i < 2 {
			continue
		}
		switch i {
		case 2: // event_id
			event_id = args[i]
			if 0 < len(event_id) && is64HexString(event_id) == false {
				if pref, err := getPrefixInString(event_id); err == nil {
					switch pref {
					case "note":
						if _, tmpEventId, err := toHex(event_id); err != nil {
							return err
						} else {
							event_id = tmpEventId.(string)
						}
					case "nevent":
						if _, tmpEventId, err := toHex(event_id); err != nil {
							return err
						} else {
							event_id = tmpEventId.(nostr.EventPointer).ID
						}
					default:
						return errors.New(fmt.Sprintf("Invalid id starting with %v", pref))
					}
				}
			}
		case 3: // kind
			kind = args[i]
		case 4: // content
			content = args[i]
			if containsNsec1(content) || containsHsec1(content) {
				return errors.New(fmt.Sprintf("STRONGEST CAUTION!! : POSTS CONTAINING PRIVATE KEYS!! YOUR POST HAS BEEN REJECTED!!"))
			}
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
	if err := cc.getRelayList(&rl, writeFlag); err != nil {
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
	t = append(t, "k")
	t = append(t, kind)
	tgs = append(tgs, t)

	ev := nostr.Event{
		PubKey:    pk,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindDeletion,
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
