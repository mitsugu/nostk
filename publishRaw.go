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

const (
	lengthHexData = 64
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
	/*
		if ret := hasPrefixInTags(tgs); ret == true {
			return ev, errors.New("Include bech32 prefixes")
		}
	*/

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
	1:     {"content-warning", "client", "e", "emoji", "expiration", "p", "q", "r", "t"},
	6:     {"e", "p"},
	10000: {"e", "p", "t", "word"},
	10001: {"e"},
	30315: {"d", "emoji", "expiration", "r"},
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
func (s StringPrefix) hasPrefix(target string) bool {
	for _, prefix := range s {
		if strings.HasPrefix(target, prefix) {
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

/*
replaceBech32 {{{
*/
func replaceBech32(kind int, tgs nostr.Tags) error {
	list := NewModifyBech32TagsList()
	bechList := NewModifyBech32List()
	const indexTagName = 0
	for i := range tgs {
		if res := list.has(kind, tgs[i][indexTagName]); res == true { // Check for presence of tag
			if err := bechList.convert(tgs[i]); err != nil { // Decoding Bech32 format ID or key or more
				return err
			}
		}
	}
	return nil
}

// Checker for target tag
//
//	Maintains a tag name list corresponding to kind and
//	uses the list to determine whether it is a target tag
type modifyBech32TagsTblMap map[int][]string

func NewModifyBech32TagsList() modifyBech32TagsTblMap {
	return modifyBech32TagsTblMap{
		1:     {"e", "p", "q"},
		6:     {"e", "p"},
		10000: {"e", "p"},
		10001: {"e"},
	}
}

// function has
//
//	map's member method of modifyBech32TagsTblMap
//	Returns a bool value indicating whether data corresponding
//	to the kind and tag name exists in modifyBech32TagsTblMap.
func (r modifyBech32TagsTblMap) has(kind int, tagName string) bool {
	tags, exists := r[kind]
	if !exists {
		return false
	}

	for _, tag := range tags {
		if tag == tagName {
			return true
		}
	}
	return false
}

// Bech32 converter
type modifyBech32TblMap map[string]map[int][]string

func NewModifyBech32List() modifyBech32TblMap {
	return modifyBech32TblMap{
		"e": {
			1: {"nevent", "note"},
			4: {"npub"},
		},
		"p": {
			1: {"npub"},
		},
		"q": {
			1: {"nevent", "note"},
			3: {"npub"},
		},
	}
}
func (r modifyBech32TblMap) exists(tagName string, pref string) bool {
	if innerMap, ok := r[tagName]; ok {
		for _, values := range innerMap {
			for _, item := range values {
				if item == pref {
					return true
				}
			}
		}
	}
	return false
}
func (r modifyBech32TblMap) convert(tag nostr.Tag) error {
	prefixs := NewStringPrefix()
	for i := range tag {
		if 0 == i { // skip tag name
			continue
		}
		if isHexString(tag[i]) == true { // hex ID or key
			if lengthHexData != len(tag[i]) {
				return errors.New("Detecting invalid hexadecimal data in tags")
			} else {
				continue
			}
		}
		if ret := prefixs.hasPrefix(tag[i]); ret != true { // check prefix
			continue
		}
		if _, tmpData, err := toHex(tag[i]); err != nil { // convert hex string for Besh32 ID or key
			return err
		} else {
			tag[i], _ = tmpData.(string)
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
