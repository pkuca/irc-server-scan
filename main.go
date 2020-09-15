package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"

	"github.com/go-irc/irc"
	"github.com/urfave/cli/v2"
)

func main() {
	app := NewApp()
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func NewApp() *cli.App {
	return &cli.App{
		Name:      "irc-server-scan",
		Usage:     "scan an irc server for channel populations",
		UsageText: "irc-server-scan [global options]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "host",
				Usage:    "target server address",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "port",
				Usage:    "target server port",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "can also be 'csv'",
				Value:   "list",
			},
			&cli.IntFlag{
				Name:    "minusers",
				Aliases: []string{"m"},
				Usage:   "only list channels with users exceeding this value",
				Value:   500, //nolint:gomnd
			},
		},
		Action: func(c *cli.Context) error {
			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", c.String("host"), c.String("port")))
			if err != nil {
				return err
			}

			var (
				results  = new(sync.Map)
				idString = fmt.Sprintf("g_%v", rand.Intn(500)) //nolint:gomnd,gosec
			)

			config := irc.ClientConfig{
				Nick:    idString,
				Pass:    "",
				User:    idString,
				Name:    idString,
				Handler: ircHandler(results, c),
			}

			client := irc.NewClient(conn, config)
			if err := client.RunContext(context.Background()); err != nil {
				return err
			}

			return nil
		},
	}
}

func ircHandler(results *sync.Map, c *cli.Context) irc.HandlerFunc {
	return func(cl *irc.Client, m *irc.Message) {
		switch m.Command {
		case "001":
			fmt.Println("starting scan...")

			if err := cl.Write("LIST"); err != nil {
				fmt.Println("error writing LIST command:", err)
				os.Exit(1)
			}
		case "322":
			var (
				fields  = m.Params
				channel = fields[1]
				count   = fields[2]
			)

			if userCount, err := strconv.Atoi(count); err != nil {
				fmt.Println("unable to convert string to int:", err)
			} else {
				results.Store(channel, userCount)
			}
		case "323":
			// Add channels with more users than `minUsers` to `filteredChannels`.
			filteredChannels := make([]string, 0)

			results.Range(func(channelName, userCount interface{}) bool {
				if userCount.(int) > c.Int("minusers") {
					filteredChannels = append(filteredChannels, channelName.(string))
				}

				return true
			})

			// Sort channels alphabetically.
			sort.Strings(filteredChannels)

			// Print total channel count.
			fmt.Println("Got", len(filteredChannels), "results")

			switch c.String("format") {
			case "list":
				for _, channelName := range filteredChannels {
					if userCount, ok := results.Load(channelName); ok {
						fmt.Println(channelName, "["+strconv.Itoa(userCount.(int))+"]")
					}
				}

				os.Exit(0)
			case "csv":
				csvString := ""
				for _, channelName := range filteredChannels {
					csvString += channelName + ","
				}

				fmt.Println(csvString)
				os.Exit(0)
			}
		}
	}
}
