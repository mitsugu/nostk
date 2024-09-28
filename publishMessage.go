package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	//"log"
	"regexp"
	"runtime"
)

const (
	Main         = "main.main"
	PubMessageTo = "main.publishMessageTo"
)

/*
publishMessage {{{
*/
func publishMessage(args []string, cc confClass) error {
	var s string    // content string
	strReason := "" // content warning reason
	strPerson := "" // hpub of person
	var err error
	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	if fn.Name() == Main {
		switch len(args) {
		case 2:
			// Receive content from standard input
			s, err = readStdIn()
			if err != nil {
				return errors.New("Not set text message")
			}
		case 3:
			// Receive content from arguments
			s = args[2]
		case 4:
			// Receive content and content warning reason from arguments
			s = args[2]
			strReason = args[3]
		default:
			return errors.New("Invalid number of arguments")
		}
	} else if fn.Name() == PubMessageTo {
		if len(args) == 4 {
			s = args[2]
			strPerson = args[3]
		} else {
			return errors.New("Invalid number of arguments")
		}
	} else {
		return errors.New("Invalid pubMessage function call")
	}

	if containsNsec1(s) || containsHsec1(s) {
		fmt.Println("STRONGEST CAUTION!! : POSTS CONTAINING PRIVATE KEYS!! YOUR POST HAS BEEN REJECTED!!")
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

	tgs := nostr.Tags{}
	if err := cc.setCustomEmoji(s, &tgs); err != nil {
		return err
	}

	// hashtags
	tmpstr, err := excludeHashtagsParsign(s)
	if err != nil {
		return err
	}
	setHashTags(tmpstr, &tgs)

	if 0 < len(strReason) {
		setContentWarning(strReason, &tgs)
	}

	if 0 < len(strPerson) {
		setPerson(strPerson, &tgs)
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
setContentWarning {{{
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

// }}}
