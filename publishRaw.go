package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/mattn/go-jsonpointer"
	"github.com/nbd-wtf/go-nostr"
	"github.com/yosuke-furukawa/json5/encoding/json5"
	"log"
	"runtime"
	"strings"
)

/*
publishRaw

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
	if fn.Name() == Main {
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
	} else {
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

	// calling Sign sets the event ID field and the event Sig field
	ev.Sign(sk)

	/*
		if 0 < len(strReason) {
			setContentWarning(strReason, &tgs)
		}

		if 0 < len(strPerson) {
			setPerson(strPerson, &tgs)
		}
	*/

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
mkEvent {{{
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
	case 10000: // publish mute list
	case 10001: // publish Pinned notes
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
	if ret := hasPrefixInTags(tgs); ret == true {
		return ev, errors.New("Include bech32 prefixes")
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

/*
getKind {{{
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

/*
getContent {{{
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
	if containsNsec1(strContent) || containsHsec1(strContent) {
		return "", errors.New("STRONGEST CAUTION!! : POSTS CONTAINING PRIVATE KEYS!! YOUR POST HAS BEEN REJECTED!!")
	}
	return strContent, nil
}

// }}}

/*
addTagsFromJson {{{
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
checkTags {{{
*/
func checkTags(kind int, tgs nostr.Tags) error {
	const tagNameIndex = 0
	list := GetChkTblMap()
	for i := range tgs {
		if result := contains(list[kind], tgs[i][tagNameIndex]); result != true {
			log.Printf("kind : %v, tagName : %v\n", kind, tgs[i][tagNameIndex])
			return errors.New("Inclusion of invalid tag in specified kind")
		}
	}
	return nil
}

var chkTblMap = map[int][]string{
	1:     {"content-warning", "client", "e", "emoji", "p", "q", "r", "t"},
	6:     {"e", "p"},
	10000: {"e", "p", "t", "word"},
	10001: {"e"},
	30315: {"d", "expiration", "r"},
}

func GetChkTblMap() map[int][]string {
	newMap := make(map[int][]string)
	for key, value := range chkTblMap {
		newMap[key] = append([]string(nil), value...)
	}
	return newMap
}
func sliceToMap(slice []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, v := range slice {
		m[v] = struct{}{} // struct{}{} はゼロサイズでメモリ効率が良い
	}
	return m
}
func contains(slice []string, target string) bool {
	m := sliceToMap(slice)
	_, exists := m[target]
	return exists
}

// }}}

/*
hasPrefixInTags {{{
*/
type StringPrefix []string

func (s StringPrefix) includes(target string) bool {
	for _, v := range s {
		if strings.Contains(target, v) {
			return true
		}
	}
	return false
}
func NewStringPrefix() StringPrefix {
	return StringPrefix{"npub", "nesc", "note", "nprofile", "nevent", "naddr", "nrelay"}
}
func hasPrefixInTags(tgs nostr.Tags) bool {
	prefixs := NewStringPrefix()
	for i := range tgs {
		for j := range tgs[i] {
			if ret := prefixs.includes(tgs[i][j]); ret == true {
				return true
			}
		}
	}
	return false
}

// }}}
