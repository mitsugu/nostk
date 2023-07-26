nostk
========
Implementing a CLI client to use [Nostr Protocol](https://github.com/nostr-protocol/nostr).

### Environment
* Ubuntu 23.04 and later
* Go Language 1.20.5 and later

### Features
* Initializing the nostk environment
* Generating a key pair
* Edit your contact list
* Edit custom emoji list
* Edit relay list
* Publish relay list
* Edit profile
* Publish profile
* Display home timeline

### ToDo
* Mention to any user
* Publishing a message with message citations
* Log viewer
* any more

### Requirements
* [nbd-wtf / go-nostr](https://github.com/nbd-wtf/go-nostr)
* Some kind of text editor
* Setting $HOME environment variable
* Setting $EDITOR environment variable

### Install nostk:
```bash
go install github.com/mitsugu/nostk@latest
```

### Usage
#### Display help documanets
``` bash
nostk help

nostk -h

nostk --help

nostk
```

#### Init nostk
``` bash
nostk init
```

#### Ganerate Key Pair
``` bash
nostk genkey
```

#### Edit contact list
``` bash
nostk editContacts
```

#### Edit custom emoji
``` bash
nostk editEmoji
```

#### Edit relay list
``` bash
nostk editRelays
```

#### Publish relay list
``` bash
nostk pubRelays
```

#### Edit profile
``` bash
nostk editProfile
```

#### Publish profile
``` bash
nostk pubProfile
```

#### Publish message
``` bash
nostk pubMessage <text message>

nostk pubMessage < (ps)

(ps) | nostk pubMessage
```

#### Display home timeline
``` bash
nostk dispHome [number] [date_time]
```

