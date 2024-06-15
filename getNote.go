package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	//"log"
	"math"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	CatHome = "main.catHome"
	CatSelf = "main.catSelf"
)

/*
getNote {{{
*/

func getNote(args []string, cc confClass) error {
	var ut int64 = 0
	var wb []NOSTRLOG

	pc, _, _, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

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
	if fn.Name() == CatHome {
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
	go func() {
		ch := pool.SubManyEose(ctx, rs, filters)
		for event := range ch {
			switch event.Kind {
			case 1:
				var buf string
				if (fn.Name() == CatHome) && (c.Settings.DefaultContentWarning) {
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
