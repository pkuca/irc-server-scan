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

	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"gopkg.in/irc.v3"
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
				Usage:    "required: target server address",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "port",
				Usage: "target server port",
				Value: "6667",
			},
			&cli.IntFlag{
				Name:  "minusers",
				Usage: "filter list results by channel population",
				Value: 500, //nolint:gomnd
			},
			&cli.StringFlag{
				Name:  "sort",
				Usage: `sorting method - can also be "pop" to sort by channel population`,
				Value: "alpha",
			},
			&cli.IntFlag{
				Name:  "topiclength",
				Usage: "shorten channel topics in list",
				Value: 125, //nolint:gomnd
			},
		},
		Action: func(cliContext *cli.Context) error {
			addr := fmt.Sprintf("%s:%s", cliContext.String("host"), cliContext.String("port"))
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				return fmt.Errorf("irc client tcp connection failure: %w", err)
			}

			var (
				results  = new(sync.Map)
				idString = fmt.Sprintf("g_%v", rand.Intn(500)) //nolint:gomnd,gosec
			)

			handlerOpts := &handlerOptions{
				MinUsers:    cliContext.Int("minusers"),
				Sort:        cliContext.String("sort"),
				TopicLength: cliContext.Int("topiclength"),
			}

			config := irc.ClientConfig{
				Nick:    idString,
				Pass:    "",
				User:    idString,
				Name:    idString,
				Handler: ircHandler(results, handlerOpts),
			}

			client := irc.NewClient(conn, config)

			return fmt.Errorf("irc client run context failure: %w", client.RunContext(context.Background()))
		},
	}
}

// channelInfo represents the structure of the result list table.
type channelInfo struct {
	Name    string
	Visible int
	Topic   string
}

type handlerOptions struct {
	MinUsers    int
	Sort        string
	TopicLength int
}

func ircHandler(results *sync.Map, options *handlerOptions) irc.HandlerFunc {
	return func(cl *irc.Client, message *irc.Message) {
		switch message.Command {
		case "001":
			fmt.Println("starting scan...")

			if err := cl.Write("LIST"); err != nil {
				fmt.Println("error writing LIST command:", err)
				os.Exit(1)
			}
		case "322":
			var (
				channel = message.Params[1]
				visible = message.Params[2]
				topic   = message.Params[3]
			)

			visibleInt, err := strconv.Atoi(visible)
			if err != nil {
				fmt.Println("unable to convert string to int:", err)

				return
			}

			results.Store(channel, &channelInfo{channel, visibleInt, topic})
		case "323":
			filteredChannels := []*channelInfo{}

			// Show channels with `minUsers` present - adjustable by user.
			results.Range(func(_, info interface{}) bool {
				channelInfo, ok := info.(*channelInfo)
				if !ok {
					fmt.Println("channelInfo type assertion failed")
				}
				if channelInfo.Visible >= options.MinUsers {
					filteredChannels = append(filteredChannels, channelInfo)
				}

				return true
			})

			fmt.Println(len(filteredChannels), "results")

			switch options.Sort {
			case "alpha":
				sort.Slice(filteredChannels, func(i, j int) bool {
					return filteredChannels[i].Name < filteredChannels[j].Name
				})
			case "pop":
				sort.Slice(filteredChannels, func(i, j int) bool {
					return filteredChannels[i].Visible > filteredChannels[j].Visible
				})
			}

			// Create table for rendering.
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Visible", "Topic"})
			table.SetAutoWrapText(false)

			// Append records to table for rendering.
			for _, channelInfo := range filteredChannels {
				table.Append([]string{
					channelInfo.Name,
					strconv.Itoa(channelInfo.Visible),
					truncateString(channelInfo.Topic, options.TopicLength),
				})
			}

			table.Render()

			os.Exit(0)
		}
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return fmt.Sprintf("%s...", s[:maxLen])
}
