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

	irc "github.com/thoj/go-ircevent"
)

func main() {
	serverAddress, minUsers, outputFormat := parseFlags()

	if serverAddress == "" {
		fmt.Println("-server argument required. See -h for usage.")
		os.Exit(1)
	}

	var (
		idString = "gopher_" + strconv.Itoa(rand.Intn(500))
		results  = new(sync.Map)
	)

	conn := irc.IRC(idString, idString)
	conn.UseTLS = true
	conn.TLSConfig = &tls.Config{ServerName: strings.Split(serverAddress, ":")[0]}
	conn.AddCallback("001", callback001(conn))
	conn.AddCallback("322", callback322(results))
	conn.AddCallback("323", callback323(conn, results, minUsers, outputFormat))

	if err := conn.Connect(serverAddress); err != nil {
		fmt.Println("conn.Connect:", err)
		os.Exit(0)
	}

	conn.Loop()
}

func parseFlags() (string, int, string) {
	serverAddress := flag.String(
		"server",
		"",
		fmt.Sprintf(
			"%v",
			"required: server to list channel populations for",
		),
	)

	minUsers := flag.Int(
		"minusers",
		500,
		fmt.Sprintf(
			"%v %v",
			"optional: minimum channel population for displaying in terminal output",
			"(default: 500)",
		),
	)

	outputFormat := flag.String(
		"outputformat",
		"list",
		fmt.Sprintf(
			"%v",
			"optional: can be 'list' or 'csv' (default: list)",
		),
	)

	flag.Parse()

	return *serverAddress, *minUsers, *outputFormat
}

// callback001 sends a LIST command to the server on a successful connection event.
func callback001(conn *irc.Connection) func(e *irc.Event) {
	return func(e *irc.Event) {
		conn.SendRaw("LIST")
	}
}

// callback322 populates `results` with channel data on LIST response events.
func callback322(results *sync.Map) func(e *irc.Event) {
	return func(e *irc.Event) {
		channel := e.Arguments[1]
		if userCount, err := strconv.Atoi(e.Arguments[2]); err != nil {
			fmt.Println("Error during int conversion:", err)
		} else {
			results.Store(channel, userCount)
		}
	}

}

// callback323 operates on collected channel data, printing populations in a given `outputFormat`.
func callback323(conn *irc.Connection, results *sync.Map, minUsers int, outputFormat string) func(e *irc.Event) {
	return func(e *irc.Event) {
		// Close the IRC connection. Operate on `results` after a clean disconnect.
		conn.Quit()

		// Add channels with more users than `minUsers` to `filteredChannels`.
		filteredChannels := make([]string, 0)
		results.Range(func(channelName, userCount interface{}) bool {
			if userCount.(int) > minUsers {
				filteredChannels = append(filteredChannels, channelName.(string))
			}
			return true
		})

		// Sort channels alphabetically.
		sort.Strings(filteredChannels)

		// Log total channel count.
		fmt.Println("Got", len(filteredChannels), "results")

		switch outputFormat {
		case "list":
			channelsListFormat(filteredChannels, results)
		case "csv":
			channelsCSVFormat(filteredChannels)
		}
	}
}

// channelsListFormat writes channel populations to stdout.
func channelsListFormat(channels []string, results *sync.Map) {
	for _, channelName := range channels {
		if userCount, ok := results.Load(channelName); ok {
			fmt.Println(channelName, "["+strconv.Itoa(userCount.(int))+"]")
		}
	}
}

// channelsCSVFormat writes channel names as comma-separated values.
func channelsCSVFormat(channels []string) {
	csvString := ""
	for _, channelName := range channels {
		csvString += channelName + ","
	}
	fmt.Println(csvString)
}
