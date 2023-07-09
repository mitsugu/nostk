package main

import (
    "log"
    "fmt"
    "os"
    "errors"
    "github.com/nbd-wtf/go-nostr"
    "github.com/nbd-wtf/go-nostr/nip19"
)

const (
  secretDir = "/.nostk"
  hsec = "hsec"
  nsec = "nsec"
  hpub = "hpub"
  npub = "npub"
  usage       = "Usage :\n  nostk <sub-command> [param...]"
  subcommand  = "    sub-command :"
  genkey      = "      genkey : create Prive Key and Public Key"
)

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
  }
}

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

