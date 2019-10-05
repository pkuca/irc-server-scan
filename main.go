package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/pkuca/irc"
)

func main() {
	host, port, tls, minUsers, format, err := parseFlags()
	if err != nil {
		fmt.Println("parseFlags:", err)
		os.Exit(1)
	}

	var (
		results = new(sync.Map)
		opts    = &irc.ClientOptions{
			Host:      host,
			Port:      port,
			EnableTLS: tls,
		}
		client = initClient(opts, format, minUsers, results)
	)

	if err := client.Connect(); err != nil {
		fmt.Println("client.Connect:", err)
		os.Exit(1)
	}

	if err := client.Run(); err != nil {
		fmt.Println("client.Run:", err)
		os.Exit(1)
	}
}

func initClient(ircOptions *irc.ClientOptions, format string, minUsers int, results *sync.Map) *irc.Client {
	idString := fmt.Sprintf("g_%v", rand.Intn(500))

	ircOptions.Ident = idString
	ircOptions.Nickname = idString
	ircOptions.Realname = idString

	client := &irc.Client{
		Options: ircOptions,
		Handler: irc.NewEventHandler(),
	}

	client.Handler.On("001", callback001())
	client.Handler.On("322", callback322(results))
	client.Handler.On("323", callback323(results, minUsers, format))

	return client
}

func callback001() func(c *irc.Client, m *irc.Message) {
	return func(c *irc.Client, m *irc.Message) {
		fmt.Println("starting scan...")
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
			fmt.Println("couldn't convert string to int:", err)
		} else {
			results.Store(channel, userCount)
		}
	}
}

func callback323(results *sync.Map, minUsers int, format string) func(c *irc.Client, m *irc.Message) {
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

		// Print total channel count.
		fmt.Println("Got", len(filteredChannels), "results")

		switch format {
		case "list":
			channelsListFormat(filteredChannels, results)
			if err := c.Close(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			os.Exit(0)
		case "csv":
			channelsCSVFormat(filteredChannels)
			if err := c.Close(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}
}

// parseFlags obtains irc client options from the command line
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
		"optional: connect using tls",
	)

	minUsers := flag.Int(
		"minusers",
		500,
		"optional: minimum channel population for results",
	)

	format := flag.String(
		"format",
		"list",
		"optional: can be 'list' or 'csv'",
	)

	flag.Parse()

	if *host == "" {
		return "", "", false, 0, "", fmt.Errorf("-host argument is required")
	}

	if *port == "" {
		return "", "", false, 0, "", fmt.Errorf("-port argument is required")
	}

	return *host, *port, *tls, *minUsers, *format, nil
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
