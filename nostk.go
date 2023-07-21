package main

import (
	"os"
	"time"
	"os/exec"
	"strings"
	"bufio"
	"io/ioutil"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
)

const (
	secretDir = "/.nostk"
	hsec	= ".hsec"
	nsec	= ".nsec"
	hpub	= ".hpub"
	npub	= ".npub"
	relays	= "relays.json"
	profile	= "profile.json"
	emoji	= "customemoji.json"
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
	case "lsRelays":
		if err := listRelays(); err != nil {
			log.Fatal(err)
		}
	case "editRelays":
		if err := editRelayList(); err != nil {
			log.Fatal(err)
		}
	case "editProfile":
		if err := editProfile(); err != nil {
			log.Fatal(err)
		}
	case "editEmoji":
		if err := editCustomEmojiList(); err != nil {
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
		if len(os.Args) > 2 {
			if err := publishMessage(os.Args[2]); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		} else {
			buff, err := readStdIn()
			if err!=nil {
				fmt.Println("Nothing text message.")
				log.Fatal(errors.New("Not set text message"))
				os.Exit(1)
			}
			if err := publishMessage(buff); err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
		}
	}
}
// }}}

/*
dispHelp {{{
*/
func dispHelp() {
	const (
		usage				= "Usage :\n  nostk <sub-command> [param...]"
		subcommand			= "    sub-command :"
		strInit				= "        init : Initializing the nostk environment"
		genkey				= "        genkey : create Prive Key and Public Key"
		strListRelay		= "        lsRelay : Show relay list"
		strEditRelay		= "        editRelays : edit relay list."
		strPubRelay			= "        pubRelays : Publish relay list."
		strEditProfile		= "        editProfile : Edit your profile."
		strCustomEmoji		= "        editEmoji : Edit custom emoji list."
		strPublishProfile	= "        pubProfile: Publish your profile."
		strPublishMessage	= "        pubMessage <text message>: Publish message to relays."
	)

	fmt.Println(usage)
	fmt.Println(subcommand)
	fmt.Println(strInit)
	fmt.Println(genkey)
	fmt.Println(strListRelay)
	fmt.Println(strEditRelay)
	fmt.Println(strPubRelay)
	fmt.Println(strEditProfile)
	fmt.Println(strCustomEmoji)
	fmt.Println(strPublishProfile)
	fmt.Println(strPublishMessage)
}

// }}}

/*
initEnv {{{
*/
func initEnv() error {
	// make skeleton of user profile
	if err := createProfile(); err != nil {
		return err
	}
	// make skeleton of user profile
	if err := createRelayList(); err != nil {
		return err
	}
	// make skeleton of custom emoji list
	if err := createCustomEmojiList(); err != nil {
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
	nsec, npub, err := genNKey(sk, pk)
	if err != nil {
		return err
	}
	if err = saveHSecKey(dirName, sk); err != nil {
		return err
	}
	if err = saveHPubKey(dirName, pk); err != nil {
		return err
	}
	if err = saveNSecKey(dirName, nsec); err != nil {
		return err
	}
	if err = saveNPubKey(dirName, npub); err != nil {
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

func saveHSecKey(dn string, sk string) error {
	path := dn + "/" + hsec
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	fp.WriteString(sk)
	return nil
}

func saveNSecKey(dn string, nkey string) error {
	path := dn + "/" + nsec
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	fp.WriteString(nkey)
	return nil
}

func saveHPubKey(dn string, pk string) error {
	path := dn + "/" + hpub
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	fp.WriteString(pk)
	return nil
}

func saveNPubKey(dn string, nkey string) error {
	path := dn + "/" + npub
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	fp.WriteString(nkey)
	return nil
}

// }}}

/*
Listing Relays {{{
*/
func listRelays() error {
	p := make(map[string]RwFlag)
	b, err := readRelayList()
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(b), &p)
	if err != nil {
		return err
	}
	for i := range p {
		fmt.Printf("%v R:%v W:%v\n", i, p[i].Read, p[i].Write)
	}
	return nil
}

// }}}

/*
editRelayList {{{
*/
func editRelayList() error {
	e := os.Getenv("EDITOR")
	if e == "" {
		return errors.New("Not set EDITOR environmental variables")
	}
	d, err := getDir()
	if err != nil {
		return err
	}
	path := d + "/" + relays
	if _, err := os.Stat(path); err != nil {
		fmt.Println("Not found relay list. Use \"nostk init\"")
		return errors.New("Not found relay list")
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
editCustomEmojiList {{{
*/
func editCustomEmojiList() error {
	e := os.Getenv("EDITOR")
	if e == "" {
		return errors.New("Not set EDITOR environmental variables")
	}
	d, err := getDir()
	if err != nil {
		return err
	}
	path := d + "/" + emoji
	if _, err := os.Stat(path); err != nil {
		fmt.Println("Not found custom emoji list. Use \"nostk init\"")
		return errors.New("Not found custom emoji list")
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
editProfile {{{
*/
func editProfile() error {
	e := os.Getenv("EDITOR")
	if e == "" {
		return errors.New("Not set EDITOR environmental variables")
	}
	d, err := getDir()
	if err != nil {
		return err
	}
	path := d + "/" + profile
	if _, err := os.Stat(path); err != nil {
		fmt.Println("Not found profile file.")
		return errors.New("Not found profile file")
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
publishProfile {{{
*/
func publishProfile() error {
	var rl []string
	s, err := readProfile()
	if err != nil {
		fmt.Println("Not found your profile. Use \"nostk init\" and \"nostk editProfile\".")
		return err
	}
	sk, err := readPrivateKey()
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

	pr := strings.Replace(s,"\\n","\n",-1)
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
publishMessage {{{
*/
func publishMessage(s string) error {
	var rl []string

	if len(s) < 1 {
		fmt.Println("Nothing text message.")
		return errors.New("Not set text message")
	}

	sk, err := readPrivateKey()
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

	tgs := nostr.Tags{}
	if err := setCustomEmoji(s, &tgs); err!=nil {
		return err
	}

	s = strings.Replace(s,"\\n","\n",-1)
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
	publishRelayList {{{
*/
func publishRelayList() error {
	p := make(map[string]RwFlag)
	b, err := readRelayList()
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(b), &p)
	if err != nil {
		return err
	}
	const (
		cRead	= "read"
		cWrite	= "write"
	)
	tags := nostr.Tags{}
	for i := range p {
		t := nostr.Tag {"r", i}
	    if p[i].Read == true && p[i].Write == true {
		}else if p[i].Read == true {
			t = append(t, cRead)
		}else if p[i].Write == true {
			t = append(t, cWrite)
		}
		tags = append(tags,t)
	}

	sk, err := readPrivateKey()
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
getDir {{{
*/
func getDir() (string, error) {
	home := os.Getenv("HOME")
	if home == "" {
		return "", errors.New("Not set HOME environmental variables")
	}
	home += secretDir
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
	b, err := readRelayList()
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
saveRelays {{{
*/
func saveRelays(rl []string) error {
	dn, err := getDir()
	if err != nil {
		return err
	}

	path := dn + "/" + relays
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	for _, l := range rl {
		_, err = fp.WriteString(l + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

// }}}

/*
readPrivateKey {{{
*/
func readPrivateKey() (string, error) {
	var k []string
	dir, err := getDir()
	if err != nil {
		return "", err
	}
	path := dir + "/" + hsec
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	b, err := ioutil.ReadFile(path)
	rs := strings.Split(string(b), "\n")
	for _, r := range rs {
		if r != "" { // 最終行の\nにより発生する余分なレコードを排除
			k = append(k, r)
		}
	}
	return k[0], nil
}

// }}}

/*
	setCustomEmoji {{{
*/
func setCustomEmoji(s string, tgs *nostr.Tags)error{
	*tgs = nil
	ts := make(map[string]string)
	if err := getCustomEmoji(&ts);err!=nil {
		return  err
	}
	var t []string
	for i:=range ts {
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
	b, err := readCustomEmojiList()
	if err!=nil {
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
createProfile {{{
*/
func createProfile() error {
	p := ProfileMetadata{"", "", "", "", "", "", "", ""}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}

	d, err := getDir()
	if err != nil {
		return err
	}
	path := d + "/" + profile
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = fp.WriteString(string(s))
	if err != nil {
		return err
	}
	return nil
}

// }}}

/*
createRelayList {{{
*/
func createRelayList() error {
	p := make(map[string]RwFlag)
	p[""] = RwFlag{true, true}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}

	d, err := getDir()
	if err != nil {
		return err
	}
	path := d + "/" + relays
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = fp.WriteString(string(s))
	if err != nil {
		return err
	}
	return nil
}

// }}}

/*
createCustomEmojiList {{{
*/
func createCustomEmojiList() error {
	p := map[string]string{
		"name" : "url",
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}

	d, err := getDir()
	if err != nil {
		return err
	}
	path := d + "/" + emoji
	fp, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fp.Close()
	_, err = fp.WriteString(string(s))
	if err != nil {
		return err
	}
	return nil
}

// }}}

/*
readProfile {{{
*/
func readProfile() (string, error) {
	d, err := getDir()
	if err != nil {
		return "", err
	}
	path := d + "/" + profile
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	b, err := ioutil.ReadFile(path)
	r := strings.ReplaceAll(string(b), "\n", "")
	return r, nil
}

//}}}

/*
readRelayList {{{
*/
func readRelayList() (string, error) {
	//var k [] string
	d, err := getDir()
	if err != nil {
		return "", err
	}
	path := d + "/" + relays
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	r := strings.ReplaceAll(string(b), "\n", "")
	return r, nil
}

//}}}

/*
readCustomEmojiList {{{
*/
func readCustomEmojiList() (string, error) {
	d, err := getDir()
	if err != nil {
		return "", err
	}
	path := d + "/" + emoji
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	r := strings.ReplaceAll(string(b), "\n", "")
	return r, nil
}

//}}}

/*
  readStdIn {{{
*/
func readStdIn() (string,error) {
	cn := make(chan string, 1)
		go func() {
		sc := bufio.NewScanner(os.Stdin)
		sc.Scan()
		cn <- sc.Text()
	}()
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	select {
	case text := <-cn:
		return text,nil
	case <-timer.C:
		return "", errors.New("Time out input from standerd input")
	}
}

// }}}

