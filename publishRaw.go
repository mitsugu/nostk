package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/mattn/go-jsonpointer"
	"github.com/nbd-wtf/go-nostr"
	"github.com/yosuke-furukawa/json5/encoding/json5"
	//"log"
	"runtime"
)

const (
	Main          = "main.main"
	PubMessage    = "main.publishMessage"
	PubMessageTo  = "main.publishMessageTo"
	EmojiReaction = "main.emojiReaction"
	lengthHexData = 64
	indexTagName  = 0
)

/*
	 publishRaw {{{

		Data format example:
			{
				"kind": kind string,
				"content": content string,
				"tags": [
					"d": status's type string,
						 general or music
					"expiration": expiration's number (optional)
					"r": []relay's url, (optional)
						 ex)
						 	https://nostr.world
							spotify:search:Intergalatic%20-%20Beastie%20Boys
				]
			}
*/
func publishRaw(args []string, cc confClass) error {
	var err error
	var strjson string

	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	switch fn.Name() {
	case Main:
		switch len(args) {
		case 2:
			// Receive content from standard input
			strjson, err = readStdIn()
			if err != nil {
				return errors.New("Not set json data")
			}
		case 3:
			// Receive content from arguments
			strjson = args[2]
		default:
			return errors.New("Invalid pubRaw subcommand argument")
		}
	case PubMessage, EmojiReaction:
		strjson = args[2]
	default:
		return errors.New("pubRaw function call from illegal function")
	}

	var objJson interface{}
	if err := json5.Unmarshal([]byte(strjson), &objJson); err != nil {
		return err
	}

	ev, err := mkEvent(objJson, cc)
	if err != nil {
		return err
	}

	var rl []string
	if err := cc.getRelayList(&rl, writeFlag); err != nil {
		return err
	}

	sk, err := cc.load(cc.ConfData.Filename.Hsec)
	if err != nil {
		fmt.Println("Nothing key pair. Make key pair.")
		return err
	}

	ev.Sign(sk)

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

/* mkEvent {{{
 */
func mkEvent(pJson interface{}, cc confClass) (nostr.Event, error) {
	var ev nostr.Event
	kind, err := getKind(pJson)
	if err != nil {
		return ev, err
	}

	content, err := getContent(pJson)
	if err != nil {
		return ev, err
	}

	switch kind {
	case 1: // publish kind 1 message
	case 6: // publish Reposts
  case 7: // publish emojiReaction
	case 10000: // publish mute list
	case 10001: // publish Pinned notes
	case 30030: // publish custom emoji list
	case 30315: // publish status
	default:
		return ev, errors.New("not yet suppoted kind")
	}

	tgs := nostr.Tags{}
	// custom emojis
	if err := cc.setCustomEmoji(content, &tgs); err != nil {
		return ev, err
	}

	// hashtags
	tmpstr, err := excludeHashtagsParsign(content)
	if err != nil {
		return ev, err
	}
	setHashTags(tmpstr, &tgs)

	addTagsFromJson(pJson, &tgs)

	if err := checkTags(kind, tgs); err != nil {
		return ev, err
	}
	tagsFuncs := Tags{}
	if ret := tagsFuncs.hasPrefix(tgs, "nsec"); ret == true {
		return ev, errors.New("STRONGEST CAUTION!! : POSTS CONTAINING PRIVATE KEYS!! YOUR POST HAS BEEN REJECTED!!")
	}

	if err := replaceBech32(kind, tgs); err != nil {
		return ev, err
	}

	sk, err := cc.load(cc.ConfData.Filename.Hsec)
	if err != nil {
		fmt.Println("Nothing key pair. Make key pair.")
		return ev, err
	}
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return ev, err
	}

	ev = nostr.Event{
		PubKey:    pk,
		CreatedAt: nostr.Now(),
		Kind:      kind,
		Tags:      tgs,
		Content:   content,
	}

	return ev, nil
}

// }}}

/* getKind {{{
 */
func getKind(pJson interface{}) (int, error) {
	kind, err := jsonpointer.Get(pJson, "/kind")
	if err != nil {
		return -1, err
	}

	var intKind int
	if kindValue, ok := kind.(float64); ok {
		intKind = int(kindValue)
	} else {
		return -1, errors.New("Failed to convert 'kind' value to int64")
	}
	return intKind, nil
}

// }}}

/* getContent {{{
 */
func getContent(pJson interface{}) (string, error) {
	content, err := jsonpointer.Get(pJson, "/content")
	if err != nil {
		return "", err
	}

	var strContent string
	if contentValue, ok := content.(string); ok {
		strContent = string(contentValue)
	} else {
		return "", errors.New("Failed to convert 'kind' value to int64")
	}
	if containsNsec1(strContent) {
		return "", errors.New("STRONGEST CAUTION!! : POSTS CONTAINING PRIVATE KEYS!! YOUR POST HAS BEEN REJECTED!!")
	}
	return strContent, nil
}

// }}}

/* addTagsFromJson {{{
 */
func addTagsFromJson(pJson interface{}, tgs *nostr.Tags) error {
	jsonMap, ok := pJson.(map[string]interface{})
	if !ok {
		return errors.New("pJson is not a valid map")
	}

	if tags, ok := jsonMap["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagArray, ok := tag.([]interface{}); ok {
				var t []string
				t = nil
				for _, elm := range tagArray {
					t = append(t, elm.(string))
				}
				*tgs = append(*tgs, t)
			}
		}
	}
	return nil
}

// }}}

/*
	Error check function table {{{

IMPLEMENTED IN A FUTURE MAJOR VERSION
*/
var defChkTblMap = map[string]map[int]string{
	"content-warning": {1: "checkSeckey"},
	"client":          {1: "checkEmpty", 2: "noncheck", 3: "noncheck"},
	"e":               {1: "checkEventId", 2: "noncheck", 3: "checkMarker", 4: "checkPubkey"},
}

// }}}

// vim: set ts=2 sw=2 et:
