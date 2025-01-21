package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fiatjaf/khatru"
	"github.com/joho/godotenv"
	"github.com/nbd-wtf/go-nostr"
)

var pool = nostr.NewSimplePool(context.Background())

type CacheEntry struct {
	Data       []string
	Expiration time.Time
}

var (
	relayCache    = make(map[string]CacheEntry)
	followCache   = make(map[string]CacheEntry)
	cacheMutex    sync.Mutex
	cacheDuration = 1 * time.Hour
)

func main() {
	_ = godotenv.Load(".env")
	discoveryRelays := GetDiscoveryRelays()
	fmt.Printf("Discovery relays: %v\n", discoveryRelays)

	http.HandleFunc("/", dynamicRelayHandler)

	addr := fmt.Sprintf("%s:%d", "0.0.0.0", 3334)

	log.Printf("🔗 listening at %s", addr)
	http.ListenAndServe(addr, nil)
}

func dynamicRelayHandler(w http.ResponseWriter, r *http.Request) {
	npub := r.URL.Path
	npub = strings.TrimPrefix(npub, "/")

	npubRelay := khatru.NewRelay()
	npubRelay.Info.Description = "utxo's algo relay"
	npubRelay.Info.Name = "utxo's algo relay"
	npubRelay.Info.PubKey = "e2ccf7cf20403f3f2a4a55b328f0de3be38558a7d5f33632fdaaefc726c1c8eb"
	npubRelay.Info.Software = "https://github.com/bitvora/outbox-relay"
	npubRelay.Info.Version = "0.1.1"
	npubRelay.Info.Icon = "https://i.nostr.build/6G6wW.gif"

	follows := GetFollows(npub, GetDiscoveryRelays())
	relays := GetFollowsRelays(npub, follows, GetDiscoveryRelays())

	npubRelay.QueryEvents = append(npubRelay.QueryEvents, func(ctx context.Context, filter nostr.Filter) (chan *nostr.Event, error) {
		ch := make(chan *nostr.Event)
		npubFilter := []nostr.Filter{{
			Authors: filter.Authors,
			Kinds:   filter.Kinds,
			Tags:    filter.Tags,
			Limit:   filter.Limit,
		}}

		go func() {
			defer close(ch)

			ctx := context.Background()

			for ev := range pool.SubMany(ctx, relays, npubFilter) {
				ch <- ev.Event
			}
		}()

		return ch, nil
	})

	npubRelay.ServeHTTP(w, r)
}

func GetFollowsRelays(npub string, follows, discoveryRelays []string) []string {
	cacheKey := "follows_relays_" + npub

	cacheMutex.Lock()
	if entry, found := relayCache[cacheKey]; found && time.Now().Before(entry.Expiration) {
		cacheMutex.Unlock()
		return entry.Data
	}
	cacheMutex.Unlock()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	filters := []nostr.Filter{{
		Kinds:   []int{nostr.KindRelayListMetadata},
		Authors: follows,
	}}

	relaySet := make(map[string]struct{})
	var relays []string

	for ev := range pool.SubMany(ctx, discoveryRelays, filters) {
		for _, tag := range ev.Event.Tags.GetAll([]string{"r"}) {
			url := tag.Value()
			if isValidWebSocketURL(url) {
				if _, exists := relaySet[url]; !exists {
					relaySet[url] = struct{}{}
					relays = append(relays, url)
				}
			}
		}
	}

	cacheMutex.Lock()
	relayCache[cacheKey] = CacheEntry{
		Data:       relays,
		Expiration: time.Now().Add(cacheDuration),
	}
	cacheMutex.Unlock()

	return relays
}

func GetFollows(npub string, discoveryRelays []string) []string {
	cacheKey := "follows_" + npub

	cacheMutex.Lock()
	if entry, found := followCache[cacheKey]; found && time.Now().Before(entry.Expiration) {
		cacheMutex.Unlock()
		return entry.Data
	}
	cacheMutex.Unlock()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	filters := []nostr.Filter{{
		Kinds:   []int{nostr.KindFollowList},
		Authors: []string{nPubToPubkey(npub)},
		Limit:   1,
	}}

	var followedPubkeys []string
	for ev := range pool.SubMany(ctx, discoveryRelays, filters) {
		for _, tag := range ev.Event.Tags.GetAll([]string{"p"}) {
			followedPubkeys = append(followedPubkeys, tag.Value())
		}

		if len(followedPubkeys) > 0 {
			break
		}
	}

	cacheMutex.Lock()
	followCache[cacheKey] = CacheEntry{
		Data:       followedPubkeys,
		Expiration: time.Now().Add(cacheDuration),
	}
	cacheMutex.Unlock()

	return followedPubkeys
}

func GetDiscoveryRelays() []string {
	filePath := "discovery_relays.json"

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return nil
	}

	var relays []string
	if err := json.Unmarshal(data, &relays); err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil
	}

	return relays
}
