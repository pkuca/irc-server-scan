# irc-server-scan

## Installation
```bash
go install github.com/pkuca/irc-server-scan@latest
```

## Basic Usage
```bash
$ irc-server-scan --host some.irc.host --minusers 750
starting scan...
13 results
+-------------+---------+-------------------------------------------------------+
|    NAME     | VISIBLE |                         TOPIC                         |
+-------------+---------+-------------------------------------------------------+
| #archlinux  |    1319 | Welcome to Arch Linux, Good Luck: https://archlinu... |
| #emacs      |     765 | EmacsConf 2021 CFP now open! https://emacsconf.org... |
| #fedora     |    1254 | Fedora Linux F33, F34 end user support || Please R... |
| #gentoo     |     805 | Gentoo Linux Support | Can't speak? /j #gentoo-ops... |
| #gnuradio   |    1066 | GNU Radio — The Free & Open Software Radio Ecosy...   |
| #kde_ru     |    1154 | Русскоязычное сообщество KD...                        |
| #libera     |    2036 | welcome to libera chat support (pls no politics, i... |
| #linux      |    1739 | Welcome to #Linux. Help/support for any Linux dist... |
| #neovim     |    1451 | neovim is a great text editor | https://neovim.io ... |
| #networking |     796 | Computer Networking | If you have a question, just... |
| #pyar       |    2533 | http://python.org.ar | ¿Arrancando con Python? ht...  |
| #python     |    1218 | Anything Python is on-topic. | Don't paste, use h...  |
| #ubuntu     |     967 | Official Ubuntu Support Channel | IRC Guidelines: ... |
+-------------+---------+-------------------------------------------------------+
```

## Options
```
NAME:
   irc-server-scan - scan an irc server for channel populations

USAGE:
   irc-server-scan [global options]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --host value                target server address
   --port value                target server port (default: "6667")
   --minusers value, -m value  only list channels with users exceeding this value (default: 50)
   --topiclength value         truncate channel descriptions in 'list' format to this length (default: 125)
   --help, -h                  show help (default: false)
```