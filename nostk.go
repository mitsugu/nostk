package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

const (
	secretDir = ".nostk"
	hsec      = ".hsec"
	nsec      = ".nsec"
	hpub      = ".hpub"
	npub      = ".npub"
	relays    = "relays.json"
	profile   = "profile.json"
	emoji     = "customemoji.json"
	contacts  = "contacts.json"
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
	Url	 string	`json:"url"`
	Name string	`json:"name"`
}

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
	const (
		usage             = "Usage :\n  nostk <sub-command> [param...]"
		subcommand        = "    sub-command :"
		strInit           = "        init : Initializing the nostk environment"
		genkey            = "        genkey : create Prive Key and Public Key"
		strEditContacts   = "        editContacts : Edit your contact list."
		strCustomEmoji    = "        editEmoji : Edit custom emoji list."
		strEditRelay      = "        editRelays : edit relay list."
		strPubRelay       = "        pubRelays : Publish relay list."
		strEditProfile    = "        editProfile : Edit your profile."
		strPublishProfile = "        pubProfile : Publish your profile."
		strPublishMessage = "        pubMessage <text message>: Publish message to relays."
		strDispHome       = "        dispHome [number] : Display the specified number of home timelines. The default number is 20."
	)

	fmt.Println(usage)
	fmt.Println(subcommand)
	fmt.Println(strInit)
	fmt.Println(genkey)
	fmt.Println(strCustomEmoji)
	fmt.Println(strEditContacts)
	fmt.Println(strEditRelay)
	fmt.Println(strPubRelay)
	fmt.Println(strEditProfile)
	fmt.Println(strPublishProfile)
	fmt.Println(strPublishMessage)
	fmt.Println(strDispHome)
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

	if err := getRelayList(&rl); err != nil {
		fmt.Println("Nothing relay list. Make a relay list.")
		return err
	}

	pr := strings.Replace(s, "\\n", "\n", -1)
	ev := nostr.Event{
		PubKey:    pk,
		CreatedAt: nostr.Now(),
		Kind:      nostr.KindSetMetadata,
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
		_, err = relay.Publish(ctx, ev)
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
	if err := getRelayList(&rl); err != nil {
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
		_, err = relay.Publish(ctx, ev)
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
	if len(args)<3 {
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
	if err := getRelayList(&rl); err != nil {
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
		_, err = relay.Publish(ctx, ev)
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
	home := os.Getenv("HOME")
	if home == "" {
		return "", errors.New("Not set HOME environment variables")
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
func getRelayList(rl *[]string) error {
	p := make(map[string]RwFlag)
	b, err := load(relays)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(b), &p)
	if err != nil {
		return err
	}
	for i := range p {
		*rl = append(*rl, i)
	}
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
		return err
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
	return ioutil.WriteFile(path, s, 0644)
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
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	r := strings.ReplaceAll(string(b), "\n", "")
	return r, nil
}

//}}}

/*
save {{{
*/
func save(dn string, fn string, value string) error {
	path := filepath.Join(dn, fn)
	return ioutil.WriteFile(path, []byte(value), 0644)
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
