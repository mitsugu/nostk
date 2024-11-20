package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip19"
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

type Recieve struct {
	RelayUrl string			// data.Relay.URL
	Event    nostr.Event	// data.Event
}

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
func (uf *UserFilter) replaceUserFilter(ev nostr.RelayEvent) (string, error) {
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

/* getNote {{{
*/

func getNote(args []string, cc confClass) error {
	var ut int64 = 0
	//var wb []NOSTRLOG

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
	recieveData := []Recieve{}
	defer cancel()

	wt := time.Duration(int64(math.Ceil(float64(num)*c.Settings.MultiplierReadRelayWaitTime))) * time.Second
	timer := time.NewTimer(wt)
	defer timer.Stop()
	reb := replaceEnginForBech32{}
	go func() error {
		ch := pool.SubManyEose(ctx, rs, filters)
		for event := range ch {
			conv := convertRelayEventToRecieve(&event)
			if tmp, err  := reb.replaceToBech32(conv); err != nil {
				return err
			} else {
				recieveData = append(recieveData, tmp)
			}
		}
		return nil
	}()
	select {
	case <-timer.C:
		sort.Slice(recieveData, func(i, j int) bool {
			return recieveData[i].Event.CreatedAt > recieveData[j].Event.CreatedAt
		})
		if data, err := json5.Marshal(recieveData); err != nil {
			return err
		} else {
			fmt.Printf("%v",string(data))
		}
		return nil
	}
}

// }}}

/* replaceNsfw {{{
*/
func replaceNsfw(e nostr.RelayEvent) string {
	if checkNsfw(e.Tags) == false {
		return e.Content
	}
	strReason := getNsfwReason(e.Tags)
	return fmt.Sprintf("Content Warning!!\n%v\n\nEvent ID : %v", strReason, e.ID)
}

// }}}

/* getNsfwReason {{{
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

/* checkNsfw {{{
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

/* convertRelayEventToRecieve {{{
*/
func convertRelayEventToRecieve(data *nostr.RelayEvent) Recieve {
	return Recieve{
		RelayUrl: data.Relay.URL,
		Event:    *data.Event,
	}
}

// }}}

/* replaceToBech32 {{{
*/
type replaceEnginForBech32 struct{
	cache			Recieve
}
func (r replaceEnginForBech32)replaceToBech32(data Recieve) (Recieve, error) {
	r.cache = data
	// go-nostr のコメントには event ID と書かれているが
	// 実際には note ID が入っているので注意！！
	// event ID だと思って EncordEvent を呼び出すとえらいことに！！
	if tmpdata, err:= nip19.EncodeNote(data.Event.ID); err != nil {
		return data, err
	} else {
		data.Event.ID = tmpdata
	}
	if tmpdata, err := nip19.EncodePublicKey(data.Event.PubKey); err != nil {
		return data, err
	} else {
		data.Event.PubKey = tmpdata
	}
	if tmpdata, err := r.replaceTagsToBech32(data); err != nil {
		return data, err
	} else {
		data = tmpdata
	}
	return data, nil
}
func (r replaceEnginForBech32)replaceTagsToBech32(data Recieve) (Recieve, error) {
	mbl := NewModifyBech32List()
	modified := data
	for tagIndex, tag := range data.Event.Tags {
		if len(tag) < 1 {
			continue
		}

		tagKey := tag[0]
		conversions := mbl.GetByTagKey(tagKey)
		if conversions == nil {
			continue
		}
		for elementIdx, formats := range conversions {
			if elementIdx >= len(tag) {
				continue
			}
			//element := tag[elementIdx]
			for _, format := range formats {
				newValue, err := r.convert(data, tagIndex, elementIdx, format)
				if err != nil {
					return modified, err
				} else {
					modified = newValue
				}
			}
		}
	}
	return modified, nil
}
/*
この関数は引数で指定されたタグ配列の Hex 文字列を実際に
Bech32 文字列に変換する。
replaceTagsToBech32() から呼び出すことのみを前提にしており
そのためタグ内に変換対象データが必ず存在することを前提としている。
他の言語であれば private メンバー関数であり、コードの他の部分からの
参照は行われないことを前提としている。
*/
func (r replaceEnginForBech32)convert(data Recieve, tagIndex int, elementIndex int, format string) (Recieve, error) {
	switch format {
	case "nevent":
		// tag 内の PubKey の有無をチェックすること
		// 現在 e タグを前提としているので、今後変更する可能性がある
		switch r.cache.Event.Tags[tagIndex][0] {
		case "e", "q":
			strPubKey := ""
			var indexPubKey int
			switch r.cache.Event.Tags[tagIndex][0] {
			case "e":
				indexPubKey = 4
			case "q":
				indexPubKey = 3
			}
			if (indexPubKey + 1) == len(data.Event.Tags[tagIndex]) {
				strPubKey = r.cache.Event.Tags[tagIndex][indexPubKey]
			}
			if tmpdata, err := nip19.EncodeEvent(data.Event.Tags[tagIndex][elementIndex], []string{data.RelayUrl}, strPubKey); err != nil {
				return data, err
			} else {
				data.Event.Tags[tagIndex][elementIndex] = tmpdata
			}
		}
	case "npub":
		if tmpdata, err := nip19.EncodePublicKey(data.Event.Tags[tagIndex][elementIndex]); err != nil {
			return data, err
		} else {
			data.Event.Tags[tagIndex][elementIndex] = tmpdata
		}
	}
	return data, nil
}

// }}}

