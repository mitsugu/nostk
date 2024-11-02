package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

// type declaration
/*
const
*/
const (
	singleReadNo  = 1
	readWriteFlag = 0
	readFlag      = 1
	writeFlag     = 2
	unnsfw        = false
	nsfw          = true
	layout        = "2006/01/02 15:04:05 MST"
)

//

/*
for contact lists structure {{{
*/
type CONTACT struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

// }}}

/*
Log data structures {{{
*/
type CONTENTS struct {
	Date    string `json:"date"`
	PubKey  string `json:"pubkey"`
	Content string `json:"content"`
}
type NOSTRLOG struct {
	Id       string
	Contents CONTENTS
}

// }}}

/*
main {{{
*/
func main() {
	if len(os.Args) < 2 {
		dispHelp()
		os.Exit(0)
	}

	// load config.json
	var cc confClass
	if err := cc.existConfiguration(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	if err := cc.loadConfiguration(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help":
		dispHelp()
		os.Exit(0)
	case "--help":
		dispHelp()
		os.Exit(0)
	case "-h":
		dispHelp()
		os.Exit(0)
	case "init":
		if err := initEnv(cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "genkey":
		if err := genKey(cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "editRelays":
		if err := cc.edit(cc.ConfData.Filename.Relays); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "editProfile":
		if err := cc.edit(cc.ConfData.Filename.Profile); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "editEmoji":
		if err := cc.edit(cc.ConfData.Filename.Emoji); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "editContacts":
		if err := cc.edit(cc.ConfData.Filename.Contacts); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "pubProfile":
		if err := publishProfile(cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "pubRelays":
		if err := publishRelayList(cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "pubMessage":
		if err := publishMessage(os.Args, cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "pubMessageTo":
		if err := publishMessageTo(os.Args, cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "pubRaw":
		if err := publishRaw(os.Args, cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "catHome":
		if err := catHome(os.Args, cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "catSelf":
		if err := catSelf(os.Args, cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "catNSFW":
		if err := catNSFW(os.Args, cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "catEvent":
		if err := catEvent(os.Args, cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "emojiReaction":
		if err := emojiReaction(os.Args, cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "removeEvent":
		if err := removeEvent(os.Args, cc); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	default:
		log.Fatal(errors.New("Subcommand does not exist."))
		os.Exit(1)
	}
}

// }}}

/*
initEnv {{{
*/
func initEnv(cc confClass) error {
	// make skeleton of user profile
	if err := cc.create(cc.ConfData.Filename.Profile,
		`{
	"name":"",
	"display_name":"",
	"about":"",
	"website":"",
	"picture":"",
	"banner":"",
	"nip05":"",
	"lud16":""
}
`); err != nil {
		return err
	}
	// make skeleton of relay list
	if err := cc.create(cc.ConfData.Filename.Relays,
		`{
	"wws://" : {
		"read":true,
		"write":true
	}
}
`); err != nil {
		return err
	}
	// make skeleton of custom emoji list
	if err := cc.create(cc.ConfData.Filename.Emoji,
		`{
	"short code" : "image url"
}
`); err != nil {
		return err
	}
	// make skeleton of contact list
	if err := cc.create(cc.ConfData.Filename.Contacts,
		`{
	"hex pubkey" : {
		"url" : "",
		"name" : ""
	}
}
`); err != nil {
		return err
	}
	return nil
}

// }}}

/*
Generated Key Pair {{{
*/
func genKey(cc confClass) error {
	dirName, err := cc.getDir()
	if err != nil {
		return err
	}
	sk, pk, err := genHexKey()
	if err != nil {
		return err
	}
	ns, np, err := genNKey(sk, pk)
	if err != nil {
		return err
	}
	if err = cc.save(dirName, cc.ConfData.Filename.Hsec, sk); err != nil {
		return err
	}
	if err = cc.save(dirName, cc.ConfData.Filename.Hpub, pk); err != nil {
		return err
	}
	if err = cc.save(dirName, cc.ConfData.Filename.Nsec, ns); err != nil {
		return err
	}
	if err = cc.save(dirName, cc.ConfData.Filename.Npub, np); err != nil {
		return err
	}
	return nil
}

func genHexKey() (string, string, error) {
	sk := nostr.GeneratePrivateKey()
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return "", "", err
	}
	return sk, pk, nil
}

func genNKey(sk string, pk string) (string, string, error) {
	nsec, err := nip19.EncodePrivateKey(sk)
	if err != nil {
		return "", "", err
	}
	npub, err := nip19.EncodePublicKey(pk)
	if err != nil {
		return "", "", err
	}
	return nsec, npub, nil
}

// }}}

/*
publishProfile {{{
*/
func publishProfile(cc confClass) error {
	var rl []string
	s, err := cc.load(cc.ConfData.Filename.Profile)
	if err != nil {
		fmt.Println("Not found your profile. Use \"nostk init\" and \"nostk editProfile\".")
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

	if err := cc.getRelayList(&rl, writeFlag); err != nil {
		fmt.Println("Nothing relay list. Make a relay list.")
		return err
	}

	pr := strings.Replace(s, "\\n", "\n", -1)
	ev := nostr.Event{
		PubKey:    pk,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindProfileMetadata,
		Tags:      nil,
		Content:   string(pr),
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
publishRelayList {{{
*/
func publishRelayList(cc confClass) error {
	p := make(map[string]RwFlag)
	b, err := cc.load(cc.ConfData.Filename.Relays)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(b), &p)
	if err != nil {
		return err
	}
	const (
		cRead  = "read"
		cWrite = "write"
	)
	tags := nostr.Tags{}
	for i := range p {
		t := nostr.Tag{"r", i}
		if p[i].Read == true && p[i].Write == true {
		} else if p[i].Read == true {
			t = append(t, cRead)
		} else if p[i].Write == true {
			t = append(t, cWrite)
		}
		tags = append(tags, t)
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

	ev := nostr.Event{
		PubKey:    pk,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindRelayListMetadata,
		Tags:      tags,
		Content:   "",
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
		fmt.Printf("published relay list to %s\n", url)
	}

	return nil
}

// }}}

/*
readStdIn {{{
*/
func readStdIn() (string, error) {
	cn := make(chan string, 1)
	go func() {
		sc := bufio.NewScanner(os.Stdin)
		var buff bytes.Buffer
		for sc.Scan() {
			fmt.Fprintln(&buff, sc.Text())
		}
		cn <- buff.String()
	}()
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case text := <-cn:
		return text, nil
	case <-timer.C:
		return "", errors.New("Time out input from standard input")
	}
}

// }}}

/*
debugPrint {{{
*/
func startDebug(s string) {
	f, err := os.OpenFile(s, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)
	log.Println("start debug")
}

// }}}

/*
toHex {{{
*/
func toHex(str string) (string, any, error) {
	pref, data, err := nip19.Decode(str)
	if err != nil {
		return "", "", err
	}
	return pref, data, nil
}

// }}}

/*
is64HexString {{{
*/
func is64HexString(s string) bool {
	if len(s) != 64 {
		return false
	}
	match, _ := regexp.MatchString("^[a-fA-F0-9]{64}$", s)
	return match
}

// }}}

/*
getPrefixInString {{{
*/
func getPrefixInString(str string) (string, error) {
	pref, _, err := nip19.Decode(str)
	if err != nil {
		return "", err
	}
	return pref, nil
}
