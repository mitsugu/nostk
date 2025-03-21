nostk
========
Implementing a CLI client to use [Nostr Protocol](https://github.com/nostr-protocol/nostr).

### Develop Environment
* Ubuntu 23.04 and later
* Go Language 1.23.2 and later

### Features
* Initializing the nostk environment
* Generating a key pair
* Edit your contact list
* Edit custom emoji list
* Edit relay list
* Publish relay list
* Edit profile
* Publish profile
* Display home timeline ([kind 1](https://github.com/nostr-protocol/nips/blob/master/01.md#kinds))
* Display your's note ([kind 1](https://github.com/nostr-protocol/nips/blob/master/01.md#kinds))
* Publish Note ([kind 1](https://github.com/nostr-protocol/nips/blob/master/01.md#kinds))
* Publish Note to some user (like Mension, [kind 1](https://github.com/nostr-protocol/nips/blob/master/01.md#kinds))
* Publish raw data (For power users who understand NIPS and the source code.)
* Content warning
* Hash tags
* Publish reaction

### Requirements
* [nbd-wtf / go-nostr](https://github.com/nbd-wtf/go-nostr)
* Some kind of text editor
* Setting $HOME environment variable
* Setting $EDITOR environment variable

### Setup
#### Install tools
1. Install [git](https://www.git-scm.com/)
2. Install [golang](https://go.dev/)

#### Install nostk:
##### Windows
```command.com
SETX EDITOR=<Text editor's full path name>
go install github.com/mitsugu/nostk@<tag name>
```

##### Ubuntu and maybe other distribution
For bash
```bash
echo 'export EDITOR=vim' >> ~/.bashrc
go install github.com/mitsugu/nostk@<tag name>
```

#### Placement of config.json
IF config.json NOT FOUND IN .nostk DIRECTORY, EXECUTE THE FOLLOWING.
1. Download [config.json](https://raw.githubusercontent.com/mitsugu/nostk/main/config.json)
2. Move config.json to "$HOME/.nostk" directory
3. Adjust defaultReadNo, multiplierReadRelayWaitTime, and defaultContentWarning in config.json to your liking.

#### Setting nostk:
1. nostk init (must)
2. nostk genkey (must)
3. nostk editRelays (must)
4. nostk editContacts (must)
5. nostk editProfile (should \*)
6. nostk pubProfile (should \*)
7. nostk editEmoji (Optional)
8. nostk pubRelays (Optional)

\* Unless there is a special reason, it is recommended to use a web app such as [nostter](https://nostter.app/home) instead of nostk.

### Usage
```
nostk <sub-command> [param...]
	init:		Initializing the nostk environment
	genkey:		Create Private Key and Public Key
	editRelays:	Edit relay list.
	editContacts:	Edit your contact list.
	editEmoji:	Edit custom emoji list.

	pubRelays:	Publish relay list.
	editProfile:	Edit your profile.
	pubProfile:	Publish your profile.

	pubMessage <text message> [reason for content warning]:
			Publish text message to relays.
	pubMessageTo <text message> <pubkey>:
			Publish text message to a some user.
	pubRaw <raw data>:
			Publish raw data in json format.
			format: See: https://spec.json5.org/
			ex) "{\"kind\" : 1,\"content\" : \"test\",\"tags\":[[\"p\",\"c088_cut_off_05f9e6b5157b7d3416\"]]}"

	catHome [number]: Display home timeline.
	catNSFW [number]: Display home timeline include content warning contents.
	catSelf [number]: Display your posts.
	catEvent <ID>:	  Display the event specified by Event ID or Note ID.

	emojiReaction <ID> <pubkey> <kind> <reaction>:
			React to specified events.
	removeEvent <ID> <kind> [reason]:
			Remove the event specified by Event ID or Note ID.

	decord <bech32 string>
		Decode bech32 string to hex string.
```

### About content warning note
  The catHome subcommand does not directly display notes with content warnings.  

  The corresponding note will be printed to indicate that it is a content warning note, the reason will be displayed if a reason is set, and the event ID of the note will also be displayed.  
To display a content warning note, run the catEvent subcommand by specifying the note's Event ID in hex.  

### .vimrc sample code to call from vim

``` vimscript
" Usage
"   1. Write content current buffer
"   2. execute next command on command-line
"      : NPM [reason]
"        If reason is given as an argument,
"          it will be published as a content warning note.
"
command! -nargs=? NPM call Pubmessage(<f-args>)
function! Pubmessage(...) abort
    let l:buffer_contents = join(getline(1, '$'), "\n")
    let l:buffer_contents = substitute(l:buffer_contents, '"', '\\"', 'g')
    let l:buffer_contents = substitute(l:buffer_contents, '`', '\\`', 'g')

    if a:0 >= 1
        let l:command = "nostk pubMessage \"" . l:buffer_contents . "\" " . a:1
    else
        let l:command = "nostk pubMessage \"" . l:buffer_contents . "\""
    endif

    let l:command_output = system(l:command)
    echo l:command_output
endfunction
```

``` vimscript
" Usage
"   1. Write json current buffer
"   2. execute next command on command-line
"      : NPR
"
command! NPR call Pubraw()
function Pubraw()
	let l:buffer_contents = join(getline(1, '$'), "\n")
	let l:buffer_contents = substitute(l:buffer_contents, '"', '\\"', 'g')
	let l:command = "nostk pubRaw \"" . l:buffer_contents . "\""
	let l:command_output = system(l:command)
	echo l:command_output
endfunction
```

``` vimscript
" Usage
"   execute next command on command-line
"      : NCathome [number]
"        If you pass number as an argument,
"          it will ask the relay to subscribe to number notes.
"
command! -nargs=? NCatHome call Cathome(<f-args>)
function! Cathome(...)
	if a:0 >= 1
		let l:command = "nostk catHome " . a:1 . " | jq '.'"
		let l:command_output = system(l:command)
		call append('$', split(l:command_output, "\n"))
	else
		r! nostk catHome | jq '.'
	end
	set ft=json
endfunction
```

``` vimscript
" Usage
"   execute next command on command-line
"      : NCatnsfw [number]
"        If you pass number as an argument,
"          it will ask the relay to subscribe to number notes.
"
command! -nargs=? NCatNsfw call Catnsfw(<f-args>)
function! Catnsfw(...)
	if a:0 >= 1
		let l:command = "nostk catNSFW " . a:1 . " | jq '.'"
		let l:command_output = system(l:command)
		call append('$', split(l:command_output, "\n"))
	else
		r! nostk catNSFW | jq '.'
	end
	set ft=json
endfunction
```

``` vimscript
" Usage
"   execute next command on command-line
"      : NCatself [number]
"        If you pass number as an argument,
"          it will ask the relay to subscribe to number notes.
"
command! -nargs=? NCatSelf call Catself(<f-args>)
function! Catself(...)
	if a:0 >= 1
		let l:command = "nostk catSelf " . a:1 . " | jq '.'"
		let l:command_output = system(l:command)
		call append('$', split(l:command_output, "\n"))
	else
		r! nostk catSelf | jq '.'
	end
	set ft=json
endfunction
```

``` vimscript
" Usage
"   execute next command on command-line
"      : NRemoveEvent <Event Id> [reason]
"        Event Id : Specify the Event ID to be deleted
"        reason   : Specify the reason for deletion. Reason is optional.
"
command! -nargs=1 NRemoveEvent call Removeevent(<f-args>)
function! Removeevent(...)
	let l:command = "nostk removeEvent"
	for arg in a:000
		let l:command .= " " . arg
	endfor
	let l:command_output = system(l:command)
	echo l:command_output
endfunction
```

``` vimscript
" Usage
"   1. Move the cursor to the EVENT ID of the post you want to react to
"      in the nostk log DISPLAYED IN THE CURRENT BUFFER.
"   2. execute next command on command-line
"      : NEmojireaction <custom emoji short code>
"        custom emoji short code : Specify a custom emoji shortcode
"
command! -nargs=1 NEmojiReaction call Emojireaction(<f-args>)
function! Emojireaction(stremoji)
	let l:topline = line('.')
	let l:btmline = l:topline + 3
	let l:lines = getline(l:topline, l:btmline)
	let l:lines_text = join(lines, "\n")
	try
		let l:json_data = json_decode("{" . l:lines_text . "}}")
		let l:keys = keys(l:json_data)
		let l:eventId = l:keys[0]
		let l:data = l:json_data[l:eventId]
		let l:pubkey = l:data.pubkey
	catch
		echoerr "Invalid JSON in selected range."
	endtry
	let l:cmd = 'nostk emojiReaction ' . l:eventId . " " . l:pubkey . ' "' . a:stremoji . '"'
	let l:command_output = system(l:cmd)
	echo l:command_output
endfunction
```

