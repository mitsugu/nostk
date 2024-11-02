package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	//"log"
	"regexp"
	"unicode"
	"unicode/utf8"
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
	var kind string
	var content string
	var err error

	if len(args) < 6 {
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
		case 3: // public_key
			public_key = args[i]
			if 0 < len(public_key) && is64HexString(public_key) == false {
				if pref, tmpPubkey, err := toHex(public_key); err != nil || pref != "npub" {
					return err
				} else {
					public_key = tmpPubkey.(string)
				}
			}
		case 4: // content
			kind = args[i]
		case 5: // content
			content = args[i]
		}
	}

	result := checkString(content)
	if result == false {
		return errors.New("The argument contains text that is neither an emoji nor a custom emoji.")
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
	t = append(t, kind)
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

/*
checkString {{{
*/
func checkString(s string) bool {
	if s == "+" || s == "-" {
		return true
	} else if utf8.RuneCountInString(s) == 1 {
		r, _ := utf8.DecodeRuneInString(s)
		return isEmoji(r)
	} else {
		return isShortCode(s)
	}
}

// }}}

/*
isEmoji {{{
*/
func isEmoji(r rune) bool {
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) ||
		(r >= 0x1F680 && r <= 0x1F6FF) ||
		(r >= 0x1F700 && r <= 0x1F77F) ||
		(r >= 0x1F780 && r <= 0x1F7FF) ||
		(r >= 0x1F800 && r <= 0x1F8FF) ||
		(r >= 0x1F900 && r <= 0x1F9FF) ||
		(r >= 0x1FA00 && r <= 0x1FA6F) ||
		(r >= 0x1FA70 && r <= 0x1FAFF) ||
		(r >= 0x2600 && r <= 0x26FF) ||
		(r >= 0x2700 && r <= 0x27BF) ||
		(r >= 0xFE00 && r <= 0xFE0F) ||
		unicode.Is(unicode.Sk, r)
}

// }}}

/*
isShortCode {{{
*/
func isShortCode(s string) bool {
	re := regexp.MustCompile(`^:.*:$`)
	return re.MatchString(s)
}

// }}}
