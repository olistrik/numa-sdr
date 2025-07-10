package main

import (
	"bufio"
	"embed"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/gin-gonic/gin"
	"github.com/olistrik/numa-sdr/api/cmd/web/internal/filesystem"
	"github.com/olistrik/numa-sdr/api/cmd/web/internal/sse"
	"github.com/olistrik/numa-sdr/api/sdr/power"
	power_history "github.com/olistrik/numa-sdr/api/sdr/power/history"
	"github.com/olistrik/numa-sdr/api/unit"
	log "github.com/sirupsen/logrus"
)

var args struct {
	Address string         `arg:"-a" default:"0.0.0.0" placeholder:"ip"`
	Port    string         `arg:"-p" default:"21753" placeholder:"port"`
	Offset  unit.Frequency `arg:"-o" default:"0" placeholder:"float"`
	History time.Duration  `arg:"--history" default:"1h" placeholder:"duration"`
	Title   string         `arg:"-t" default:"Numa" placeholder:"string"`
}

func processRow(callback func(*power.Scan) error) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// scanner reads lines by default.
		line := scanner.Text()

		// forward line
		fmt.Println(line)

		// parse the line
		scan, err := power.ParseScan(line)
		if err != nil {
			log.Errorln(err)
			continue
		}

		scan.StartFrequency += args.Offset
		scan.EndFrequency += args.Offset

		// call the callback with the result.
		if err := callback(scan); err != nil {
			log.Errorln(err)
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		log.Errorln(err)
	}
}

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {

	gin.DefaultWriter = log.StandardLogger().Out
	arg.MustParse(&args)

	var stream *sse.Stream

	hm := power_history.New(
		power_history.MaxDuration(args.History),
	)

	stream = sse.NewStream(
		sse.OnConnect(func(client sse.Client) {
			client.Send("init", hm.Scans)
		}),
	)

	go processRow(func(scan *power.Scan) error {
		complete, err := hm.Push(scan)
		if err != nil {
			return err
		}

		if complete {
			stream.Send("scan", hm.Head())
		}

		return nil
	})

	log.Info("Starting webserver...")

	router := gin.Default()

	// FIX: this is temporary, LoadHTMLFS has been merged into gin, but not yet released.
	filesystem.LoadHTMLFS(router, http.FS(templatesFS), "**/*.html.tmpl")

	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html.tmpl", gin.H{
			"title": args.Title,
		})
	})

	router.GET("/favicon.ico", func(c *gin.Context) {
		data, _ := staticFS.ReadFile("static/favicon.ico")
		c.Data(200, "image/x-icon", data)
	})

	router.GET("/stream/scans", stream.Handler())

	router.Run(fmt.Sprint(args.Address, ":", args.Port))
}
