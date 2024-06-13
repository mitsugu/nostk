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
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	"github.com/yosuke-furukawa/json5/encoding/json5"
)

// type declaration {{{
const (
	secretDir     = ".nostk"
	hsec          = ".hsec"
	nsec          = ".nsec"
	hpub          = ".hpub"
	npub          = ".npub"
	relays        = "relays.json"
	profile       = "profile.json"
	emoji         = "customemoji.json"
	contacts      = "contacts.json"
	waitTime      = 15
	defReadNo     = 20
	singleReadNo  = 1
	readWriteFlag = 0
	readFlag      = 1
	writeFlag     = 2
	unnsfw        = false
	nsfw          = true
)

type ProfileMetadata struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	About       string `json:"about"`
	Website     string `json:"website"`
	Picture     string `json:"picture"`
	Banner      string `json:"banner"`
	NIP05       string `json:"nip05"`
	LUD16       string `json:"lud16"`
}

type RwFlag struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
}

type CONTACT struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

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
	switch os.Args[1] {
	case "help":
		dispHelp()
	case "--help":
		dispHelp()
	case "-h":
		dispHelp()
	case "init":
		if err := initEnv(); err != nil {
			log.Fatal(err)
		}
	case "genkey":
		if err := genKey(); err != nil {
			log.Fatal(err)
		}
	case "editRelays":
		if err := edit(relays); err != nil {
			log.Fatal(err)
		}
	case "editProfile":
		if err := edit(profile); err != nil {
			log.Fatal(err)
		}
	case "editEmoji":
		if err := edit(emoji); err != nil {
			log.Fatal(err)
		}
	case "editContacts":
		log.Println("contacts")
		if err := edit(contacts); err != nil {
			log.Fatal(err)
		}
	case "pubProfile":
		if err := publishProfile(); err != nil {
			log.Fatal(err)
		}
	case "pubRelays":
		if err := publishRelayList(); err != nil {
			log.Fatal(err)
		}
	case "pubMessage":
		if err := publishMessage(os.Args); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "catHome":
		if err := catHome(os.Args, unnsfw); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "catEvent":
		if err := catEvent(os.Args); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "dispHome":
		if err := catHome(os.Args, unnsfw); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "catSelf":
		if err := catSelf(os.Args); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "emojiReaction":
		if err := emojiReaction(os.Args); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	case "removeEvent":
		if err := removeEvent(os.Args); err != nil {
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
dispHelp {{{
*/
func dispHelp() {
	usageTxt := `Usage :
	nostk <sub-command> [param...]
		init :
			Initializing the nostk environment
		genkey :
			create Private Key and Public Key
		editRelays :
			Edit relay list.
		editContacts :
			Edit your contact list.
		editEmoji :
			Edit custom emoji list.
		pubRelays :
			Publish relay list.
		editProfile :
			Edit your profile.
		pubProfile :
			Publish your profile.
		pubMessage <text message> :
			Publish message to relays.
		catHome [number] [date time] :
			Display home timeline.
			The default number is 20.
			date time format : \"2023/07/24 17:49:51 JST\"
		catSelf [number] [date time] :
			Display your posts.
			The default number is 20.
			date time format : \"2023/07/24 17:49:51 JST\"
		catEvent <hex type Event id> :
			Display the event specified by Event Id.

		removeEvent <hex type Event id> [reason] :
			Remove the event specified by Event Id.
			(Test implementation)`
	fmt.Fprintf(os.Stderr, "%s\n", usageTxt)
}

// }}}

/*
initEnv {{{
*/
func initEnv() error {
	// make skeleton of user profile
	if err := create(profile, ProfileMetadata{"", "", "", "", "", "", "", ""}); err != nil {
		return err
	}
	p := make(map[string]RwFlag)
	p[""] = RwFlag{true, true}
	// make skeleton of relay list
	if err := create(relays, p); err != nil {
		return err
	}
	// make skeleton of custom emoji list
	if err := create(emoji, map[string]string{"name": "url"}); err != nil {
		return err
	}
	// make skeleton of contact list
	c := make(map[string]CONTACT)
	c["hex_pubkey"] = CONTACT{"", ""}
	if err := create(contacts, c); err != nil {
		return err
	}
	return nil
}

// }}}

/*
Generated Key Pair {{{
*/
func genKey() error {
	dirName, err := getDir()
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
	if err = save(dirName, hsec, sk); err != nil {
		return err
	}
	if err = save(dirName, hpub, pk); err != nil {
		return err
	}
	if err = save(dirName, nsec, ns); err != nil {
		return err
	}
	if err = save(dirName, npub, np); err != nil {
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
func publishProfile() error {
	var rl []string
	s, err := load(profile)
	if err != nil {
		fmt.Println("Not found your profile. Use \"nostk init\" and \"nostk editProfile\".")
		return err
	}
	sk, err := load(hsec)
	if err != nil {
		fmt.Println("Nothing key pair. Make key pair.")
		return err
	}
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return err
	}

	if err := getRelayList(&rl, writeFlag); err != nil {
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
func publishRelayList() error {
	p := make(map[string]RwFlag)
	b, err := load(relays)
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

	sk, err := load(hsec)
	if err != nil {
		fmt.Println("Nothing key pair. Make key pair.")
		return err
	}
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return err
	}

	var rl []string
	if err := getRelayList(&rl, writeFlag); err != nil {
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
publishMessage {{{
*/
func publishMessage(args []string) error {
	var s string
	var err error
	if len(args) < 3 {
		s, err = readStdIn()
		if err != nil {
			return errors.New("Not set text message")
		}
	} else {
		s = args[2]
	}

	sk, err := load(hsec)
	if err != nil {
		fmt.Println("Nothing key pair. Make key pair.")
		return err
	}
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return err
	}

	var rl []string
	if err := getRelayList(&rl, writeFlag); err != nil {
		fmt.Println("Nothing relay list. Make a relay list.")
		return err
	}

	tgs := nostr.Tags{}
	if err := setCustomEmoji(s, &tgs); err != nil {
		return err
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
catHome {{{
*/
func catHome(args []string, nsfwFlag bool) error {
	num := defReadNo
	var ut int64 = 0
	var wb []NOSTRLOG
	for i := range args {
		if i < 2 {
			continue
		}
		switch i {
		case 2:
			tmpnum, err := strconv.Atoi(args[2])
			if err != nil {
				layout := "2006/01/02 15:04:05 MST"
				tp, err := time.Parse(layout, args[2])
				if err != nil {
					return errors.New("An unknown argument was specified.")
				} else {
					ut = tp.Unix()
				}
			} else {
				num = tmpnum
			}
		case 3:
			layout := "2006/01/02 15:04:05 MST"
			tptmp, err := time.Parse(layout, args[3])
			if err != nil {
				num, err = strconv.Atoi(args[3])
				if err != nil {
					return errors.New("An unknown argument was specified.")
				}
			} else {
				ut = tptmp.Unix()
			}
		}
	}

	var rs []string
	if err := getRelayList(&rs, readFlag); err != nil {
		return err
	}
	var npub []string
	if err := getContactList(&npub); err != nil {
		return err
	}

	var filters []nostr.Filter
	if ut > 0 {
		ts := nostr.Timestamp(ut)
		filters = []nostr.Filter{{
			Kinds:   []int{nostr.KindTextNote},
			Authors: npub,
			Until:   &ts,
			Limit:   num,
		}}
	} else {
		filters = []nostr.Filter{{
			Kinds:   []int{nostr.KindTextNote},
			Authors: npub,
			Limit:   num,
		}}
	}

	ctx := context.Background()
	pool := nostr.NewSimplePool(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	timer := time.NewTimer(time.Second * waitTime)
	defer timer.Stop()
	go func() {
		ch := pool.SubManyEose(ctx, rs, filters)
		for event := range ch {
			switch event.Kind {
			case 1:
				var buf string
				if nsfwFlag == unnsfw {
					buf = replaceNsfw(event)
				} else {
					buf = event.Content
				}
				buf = strings.Replace(buf, "\n", "\\n", -1)
				buf = strings.Replace(buf, "\\", "\\\\", -1)
				buf = strings.Replace(buf, "/", "\\/", -1)
				buf = strings.Replace(buf, "\"", "\\\"", -1)
				var Contents CONTENTS
				Contents.Date = fmt.Sprintf("%v", event.CreatedAt)
				Contents.PubKey = event.PubKey
				Contents.Content = buf
				tmp := NOSTRLOG{event.ID, Contents}
				wb = append(wb, tmp)
			}
		}
		return
	}()
	select {
	case <-timer.C:
		sort.Slice(wb, func(i, j int) bool {
			return wb[i].Contents.Date > wb[j].Contents.Date
		})
		fmt.Println("{")
		for i := range wb {
			fmt.Printf(
				"\t\"%v\": {\"date\": \"%v\", \"pubkey\": \"%v\", \"content\": \"%v\"},\n",
				wb[i].Id, wb[i].Contents.Date, wb[i].Contents.PubKey, wb[i].Contents.Content)
		}
		fmt.Println("}")
		return nil
	}
}
// }}}

/*
replaceNsfw {{{
*/
func replaceNsfw(e nostr.IncomingEvent) string {
	if checkNsfw(e.Tags) == false {
		return e.Content
	}
	strReason := getNsfwReason(e.Tags)
	return fmt.Sprintf("Content Warning!!\n%v\n\nEvent ID : %v", strReason, e.ID)
}
// }}}

/*
getNsfwReason {{{
*/
func getNsfwReason(tgs nostr.Tags) string {
	if checkNsfw(tgs) == false {
		return ""
	}
	for a := range tgs {
		if len(tgs[a]) < 1 {
			return ""
		}
		for cw := range tgs[a] {
			if tgs[a][cw] == "content-warning" {
				continue
			}
			return tgs[a][cw]
		}
	}
	return ""
}
// }}}

/*
checkNsfw {{{
*/
func checkNsfw(tgs nostr.Tags) bool {
	if len(tgs) < 1 {
		return false
	}
	for a := range tgs {
		if len(tgs[a]) < 1 {
			return false
		}
		for cw := range tgs[a] {
			if tgs[a][cw] == "content-warning" {
				return true
			}
		}
	}
	return false
}
// }}}

/*
catSelf {{{
*/
func catSelf(args []string) error {
	num := defReadNo
	var ut int64 = 0
	var wb []NOSTRLOG
	for i := range args {
		if i < 2 {
			continue
		}
		switch i {
		case 2:
			tmpnum, err := strconv.Atoi(args[2])
			if err != nil {
				layout := "2006/01/02 15:04:05 MST"
				tp, err := time.Parse(layout, args[2])
				if err != nil {
					return errors.New("An unknown argument was specified.")
				} else {
					ut = tp.Unix()
				}
			} else {
				num = tmpnum
			}
		case 3:
			layout := "2006/01/02 15:04:05 MST"
			tptmp, err := time.Parse(layout, args[3])
			if err != nil {
				num, err = strconv.Atoi(args[3])
				if err != nil {
					return errors.New("An unknown argument was specified.")
				}
			} else {
				ut = tptmp.Unix()
			}
		}
	}

	var rs []string
	if err := getRelayList(&rs, readFlag); err != nil {
		return err
	}
	var npub []string
	if err := getMySelfPubkey(&npub); err != nil {
		return err
	}

	var filters []nostr.Filter
	if ut > 0 {
		ts := nostr.Timestamp(ut)
		filters = []nostr.Filter{{
			Kinds:   []int{nostr.KindTextNote},
			Authors: npub,
			Until:   &ts,
			Limit:   num,
		}}
	} else {
		filters = []nostr.Filter{{
			Kinds:   []int{nostr.KindTextNote},
			Authors: npub,
			Limit:   num,
		}}
	}

	ctx := context.Background()
	pool := nostr.NewSimplePool(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	timer := time.NewTimer(time.Second * waitTime)
	defer timer.Stop()
	go func() {
		ch := pool.SubManyEose(ctx, rs, filters)
		for event := range ch {
			switch event.Kind {
			case 1:
				buf := event.Content
				buf = strings.Replace(buf, "\n", "\\n", -1)
				buf = strings.Replace(buf, "\\", "\\\\", -1)
				buf = strings.Replace(buf, "/", "\\/", -1)
				buf = strings.Replace(buf, "\"", "\\\"", -1)
				var Contents CONTENTS
				Contents.Date = fmt.Sprintf("%v", event.CreatedAt)
				Contents.PubKey = event.PubKey
				Contents.Content = buf
				tmp := NOSTRLOG{event.ID, Contents}
				wb = append(wb, tmp)
			}
		}
		return
	}()
	select {
	case <-timer.C:
		sort.Slice(wb, func(i, j int) bool {
			return wb[i].Contents.Date > wb[j].Contents.Date
		})
		fmt.Println("{")
		for i := range wb {
			fmt.Printf(
				"\t\"%v\": {\"date\": \"%v\", \"pubkey\": \"%v\", \"content\": \"%v\"},\n",
				wb[i].Id, wb[i].Contents.Date, wb[i].Contents.PubKey, wb[i].Contents.Content)
		}
		fmt.Println("}")
		return nil
	}
}

// }}}

/*
catEvent {{{
*/
func catEvent(args []string) error {
	num := singleReadNo

	if len(args) < 3 {
		return errors.New("invalid argument")
	}
	eventId := args[2]

	var rs []string
	if err := getRelayList(&rs, readFlag); err != nil {
		return err
	}

	var npub []string
	if err := getContactList(&npub); err != nil {
		return err
	}

	var filters []nostr.Filter
	filters = []nostr.Filter{{
		IDs:   []string{eventId},
		Kinds: []int{nostr.KindTextNote},
		Limit: num,
	}}

	ctx := context.Background()
	pool := nostr.NewSimplePool(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	timer := time.NewTimer(time.Second * waitTime)
	defer timer.Stop()
	go func() {
		ch := pool.SubManyEose(ctx, rs, filters)
		fmt.Println("{")
		for event := range ch {
			switch event.Kind {
			case 1:
				buf := event.Content
				buf = strings.Replace(buf, "\n", "\\n", -1)
				buf = strings.Replace(buf, "\\", "\\\\", -1)
				buf = strings.Replace(buf, "/", "\\/", -1)
				buf = strings.Replace(buf, "\"", "\\\"", -1)
				fmt.Printf("\"%v\": {\"date\": \"%v\", \"pubkey\": \"%v\", \"content\": \"%v\"},\n", event.ID, event.CreatedAt, event.PubKey, buf)
			}
		}
		fmt.Println("}")
		return
	}()
	select {
	case <-timer.C:
		//fmt.Println("}")
		return nil
	}
}

// }}}

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
func removeEvent(args []string) error {
	var event_id string
	content := ""

	if len(args) < 3 || 4 < len(args) {
		return errors.New("Wrong number of parameters")
	}
	for i := range args {
		if i < 2 {
			continue
		}
		switch i {
		case 2: // event_id
			event_id = args[i]
		case 3: // content
			content = args[i]
		}
	}

	sk, err := load(hsec)
	if err != nil {
		fmt.Println("Nothing key pair. Make key pair.")
		return err
	}
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return err
	}

	var rl []string
	if err := getRelayList(&rl, writeFlag); err != nil {
		fmt.Println("Nothing relay list. Make a relay list.")
		return err
	}

	var t []string
	tgs := nostr.Tags{}
	if err := setCustomEmoji(content, &tgs); err != nil {
		return err
	}
	t = nil
	t = append(t, "e")
	t = append(t, event_id)
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
func emojiReaction(args []string) error {
	var event_id string
	var public_key string
	var content string

	if len(args) < 5 {
		return errors.New("Wrong number of parameters")
	}
	for i := range args {
		if i < 2 {
			continue
		}
		switch i {
		case 2: // event_id
			event_id = args[i]
		case 3: // public_key
			public_key = args[i]
		case 4: // content
			content = args[i]
		}
	}

	sk, err := load(hsec)
	if err != nil {
		fmt.Println("Nothing key pair. Make key pair.")
		return err
	}
	pk, err := nostr.GetPublicKey(sk)
	if err != nil {
		return err
	}

	var rl []string
	if err := getRelayList(&rl, readWriteFlag); err != nil {
		fmt.Println("Nothing relay list. Make a relay list.")
		return err
	}

	var t []string
	tgs := nostr.Tags{}
	if err := setCustomEmoji(content, &tgs); err != nil {
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
	t = append(t, "1")
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
getDir {{{
*/
func getDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	home = filepath.Join(home, secretDir)
	if _, err := os.Stat(home); err != nil {
		if err = os.Mkdir(home, 0700); err != nil {
			return "", err
		}
	}
	return home, nil
}

// }}}

/*
getRelayList {{{
*/
func getRelayList(rl *[]string, rwFlag int) error {
	c := make(map[string]RwFlag)
	f, err := openJSON5(relays)
	if err != nil {
		return err
	}
	defer f.Close()

	var data interface{}
	dec := json5.NewDecoder(f)
	err = dec.Decode(&data)
	if err != nil {
		return err
	}
	b, err := json5.Marshal(data)
	if err != nil {
		return err
	}

	if err := json5.Unmarshal([]byte(b), &c); err != nil {
		return err
	}

	for i := range c {
		if (c[i].Read == true && rwFlag == readFlag) ||
			(c[i].Write == true && rwFlag == writeFlag) ||
			(rwFlag == readWriteFlag) {
			*rl = append(*rl, i)
		}
	}
	return nil
}

// }}}

/*
getContactList {{{
*/
func getContactList(cl *[]string) error {
	c := make(map[string]CONTACT)
	f, err := openJSON5(contacts)
	if err != nil {
		return err
	}
	defer f.Close()

	var data interface{}
	dec := json5.NewDecoder(f)
	err = dec.Decode(&data)
	if err != nil {
		return err
	}
	b, err := json5.Marshal(data)
	if err != nil {
		return err
	}

	if err := json5.Unmarshal([]byte(b), &c); err != nil {
		return err
	}

	for i := range c {
		*cl = append(*cl, i)
	}
	return nil
}

// }}}

/*
getMySelfPubkey {{{
*/
func getMySelfPubkey(cl *[]string) error {
	b, err := load(hpub)
	if err != nil {
		return err
	}
	*cl = append(*cl, b)
	return nil
}

// }}}

/*
setCustomEmoji {{{
*/
func setCustomEmoji(s string, tgs *nostr.Tags) error {
	*tgs = nil
	ts := make(map[string]string)
	if err := getCustomEmoji(&ts); err != nil {
		return nil
	}
	var t []string
	for i := range ts {
		if strings.Contains(s, ":"+i+":") {
			t = nil
			t = append(t, "emoji")
			t = append(t, i)
			t = append(t, ts[i])
			*tgs = append(*tgs, t)
		}
	}
	return nil
}

// }}}

/*
getCustomEmoji {{{
*/
func getCustomEmoji(ts *map[string]string) error {
	b, err := load(emoji)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(b), ts)
	if err != nil {
		return err
	}
	return nil
}

// }}}

/*
create {{{
*/
func create(fn string, v any) error {
	s, err := json.Marshal(v)
	if err != nil {
		return err
	}

	d, err := getDir()
	if err != nil {
		return err
	}
	path := filepath.Join(d, fn)
	return os.WriteFile(path, s, 0644)
}

// }}}

/*
openJSON5 {{{
*/
func openJSON5(fn string) (*os.File, error) {
	d, err := getDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(d, fn)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// }}}

/*
load {{{
*/
func load(fn string) (string, error) {
	d, err := getDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(d, fn)
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	r := strings.ReplaceAll(string(b), "\n", "")
	r = strings.ReplaceAll(r, "\t", "")
	return r, nil
}

//}}}

/*
save {{{
*/
func save(dn string, fn string, value string) error {
	path := filepath.Join(dn, fn)
	return os.WriteFile(path, []byte(value), 0644)
}

//}}}

/*
edit {{{
*/
func edit(fn string) error {
	e := os.Getenv("EDITOR")
	if e == "" {
		return errors.New("Not set EDITOR environment variables")
	}
	d, err := getDir()
	if err != nil {
		return err
	}
	path := filepath.Join(d, fn)
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("Not found %q. Use \"nostk init\"\n", fn)
		return fmt.Errorf("Not found %q. Use \"nostk init\"\n", fn)
	}
	c := exec.Command(e, path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return err
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
