package main

import (
    "log"
    "fmt"
    "os"
    "errors"
    "net/url"
    "strings"
    "io/ioutil"
    "github.com/nbd-wtf/go-nostr"
    "github.com/nbd-wtf/go-nostr/nip19"
)

const (
  secretDir = "/.nostk"
  hsec = ".hsec"
  nsec = ".nsec"
  hpub = ".hpub"
  npub = ".npub"
  relays = "relays.list"
  usage       = "Usage :\n  nostk <sub-command> [param...]"
  subcommand  = "    sub-command :"
  genkey      = "      genkey : create Prive Key and Public Key"
)

/*
  main
*/
func main() {
  if len(os.Args)<2 {
    fmt.Println(usage)
    fmt.Println(subcommand)
    fmt.Println(genkey)
    os.Exit(0)
  }
  switch os.Args[1] {
    case "genkey":
      if err := genKey(); err!=nil {
        log.Fatal(err)
      }
    case "addRelay":
      if err := addRelay(os.Args[2]); err!=nil {
        log.Fatal(err)
      }
    case "lsRelays":
      if err := listRelays(); err!=nil {
        log.Fatal(err)
      }
    case "rmRelay":
    case "clearRelay":
  }
}
//

/*
  Generated Key Pair {{{
*/
func genKey() error {
  dirName, err := getDir()
  if err != nil {
    return err
  }
  sk, pk, err := genHexKey()
  if err!=nil {
    return err
  }
  nsec, npub, err := genNKey(sk, pk)
  if err!=nil {
    return err
  }
  if err = saveHSecKey(dirName, sk); err!=nil {
    return err
  }
  if err = saveHPubKey(dirName, pk); err!=nil {
    return err
  }
  if err = saveNSecKey(dirName, nsec); err!=nil {
    return err
  }
  if err = saveNPubKey(dirName, npub); err!=nil {
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

func genNKey(sk string, pk string)(string, string, error) {
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
  path := dn+"/"+hsec
  fp, err := os.Create(path)
  if err != nil {
    return err
  }
  defer fp.Close()
  fp.WriteString(sk)
  return nil
}

func saveNSecKey(dn string, nkey string) error {
  path := dn+"/"+nsec
  fp, err := os.Create(path)
  if err != nil {
    return err
  }
  defer fp.Close()
  fp.WriteString(nkey)
  return nil
}

func saveHPubKey(dn string, pk string) error {
  path := dn+"/"+hpub
  fp, err := os.Create(path)
  if err != nil {
    return err
  }
  defer fp.Close()
  fp.WriteString(pk)
  return nil
}

func saveNPubKey(dn string, nkey string) error {
  path := dn+"/"+npub
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
  var rl [] string
  dirName, err := getDir()
  if err != nil {
    return err
  }
  path := dirName+"/"+relays
  if _, err := os.Stat(path); err!=nil {
    fmt.Println("Nothing relay list. Make a relay list.")
    return err
  }
  if err := getRelayList(&rl); err!=nil {
    return err
  }
  for _, r := range rl {
    fmt.Println(r)
  }
  return nil
}
// }}}

/*
  Add Relay {{{
*/
func addRelay(r string) error {
  var rl [] string
  u, err := url.Parse(r)
  if err!=nil {
    return err
  }
  if u.Scheme!="wss" {
    return errors.New("Not wss scheme.")
  }
  if err := getRelayList(&rl); err!=nil {
    return err
  }
  rl = append(rl, r)
  rmDuplication(&rl)
  err = saveRelays(rl)
  if err!=nil {
    return err
  }
  return nil
}
// }}}

/*
  getDir {{{
*/
func getDir() ( string, error ) {
  home := os.Getenv("HOME")
  if home == "" {
    return "", errors.New("Not set HOME environmental variables")
  }
  home += secretDir
  if _, err := os.Stat(home); err!=nil {
    if err = os.Mkdir(home, 0700); err != nil {
      return "", err
    }
  }
  return home, nil
}
// }}}

/*
  rmDuplication {{{
*/
func rmDuplication(rl *[]string) {
  m := make(map[string]bool)
  uniq := [] string{}

  for _, ele := range *rl {
    if !m[ele] {
      m[ele] = true
      uniq = append(uniq, ele)
    }
  }
  *rl = append(uniq)
}
// }}}

/*
  getRelayList {{{
*/
func getRelayList(rl *[]string) error {
  dir, err := getDir()
  if err!=nil {
    return err
  }
  path := dir+"/"+relays
  if _, err := os.Stat(path); err!=nil {
    return nil
  }
  b, err := ioutil.ReadFile(path)
  rs := strings.Split(string(b),"\n")
  for _, r := range rs {
    if r!="" {  // 最終行の\nにより発生する余分なレコードを排除
      *rl = append(*rl, r)
    }
  }
  return nil
}
// }}}

/*
  saveRelays ( Not yet test ) {{{
*/
func saveRelays(rl []string) error {
  dn, err := getDir()
  if err!=nil {
    return err
  }

  path := dn+"/"+relays
  fp, err := os.Create(path)
  if err != nil {
    return err
  }
  defer fp.Close()
  for _, l := range rl {
    _,err = fp.WriteString(l+"\n")
    if err!=nil {
      return err
    }
  }
  return nil
}
// }}}

