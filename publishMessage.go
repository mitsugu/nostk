package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	//"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/yosuke-furukawa/json5/encoding/json5"
	//"log"
	"regexp"
	//"runtime"
)

const (
	Main         = "main.main"
	PubMessageTo = "main.publishMessageTo"
	tagName		 = 0
	reason		 = 1
	person		 = 1
)

/*
publishMessage
*/
func publishMessage(args []string, cc confClass) error {
	dataRawArg := RawArg{}
	switch len(args) {
	case 1:
		return errors.New("Not enough arguments")
	case 2:
		// Receive content from standard input
		tmpContent, err := readStdIn()
		if err != nil {
			return errors.New("Not set text message")
		}
		dataRawArg.Kind = 1
		dataRawArg.Content = tmpContent
	case 3:
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

	// Temporary workaround for false positives
	// in hexadecimal format pubkey
	// Although I am thinking of countermeasures,
	// it seems difficult to solve the problem.
	//if containsNsec1(s) || containsHsec1(s) || containsNsec1(strReason) || containsHsec1(strReason) {
	if containsNsec1(dataRawArg.Content) {
		return errors.New(fmt.Sprintf("STRONGEST CAUTION!! : POSTS CONTAINING PRIVATE KEYS!! YOUR POST HAS BEEN REJECTED!!"))
	}
	if 0 < len(dataRawArg.Tags) && dataRawArg.Tags[0][tagName] == "content-warning" {
		if containsNsec1(dataRawArg.Tags[0][reason]) {
			return errors.New(fmt.Sprintf("STRONGEST CAUTION!! : POSTS CONTAINING PRIVATE KEYS!! YOUR POST HAS BEEN REJECTED!!"))
		}
	} else if 0 < len(dataRawArg.Tags) && dataRawArg.Tags[0][tagName] == "p" {
		if containsNsec1(dataRawArg.Tags[0][person]) {
			return errors.New(fmt.Sprintf("STRONGEST CAUTION!! : POSTS CONTAINING PRIVATE KEYS!! YOUR POST HAS BEEN REJECTED!!"))
		}
	}
	if err := replaceBech32(dataRawArg.Kind, dataRawArg.Tags); err != nil {
		return err
	}

/*
	if 0 < len(strPerson) && is64HexString(strPerson) == false {
		if pref, err := getPrefixInString(strPerson); err == nil {
			switch pref {
			case "npub":
				if _, tmpPerson, err := toHex(strPerson); err != nil {
					return err
				} else {
					strPerson = tmpPerson.(string)
				}
			default:
				return errors.New("Invalid public key")
			}
		}
	}
*/

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

/*
excludeHashtagsParsign {{{
*/
func excludeHashtagsParsign(src string) (string, error) {
	const strexp = `(?:^|\s)([#﹟＃][^#﹟＃]\S*[#﹟＃]\S*)`
	re, err := regexp.Compile(strexp)
	if err != nil {
		return "", err
	}
	result := re.ReplaceAllString(src, "")
	return result, nil
}

// }}}

/*
setHashTags {{{
*/
func setHashTags(buf string, tgs *nostr.Tags) {
	const strexp = `(?:^|\s)([#﹟＃][^#\s﹟＃]+[^\s|$])`

	re := regexp.MustCompile(strexp)
	matches := re.FindAllString(buf, -1)
	for i := range matches {
		t := ExTag{}
		t.addTagName("t")
		rtmp := regexp.MustCompile(`[\s﹟＃#]`)
		result := rtmp.ReplaceAllString(matches[i], "")
		t.addTagValue(result)
		*tgs = append(*tgs, t.getNostrTag())
	}
}

// }}}

/*
setContentWarning {{{
*/
func setContentWarning(r string, tgs *nostr.Tags) {
	const CWTag = "content-warning"
	var t []string
	t = nil
	t = append(t, CWTag)
	t = append(t, r)
	*tgs = append(*tgs, t)
}

// }}}

/*
setPerson {{{
*/
func setPerson(p string, tgs *nostr.Tags) {
	const PTag = "p"
	var t []string
	t = nil
	t = append(t, PTag)
	t = append(t, p)
	*tgs = append(*tgs, t)
}

// }}}

/*
containsNsec1 {{{
*/
func containsNsec1(text string) bool {
	pattern := `nsec1[a-zA-Z0-9]{58}`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(text, -1)

	for _, match := range matches {
		alphanumericPart := match[5:]
		if !regexp.MustCompile(`nsec1`).MatchString(alphanumericPart) {
			return true
		}
	}

	return false
}

// }}}

/*
containsHsec1 {{{
*/
/*
func containsHsec1(text string) bool {
	pattern := `[a-zA-Z0-9]{64}`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(text, -1)

	for _, match := range matches {
		if _, err := nip19.EncodePrivateKey(match); err == nil {
			return true
		}
	}
	return false
}
*/

// }}}
