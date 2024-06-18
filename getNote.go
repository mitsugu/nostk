package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"github.com/yosuke-furukawa/json5/encoding/json5"
	//"log"
	"math"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*
const {{{
*/
const (
	CatHome = "main.catHome"
	CatSelf = "main.catSelf"
	CatNSFW = "main.catNSFW"
)

// }}}

/*
filters structure {{{
*/
type Filters struct {
	Content string `json:"content"`
	Tags    string `json:"tags"`
}

// }}}

type UserFilter struct {
	strContentExp string
	strTagsExp    string
}

/*
readUserFilter {{{
*/
func (uf *UserFilter) readUserFilter(cc confClass) error {
	f, err := cc.openJSON5(cc.ConfData.Filename.Filters)
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

	var c Filters
	if err := json5.Unmarshal([]byte(b), &c); err != nil {
		return err
	}

	uf.strContentExp = c.Content
	uf.strTagsExp = c.Tags

	return nil
}

// }}}

/*
replaceUserFilter {{{
*/
func (uf *UserFilter) replaceUserFilter(ev nostr.IncomingEvent) (string, error) {
	var builder strings.Builder
	// tags process
	if 0 < len(uf.strTagsExp) {
		re, err := regexp.Compile(uf.strTagsExp)
		if err != nil {
			return "", err
		}

		for i := range ev.Tags {
			for j := range ev.Tags[i] {
				matches := re.FindAllString(ev.Tags[i][j], -1)
				if 0 < len(matches) {
					builder.WriteString("Detecting data in tags using user filters\n")
					for k := range matches {
						builder.WriteString(matches[k])
						builder.WriteString("\n")
					}
					builder.WriteString("\n")
				}
			}
		}
	}

	// content process
	if 0 < len(uf.strContentExp) {
		re, err := regexp.Compile(uf.strContentExp)
		if err != nil {
			return "", err
		}
		matches := re.FindAllString(ev.Content, -1)
		if 0 < len(matches) {
			builder.WriteString("Detecting data in content using user filters\n")
			for i := range matches {
				builder.WriteString(matches[i])
				builder.WriteString("\n")
			}
			builder.WriteString("\n")
		}
	}
	if 0 < len(builder.String()) {
		buf := fmt.Sprintf("Evint ID : %v\n", ev.ID)
		builder.WriteString(buf)
		return builder.String(), nil
	}
	return "", nil
}

// }}}

/*
getNote {{{
*/

func getNote(args []string, cc confClass) error {
	var ut int64 = 0
	var wb []NOSTRLOG

	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	var uf UserFilter
	if err := uf.readUserFilter(cc); err != nil {
		return nil
	}

	c := cc.getConf()
	num := c.Settings.DefaultReadNo

	for i := range args {
		if i < 2 {
			continue
		}
		switch i {
		case 2:
			tmpnum, err := strconv.Atoi(args[2])
			if err != nil {
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
	if err := cc.getRelayList(&rs, readFlag); err != nil {
		return err
	}
	var npub []string
	if fn.Name() == CatHome || fn.Name() == CatNSFW {
		if err := cc.getContactList(&npub); err != nil {
			return err
		}
	} else if fn.Name() == CatSelf {
		if err := cc.getMySelfPubkey(&npub); err != nil {
			return err
		}
	} else {
		return errors.New("The getNote function is called from a function that cannot use it.")
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

	wt := time.Duration(int64(math.Ceil(float64(num)*c.Settings.MultiplierReadRelayWaitTime))) * time.Second
	timer := time.NewTimer(wt)
	defer timer.Stop()
	go func() error {
		ch := pool.SubManyEose(ctx, rs, filters)
		for event := range ch {
			switch event.Kind {
			case 1:
				tmpBuf, err := uf.replaceUserFilter(event)
				if err != nil {
					return err
				}
				var buf string
				switch fn.Name() {
				case CatHome:
					if c.Settings.DefaultContentWarning {
						buf = replaceNsfw(event)
					}
					if (buf == event.Content) && (0 < len(tmpBuf)) {
						buf = tmpBuf
					}
				case CatSelf:
					buf = event.Content
				case CatNSFW:
					if 0 < len(tmpBuf) {
						buf = tmpBuf
					} else {
						buf = event.Content
					}
				default:
					if 0 < len(tmpBuf) {
						buf = tmpBuf
					} else {
						buf = event.Content
					}
				}
				buf = strings.Replace(buf, "\n", "\\n", -1)
				buf = strings.Replace(buf, "\b", "\\b", -1)
				buf = strings.Replace(buf, "\f", "\\f", -1)
				buf = strings.Replace(buf, "\r", "\\r", -1)
				buf = strings.Replace(buf, "\t", "\\t", -1)
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
		return nil
	}()
	select {
	case <-timer.C:
		sort.Slice(wb, func(i, j int) bool {
			return wb[i].Contents.Date > wb[j].Contents.Date
		})
		fmt.Println("{")
		last := len(wb) - 1
		cnt := 0
		for i := range wb {
			if cnt < last {
				fmt.Printf(
					"\t\"%v\": {\"date\": \"%v\", \"pubkey\": \"%v\", \"content\": \"%v\"},\n",
					wb[i].Id, wb[i].Contents.Date, wb[i].Contents.PubKey, wb[i].Contents.Content)
			} else {
				fmt.Printf(
					"\t\"%v\": {\"date\": \"%v\", \"pubkey\": \"%v\", \"content\": \"%v\"}\n",
					wb[i].Id, wb[i].Contents.Date, wb[i].Contents.PubKey, wb[i].Contents.Content)
			}
			cnt++
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
