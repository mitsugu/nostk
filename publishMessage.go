package main

import (
	"errors"
	"github.com/yosuke-furukawa/json5/encoding/json5"
)

const (
	tagName = 0
	reason  = 1
	person  = 1
)

/* publishMessage {{{
 */
func publishMessage(args []string, cc confClass) error {
	dataRawArg := RawArg{}
	tmpArgs := []string{
		"nostk",
		"pubMessage",
	}
	switch len(args) {
	case 1:
		return errors.New("Not enough arguments")
	case 2:
		if tmpContent, err := readStdIn(); err != nil {
			return errors.New("Not set text message")
		} else {
			dataRawArg.Kind = 1
			dataRawArg.Content = tmpContent
		}
	case 3, 4:
		if tmpArgJson, err := buildJson(args); err != nil {
			return err
		} else {
			err = json5.Unmarshal([]byte(tmpArgJson), &dataRawArg)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("Too meny argument")
	}
	if tmp, err := json5.Marshal(dataRawArg); err != nil {
		return err
	} else {
		tmpArgs = append(tmpArgs, string(tmp))
	}
	if err := publishRaw(tmpArgs, cc); err != nil {
		return err
	}

	return nil
}

// }}}

// vim: set ts=2 sw=2 et:
