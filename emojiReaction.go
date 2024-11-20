package main

import (
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"github.com/yosuke-furukawa/json5/encoding/json5"
	"regexp"
	"unicode"
	"unicode/utf8"
)

/* emojiRreaction {{{
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
	dataRawArg := RawArg{}
	tmpArgs := []string{
		"nostk",
		"emojiReaction",
	}

	tgs := nostr.Tags{}
	if len(args) < 6 {
		return errors.New("Wrong number of parameters")
	}

	dataRawArg.Kind = 7 // kind-07 (リアクション) を設定
	for i := range args {
		if i < 2 { // "nostr emojiReaction をスキップ
			continue
		}
		switch i {
		case 2: // tags に event_id を設定
			event_id := args[i]
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
			t := []string{}
			t = append(t, "e")
			t = append(t, event_id)
			tgs = append(tgs, t)
		case 3: // tags に public_key を設定
			public_key := args[i]
			if 0 < len(public_key) && is64HexString(public_key) == false {
				if pref, tmpPubkey, err := toHex(public_key); err != nil {
					return err
				} else if pref != "npub" {
					return errors.New(fmt.Sprintf("Invalid Pubkey %v", public_key))
				} else {
					public_key = tmpPubkey.(string)
				}
			}
			t := []string{}
			t = append(t, "p")
			t = append(t, public_key)
			tgs = append(tgs, t)
		case 4: // tags にリアクション対象の kind を設定
			t := []string{}
			t = append(t, "k")
			t = append(t, args[i])
			tgs = append(tgs, t)
		case 5: // content に "+"、"-"、絵文字またはカスタム絵文字の
				// ショートコードを設定
			dataRawArg.Content = args[i]
		}
	}

	// tags に emoji タグを追加する
	if err := cc.setCustomEmoji(dataRawArg.Content, &tgs); err != nil {
		return err
	}
	dataRawArg.Tags = tgs

	if tmp, err := json5.Marshal(dataRawArg); err != nil {
		return err
	} else {
		tmpArgs = append(tmpArgs, string(tmp))
	}
	if err := publishRaw(tmpArgs, cc); err != nil {
		return err
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

