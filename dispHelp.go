package main

import (
	"fmt"
	"os"
)

/* dispHelp {{{
 */
func dispHelp() {
	usageTxt := `Usage :
	nostk <sub-command> [param...]
		init :
			Initializing the nostk environment
		genkey :
			create Private Key and Public Key
		editRelays :
			Edit relay list.
		editContacts :
			Edit your contact list.
		editEmoji :
			Edit custom emoji list.

		pubRelays :
			Publish relay list.
		editProfile :
			Edit your profile.
		pubProfile :
			Publish your profile.

		pubMessage <text message> [reason for content warning]:
			Publish text message to relays.
		pubMessageTo <text message> <pubkey>:
			Publish text message to a some user.
		pubRaw <raw data>:
			Publish raw data in json format.
			format:
				See: https://spec.json5.org/
				ex) "{\"kind\" : 1,\"content\" : \"test\",\"tags\":[[\"p\",\"c088_cut_off_05f9e6b5157b7d3416\"]]}"

		catHome [number]:
			Display home timeline.
		catNSFW [number]:
			Display home timeline include content warning contents.
		catSelf [number]:
			Display your posts.
		catEvent <ID> :
			Display the event specified by Event ID.

		emojiReaction <ID> <pubkey> <kind> <reaction>:
			React to specified events.
		removeEvent <ID> <kind> [reason]:
			Remove the event specified by Event ID or Note ID.

		decord <bech32 string>
			Decode bech32 string to hex string.
`
	fmt.Fprintf(os.Stderr, "%s\n", usageTxt)
}

// }}}
