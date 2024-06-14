package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"math"
	"strings"
	"time"
)

/*
catEvent {{{
*/
func catEvent(args []string, cc confClass) error {
	if len(args) < 3 {
		return errors.New("invalid argument")
	}
	eventId := args[2]

	c := cc.getConf()
	num := 1

	var rs []string
	if err := cc.getRelayList(&rs, readFlag); err != nil {
		return err
	}

	var npub []string
	if err := cc.getContactList(&npub); err != nil {
		return err
	}

	var filters []nostr.Filter
	filters = []nostr.Filter{{
		IDs:   []string{eventId},
		Kinds: []int{nostr.KindTextNote},
		Limit: num,
	}}

	ctx := context.Background()
	pool := nostr.NewSimplePool(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wt := time.Duration(int64(math.Ceil(float64(num)*c.Settings.MultiplierReadRelayWaitTime))) * time.Second
	timer := time.NewTimer(wt)
	defer timer.Stop()
	go func() {
		ch := pool.SubManyEose(ctx, rs, filters)
		fmt.Println("{")
		for event := range ch {
			switch event.Kind {
			case 1:
				buf := event.Content
				buf = strings.Replace(buf, "\n", "\\n", -1)
				buf = strings.Replace(buf, "\\", "\\\\", -1)
				buf = strings.Replace(buf, "/", "\\/", -1)
				buf = strings.Replace(buf, "\"", "\\\"", -1)
				fmt.Printf("\"%v\": {\"date\": \"%v\", \"pubkey\": \"%v\", \"content\": \"%v\"},\n", event.ID, event.CreatedAt, event.PubKey, buf)
			}
		}
		fmt.Println("}")
		return
	}()
	select {
	case <-timer.C:
		//fmt.Println("}")
		return nil
	}
}

// }}}
