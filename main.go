package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/pkuca/irc"
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC | log.Lmicroseconds)

	host, port, useTLS, minUsers, outputFormat, err := parseFlags()
	if err != nil {
		log.Fatalln(err)
	}

	var (
		handler  = irc.NewEventHandler()
		idString = "gopher_" + strconv.Itoa(rand.Intn(500))
		results  = new(sync.Map)
		client   = &irc.Client{
			Handler:  handler,
			Host:     host,
			Ident:    idString,
			Nickname: idString,
			Port:     port,
			Realname: idString,
			TLS:      useTLS,
		}
	)

	handler.On("001", callback001())
	handler.On("322", callback322(results))
	handler.On("323", callback323(results, minUsers, outputFormat))
	log.Fatalln(client.Run())
}

func callback001() func(c *irc.Client, m *irc.Message) {
	return func(c *irc.Client, m *irc.Message) {
		log.Println("starting scan...")
		c.Write("LIST")
	}
}

func callback322(results *sync.Map) func(c *irc.Client, m *irc.Message) {
	return func(c *irc.Client, m *irc.Message) {
		var (
			fields  = strings.Fields(m.Content)
			channel = fields[1]
			count   = fields[2]
		)
		if userCount, err := strconv.Atoi(count); err != nil {
			fmt.Println("Error during int conversion:", err)
		} else {
			results.Store(channel, userCount)
		}
	}
}

func callback323(results *sync.Map, minUsers int, outputFormat string) func(c *irc.Client, m *irc.Message) {
	return func(c *irc.Client, m *irc.Message) {
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
			os.Exit(0)
		case "csv":
			channelsCSVFormat(filteredChannels)
			os.Exit(0)
		}

	}
}

func parseFlags() (string, string, bool, int, string, error) {
	host := flag.String(
		"host",
		"",
		"required: target server host address",
	)

	port := flag.String(
		"port",
		"",
		"required: target server port",
	)

	tls := flag.Bool(
		"tls",
		false,
		"optional: connect using tls (default: false)",
	)

	minUsers := flag.Int(
		"minusers",
		500,
		"optional: minimum channel population for results (default: 500)",
	)

	outputFormat := flag.String(
		"outputformat",
		"list",
		"optional: can be 'list' or 'csv' (default: list)",
	)

	flag.Parse()

	if *host == "" {
		return "", "", false, 0, "", fmt.Errorf("-host argument is required")
	}

	if *port == "" {
		return "", "", false, 0, "", fmt.Errorf("-port argument is required")
	}

	return *host, *port, *tls, *minUsers, *outputFormat, nil
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
