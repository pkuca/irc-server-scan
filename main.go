package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/thoj/go-ircevent"
)

func main() {
	// Parse command line arguments.
	serverFlag := flag.String(
		"server",
		"",
		"required: server to list channel populations for - 'irc.freenode.net:6697'",
	)
	minUserCountFlag := flag.Int(
		"minusers",
		500,
		"optional: minimum channel population for displaying in terminal output - default '500'",
	)

	flag.Parse()

	if *serverFlag == "" {
		fmt.Println("-server argument required. See -h for usage.")
		os.Exit(1)
	}

	randomIntString := strconv.Itoa(rand.Intn(500))
	clientID := "gopher_" + randomIntString

	// Define IRC connection parameters.
	var (
		server = *serverFlag
		nick   = clientID
		user   = clientID
		// listResults = make(map[string]int)
		listResults = new(sync.Map)
	)

	// Initialize IRC connection.
	ircConn := irc.IRC(nick, user)
	ircConn.Debug = false
	ircConn.VerboseCallbackHandler = false
	ircConn.UseTLS = true
	ircConn.TLSConfig = &tls.Config{
		ServerName: strings.Split(*serverFlag, ":")[0],
	}

	// Register "welcome" event callback.
	ircConn.AddCallback("001", func(e *irc.Event) {
		// Send "LIST" command.
		ircConn.SendRaw("LIST")
	})

	// Register LIST item response callback.
	ircConn.AddCallback("322", func(e *irc.Event) {
		// On "322" events, populate `listResults` with channel data.
		channel := e.Arguments[1]
		userCount, err := strconv.Atoi(e.Arguments[2])
		if err != nil {
			fmt.Println("Error during int conversion:", err)
		} else {
			listResults.Store(channel, userCount)
		}
	})

	// Register LIST Complete response callback.
	ircConn.AddCallback("323", func(e *irc.Event) {
		// Add channels with more users than `minChannelUsers` to `filteredChannels`.
		filteredChannels := make([]string, 0)
		listResults.Range(func(channelName, userCount interface{}) bool {
			if userCount.(int) > *minUserCountFlag {
				filteredChannels = append(filteredChannels, channelName.(string))
			}
			return true
		})

		// Sort channels alphabetically.
		sort.Strings(filteredChannels)

		// Log total channel count.
		fmt.Println("Got", len(filteredChannels), "results")

		// Log channel info for all `filteredChannels`.
		for _, channelName := range filteredChannels {
			userCount, ok := listResults.Load(channelName)
			if ok {
				fmt.Println(
					"["+channelName+"]",
					"["+strconv.Itoa(userCount.(int))+"]",
				)
			}
		}

		// Close the IRC connection.
		ircConn.Quit()

	})

	// Start connection goroutines.
	if err := ircConn.Connect(server); err != nil {
		fmt.Println("Connection failed:", err)
	}

	ircConn.Loop()
}
