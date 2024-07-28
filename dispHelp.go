package main

import (
	"fmt"
	"os"
)

/*
dispHelp {{{
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

		pubMessage <text message> [content warning reason]:
			Publish message to relays.
			content warning reason (optional) :
				If this argument is specified,
				the note will be treated as a CONENT WARNING NOTE.
		pubMessageTo <text message> <hex pubkey>:
			Publish message to some user.
			hex pubkey>:
				Specify the user Pubkey to which the note will be sent.
				This argument is required.

		catHome [number] [date time] :
			Display home timeline.
			The default number is 20.
			date time format : \"2023/07/24 17:49:51 JST\"
		catNSFW [number] [date time] :
			Display home timeline include content warning contents.
			The default number is 20.
			date time format : \"2023/07/24 17:49:51 JST\"
		catSelf [number] [date time] :
			Display your posts.
			The default number is 20.
			date time format : \"2023/07/24 17:49:51 JST\"
		catEvent <hex type Event id> :
			Display the event specified by Event Id.

		removeEvent <hex type Event id> [reason] :
			Remove the event specified by Event Id.
			(Test implementation)`
	fmt.Fprintf(os.Stderr, "%s\n", usageTxt)
}

// }}}
