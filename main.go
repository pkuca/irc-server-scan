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
				Usage:    "target server address",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "port",
				Usage:    "target server port",
				Required: true,
				Value:    "6667",
			},
			&cli.IntFlag{
				Name:    "minusers",
				Aliases: []string{"m"},
				Usage:   "filter list results by channel population",
				Value:   50, //nolint:gomnd
			},
			&cli.IntFlag{
				Name:  "topiclength",
				Usage: "shorten channel topics in list",
				Value: 125, //nolint:gomnd
			},
		},
		Action: func(c *cli.Context) error {
			addr := fmt.Sprintf("%s:%s", c.String("host"), c.String("port"))
			conn, err := net.Dial("tcp", addr)
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

			return client.RunContext(context.Background())
		},
	}
}

type channelInfo struct {
	Name    string
	Visible int
	Topic   string
}

func ircHandler(results *sync.Map, c *cli.Context) irc.HandlerFunc {
	options := &struct {
		MinUsers    int
		TopicLength int
	}{
		MinUsers:    c.Int("minusers"),
		TopicLength: c.Int("topiclength"),
	}

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
				channel = m.Params[1]
				visible = m.Params[2]
				topic   = m.Params[3]
			)

			visibleInt, err := strconv.Atoi(visible)
			if err != nil {
				fmt.Println("unable to convert string to int:", err)

				return
			}

			results.Store(channel, &channelInfo{channel, visibleInt, topic})
		case "323":
			// Add channels with more users than `minUsers` to `filteredChannels`.
			filteredChannels := []*channelInfo{}

			results.Range(func(_, info interface{}) bool {
				channelInfo := info.(*channelInfo)
				if channelInfo.Visible > options.MinUsers {
					filteredChannels = append(filteredChannels, channelInfo)
				}

				return true
			})

			// Print total channel count.
			fmt.Println(len(filteredChannels), "results")

			// Sort channels alphabetically.
			sort.Slice(filteredChannels, func(i, j int) bool {
				return filteredChannels[i].Name < filteredChannels[j].Name
			})

			data := [][]string{}
			for _, channelInfo := range filteredChannels {
				data = append(data, []string{
					channelInfo.Name,
					strconv.Itoa(channelInfo.Visible),
					truncateString(channelInfo.Topic, options.TopicLength),
				})
			}

			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Name", "Visible", "Topic"})
			table.SetAutoWrapText(false)
			table.AppendBulk(data)
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
