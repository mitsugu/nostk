#! /bin/sh
go build nostk.go catEvent.go catHome.go catSelf.go config.go dispHelp.go emojiReaction.go publishMessage.go removeEvent.go
<<EOL
gofmt -w nostk.go
gofmt -w catEvent.go
gofmt -w catHome.go
gofmt -w catSelf.go
gofmt -w config.go
gofmt -w dispHelp.go
gofmt -w emojiReaction.go
gofmt -w publishMessage.go
gofmt -w removeEvent.go
EOL
