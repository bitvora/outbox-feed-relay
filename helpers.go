package main

import (
	"regexp"

	"github.com/nbd-wtf/go-nostr/nip19"
)

func nPubToPubkey(nPub string) string {
	_, v, err := nip19.Decode(nPub)
	if err != nil {
		panic(err)
	}
	return v.(string)
}

func isValidWebSocketURL(url string) bool {
	const wsURLPattern = `^wss://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/[a-zA-Z0-9._~%!$&'()*+,;=:@/?-]*)?$`
	re := regexp.MustCompile(wsURLPattern)
	return re.MatchString(url)
}
