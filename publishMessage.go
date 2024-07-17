package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	//"log"
	"regexp"
)

/*
publishMessage {{{
*/
func publishMessage(args []string, cc confClass) error {
	var s string
	strReason := ""
	var err error
	switch len(args) {
	case 2:
		s, err = readStdIn()
		if err != nil {
			return errors.New("Not set text message")
		}
	case 3:
		s = args[2]
	case 4:
		s = args[2]
		strReason = args[3]
	default:
		return errors.New("Invalid number of arguments")
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
