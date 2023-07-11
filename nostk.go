package main

import (
    "log"
    "fmt"
    "os"
    "os/exec"
    "errors"
    "net/url"
    "strings"
    "context"
    "io/ioutil"
    "encoding/json"
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
  profile = "profile"
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

/*
  main
*/
func main() {
  if len(os.Args)<2 {
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
      if err := initEnv(); err!=nil {
        log.Fatal(err)
      }
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
      if err := rmRelay(os.Args[2]); err!=nil {
        log.Fatal(err)
      }
    case "clearRelay":
      if err := clearRelay(); err!=nil {
        log.Fatal(err)
      }
    case "pubMessage":
      if len(os.Args)<3 {
      fmt.Println("Nothing text message.")
        log.Fatal(errors.New("Not set text message"))
        os.Exit(1)
      }
      if err := publishMessage(os.Args[2]); err!=nil {
        log.Fatal(err)
      }
    case "editProfile":
      if err:=editProfile();err!=nil {
        log.Fatal(err)
      }
    case "pubProfile":
      if err := publishProfile(); err!=nil {
        log.Fatal(err)
      }
  }
}

/*
  dispHelp {{{
*/
func dispHelp() {
  const (
    usage = "Usage :\n  nostk <sub-command> [param...]"
    subcommand = "    sub-command :"
    strInit = "        init : Initializing the nostk environment"
    genkey = "        genkey : create Prive Key and Public Key"
    strAddRelay = "        addRelay <relay's URL> : add relay to nostk\n            ex) nostk addRelay wss://relay.nostr.wirednet.jp"
    strListRelay = "        lsRelay : Show relay list"
    strRmRelay = "        rmRelay <relay's URL> : remove relay to nostk\n            ex) nostk rmRelay wss://relay.nostr.wirednet.jp"
    strClearRelays = "        clearRelay : Clear relay list"
    strPublishMessage = "        pubMessage <text message>: Publish message to relays."
    strEditProfile= "        editProfile : Edit your profile."
    strPublishProfile= "        pubProfile <-n user_name> [-a about you text] [-u your icon image URL]: Publish your profile."
  )

  fmt.Println(usage)
  fmt.Println(subcommand)
  fmt.Println(strInit)
  fmt.Println(genkey)
  fmt.Println(strAddRelay)
  fmt.Println(strListRelay)
  fmt.Println(strRmRelay)
  fmt.Println(strClearRelays)
  fmt.Println(strPublishMessage)
  fmt.Println(strEditProfile)
  fmt.Println(strPublishProfile)
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
  rmRelay {{{
*/
func rmRelay(r string) error {
  var rl [] string
  var tmp [] string
  dirName, err := getDir()
  if err != nil {
    return err
  }
  path := dirName+"/"+relays
  if _, err := os.Stat(path); err!=nil {
    return nil
  }
  if err := getRelayList(&rl); err!=nil {
    return err
  }
  for _, rs := range rl {
    if r!=rs {
      tmp = append(tmp,rs)
    }
  }
  if err := saveRelays(tmp); err!=nil {
    return err
  }
  return nil
}
// }}}

/*
  clearRelay {{{
*/
func clearRelay() error {
  dirName, err := getDir()
  if err != nil {
    return err
  }
  path := dirName+"/"+relays
  if _, err := os.Stat(path); err!=nil {
    return nil
  }
  if err := os.Remove(path); err!=nil {
    return err
  }
  return nil
}
// }}}

/*
  publishMessage {{{
*/
func publishMessage(s string) error {
  var rl [] string

  if len(s)<1 {
    fmt.Println("Nothing text message.")
    return errors.New("Not set text message")
  }
  sk, err := readPrivateKey()
  if err!=nil {
    fmt.Println("Nothing key pair. Make key pair.")
    return err
  }
  pk, err := nostr.GetPublicKey(sk)
  if err!=nil {
    return err
  }

  if err := getRelayList(&rl);err!=nil {
    fmt.Println("Nothing relay list. Make a relay list.")
    return err
  }

  ev := nostr.Event{
    PubKey:    pk,
    CreatedAt: nostr.Now(),
    Kind:      nostr.KindTextNote,
    Tags:      nil,
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
  editProfile {{{
*/
func editProfile()error{
  e := os.Getenv("EDITOR")
  if e == "" {
    return errors.New("Not set EDITOR environmental variables")
  }
  d, err := getDir()
  if err!=nil {
    return err
  }
  path := d+"/"+profile
  if _, err := os.Stat(path); err!=nil {
    fmt.Println("Not found profile file.")
    return errors.New("Not found profile file")
  }
  c := exec.Command(e, path)
  c.Stdin = os.Stdin
  c.Stdout = os.Stdout
  c.Stderr = os.Stderr
  if err :=c.Run(); err!=nil {
    return err
  }

  return nil
}
// }}}

/*
  publishProfile {{{
*/
func publishProfile() error {
  var rl [] string
  s, err := readProfile()
  if err != nil {
    fmt.Println("Not found your profile. Use \"nostk init\" and \"nostk editProfile\".")
    return err
  }

  sk, err := readPrivateKey()
  if err!=nil {
    fmt.Println("Nothing key pair. Make key pair.")
    return err
  }
  pk, err := nostr.GetPublicKey(sk)
  if err!=nil {
    return err
  }

  if err := getRelayList(&rl);err!=nil {
    fmt.Println("Nothing relay list. Make a relay list.")
    return err
  }

  ev := nostr.Event{
    PubKey:    pk,
    CreatedAt: nostr.Now(),
    Kind:      nostr.KindSetMetadata,
    Tags:      nil,
    Content:   string(s),
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
  initEnv {{{
*/
func initEnv() error {
  // make .nostk directory
  dir, err := getDir()
  if err!=nil {
    return err
  }
  // make skeleton of user profile
  if err := createProfile(dir); err!=nil {
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
  saveRelays {{{
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

/*
  readPrivateKey {{{
*/
func readPrivateKey() (string, error) {
  var k [] string
  dir, err :=getDir()
  if err!=nil {
    return "", err
  }
  path := dir+"/"+hsec
  if _, err := os.Stat(path); err!=nil {
    return "", err
  }
  b, err := ioutil.ReadFile(path)
  rs := strings.Split(string(b),"\n")
  for _, r := range rs {
    if r!="" {  // 最終行の\nにより発生する余分なレコードを排除
      k = append(k, r)
    }
  }
  return k[0], nil
}
// }}}

/*
  createProfile {{{
*/
func createProfile(d string) error {
  p := ProfileMetadata{"","","","","","","","",}
  s, err := json.Marshal(p)
  if err!=nil {
    return err
  }
  path := d+"/"+profile
  fp, err := os.Create(path)
  if err != nil {
    return err
  }
  defer fp.Close()
  _,err = fp.WriteString(string(s))
  if err!=nil {
    return err
  }
  return nil
}
// }}}

/*
  readProfile {{{
*/
func readProfile() (string,error) {
  var k [] string
  d, err :=getDir()
  if err!=nil {
    return "", err
  }
  path := d+"/"+profile
  if _, err := os.Stat(path); err!=nil {
    return "", err
  }
  b, err := ioutil.ReadFile(path)
  rs := strings.Split(string(b),"\n")
  for _, r := range rs {
    if r!="" {  // 最終行の\nにより発生する余分なレコードを排除
      k = append(k, r)
    }
  }
  return k[0], nil
}
//}}}

