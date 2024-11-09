package main

import (
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
	//"log"
	"os"
	//"reflect"
	)
/*
decord
*/
type Decorder struct {}
func (r Decorder) decord(str string) (string, string, error) {
	if pref,temp, err := nip19.Decode(str); err != nil {
		return "","", err
	} else {
		switch temp.(type) {
		case string:
			result := temp.(string)
			return pref, result, nil
		case nostr.EventPointer:
			data := temp.(nostr.EventPointer)
			return pref, data.ID, nil
		case nostr.ProfilePointer:
			data := temp.(nostr.ProfilePointer)
			return pref, data.PublicKey, nil
		case nostr.EntityPointer:
			data := temp.(nostr.EntityPointer)
			return pref, data.PublicKey, nil
		}
	}
	return "", "", errors.New("Not support bech32 type")
}
func decord(args []string, cc confClass) error {
	decorder := Decorder{}
	switch len(args) {
	case 1:
		return errors.New("Not enough arguments")
	case 2:
		// Receive content from standard input
		if str, err := readStdIn(); err != nil {
			return errors.New("Not set text message")
		} else {
			if pref, data, err := decorder.decord(str); err != nil {
				return err
			} else {
				fmt.Printf("%v: %v\n", pref, data)
			}
		}
	case 3:
		str := os.Args[2]
		if pref, hex, err := decorder.decord(str); err != nil {
			return err
		} else {
			fmt.Printf("%v: %v\n", pref, hex)
		}
	default:
		return errors.New("Too meny argument")
	}

	return nil
}

// }}}
