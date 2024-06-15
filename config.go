package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"github.com/yosuke-furukawa/json5/encoding/json5"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type confClass struct {
	ConfData Conf
}

/*
configuration file structure {{{
*/
type WrapConf struct {
	Conf Conf `json:"conf"`
}
type Filename struct {
	Contacts string `json:"contacts"`
	Emoji    string `json:"emoji"`
	Filters  string `json:"filters"`
	Hpub     string `json:"hpub"`
	Hsec     string `json:"hsec"`
	Npub     string `json:"npub"`
	Nsec     string `json:"nsec"`
	Profile  string `json:"profile"`
	Relays   string `json:"relays"`
}
type Settings struct {
	DefaultContentWarning       bool    `json:"defaultContentWarning"`
	DefaultReadNo               int     `json:"defaultReadNo"`
	MultiplierReadRelayWaitTime float64 `json:"multiplierReadRelayWaitTime"`
}
type Conf struct {
	Filename Filename `json:"filename"`
	Settings Settings `json:"settings"`
}

// }}}

/*
profile structure {{{
*/
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

// }}}

/*
for relays structure {{{
*/
type RwFlag struct {
	Read  bool `json:"read"`
	Write bool `json:"write"`
}

// }}}

/*
const {{{
*/
const (
	secretDir = ".nostk"
	confFile  = "config.json"
)

// }}}

/*
existConfiguration {{{
*/
func (cc *confClass) existConfiguration() error {
	dir, err := cc.getDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, confFile)
	if _, err := os.Stat(path); err != nil {
		// make config.json
		if err := cc.create(confFile,
			`{
  "conf" : {
    "filename" : {
      "hsec" : ".hsec",
      "nsec" : ".nsec",
      "hpub" : ".hpub",
      "npub" : ".npub",
      "relays" : "relays.json",
      "profile" : "profile.json",
      "emoji" : "customemoji.json",
      "contacts" : "contacts.json"
    },
    "settings" : {
      "defaultReadNo" : 20,
      "multiplierReadRelayWaitTime" : 0.001,
      "defaultContentWarning" : true
    }
  }
}
`); err != nil {
			return err
		}
	}
	return nil
}

// }}}

/*
loadConfiguration {{{
*/
func (cc *confClass) loadConfiguration() error {
	var ags WrapConf
	f, err := cc.openJSON5(confFile)
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

	if err := json5.Unmarshal([]byte(b), &ags); err != nil {
		return err
	}
	cc.ConfData = ags.Conf
	return nil
}

// }}}

/*
getConf {{{
*/
func (cc *confClass) getConf() Conf {
	return cc.ConfData
}

// }}}

/*
openJSON5 {{{
*/
func (cc *confClass) openJSON5(fn string) (*os.File, error) {
	d, err := cc.getDir()
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
getDir {{{
*/
func (cc *confClass) getDir() (string, error) {
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
func (cc *confClass) getRelayList(rl *[]string, rwFlag int) error {
	f, err := cc.openJSON5(cc.ConfData.Filename.Relays)
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

	m := make(map[string]RwFlag)
	if err := json5.Unmarshal([]byte(b), &m); err != nil {
		return err
	}

	for i := range m {
		if (m[i].Read == true && rwFlag == readFlag) ||
			(m[i].Write == true && rwFlag == writeFlag) ||
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
func (cc *confClass) getContactList(cl *[]string) error {
	c := make(map[string]CONTACT)
	f, err := cc.openJSON5(cc.ConfData.Filename.Contacts)
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
setCustomEmoji {{{
*/
func (cc *confClass) setCustomEmoji(s string, tgs *nostr.Tags) error {
	*tgs = nil
	ts := make(map[string]string)
	if err := cc.getCustomEmoji(&ts); err != nil {
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
func (cc *confClass) getCustomEmoji(ts *map[string]string) error {
	b, err := cc.load(cc.ConfData.Filename.Emoji)
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
getMySelfPubkey {{{
*/
func (cc *confClass) getMySelfPubkey(cl *[]string) error {
	b, err := cc.load(cc.ConfData.Filename.Hpub)
	if err != nil {
		return err
	}
	*cl = append(*cl, b)
	return nil
}

// }}}

/*
create {{{
*/
func (cc *confClass) create(fn string, s string) error {
	d, err := cc.getDir()
	if err != nil {
		return err
	}
	path := filepath.Join(d, fn)
	return os.WriteFile(path, []byte(s), 0644)
}

// }}}

/*
load {{{
*/
func (cc *confClass) load(fn string) (string, error) {
	d, err := cc.getDir()
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
func (cc *confClass) save(dn string, fn string, value string) error {
	path := filepath.Join(dn, fn)
	return os.WriteFile(path, []byte(value), 0644)
}

//}}}

/*
edit {{{
*/
func (cc *confClass) edit(fn string) error {
	e := os.Getenv("EDITOR")
	if e == "" {
		return errors.New("Not set EDITOR environment variables")
	}
	d, err := cc.getDir()
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
