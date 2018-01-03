package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/codefresh-io/nomios/pkg/dockerhub"
	"github.com/codefresh-io/nomios/pkg/event"
	"github.com/codefresh-io/nomios/pkg/hermes"
	"github.com/codefresh-io/nomios/pkg/version"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// HermesDryRun dry run stub
type HermesDryRun struct {
}

// PublicDNS public dns name for Codefresh environment
var PublicDNS string

// TriggerEvent dry run version
func (m *HermesDryRun) TriggerEvent(eventURI string, event *hermes.NormalizedEvent) error {
	fmt.Println(eventURI)
	fmt.Println("\tSecret: ", event.Secret)
	fmt.Println("\tVariables:")
	for k, v := range event.Variables {
		fmt.Println("\t\t", k, "=", v)
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "nomios"
	app.Authors = []cli.Author{{Name: "Alexei Ledenev", Email: "alexei@codefresh.io"}}
	app.Version = version.HumanVersion
	app.EnableBashCompletion = true
	app.Usage = "handle DockerHub webhook payload"
	app.UsageText = fmt.Sprintf(`Run DockerHub WebHook handler server.
%s
nomios respects following environment variables:

   - HERMES_SERVICE     - set the url to the Hermes service (default "hermes")
   
Copyright © Codefresh.io`, version.ASCIILogo)
	app.Before = before

	app.Commands = []cli.Command{
		{
			Name: "server",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "hermes, hm",
					Usage:  "Codefresh Hermes service",
					Value:  "http://hermes/",
					EnvVar: "HERMES_SERVICE",
				},
				cli.StringFlag{
					Name:   "token, t",
					Usage:  "Codefresh Hermes API token",
					Value:  "TOKEN",
					EnvVar: "HERMES_TOKEN",
				},
				cli.StringFlag{
					Name:   "dns, n",
					Usage:  "Public DNS name for the Codefresh environment",
					Value:  "https://g.codefresh.io",
					EnvVar: "PUBLIC_DNS_NAME",
				},
				cli.IntFlag{
					Name:  "port",
					Usage: "TCP port for the dockerhub provider server",
					Value: 8080,
				},
			},
			Usage: "start nomios webhook handler server",
			Description: `Run DockerHub WebHook handler server. Process and send normalized event payload to the Codefresh Hermes trigger manager service to invoke associated Codefresh pipelines.
			
		Event URI Pattern: index.docker.io:<namespace>:<name>:push`,
			Action: runServer,
		},
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "enable debug mode with verbose logging",
			EnvVar: "DEBUG_NOMIOS",
		},
		cli.BoolFlag{
			Name:  "dry-run",
			Usage: "do not execute commands, just log",
		},
		cli.BoolFlag{
			Name:  "json",
			Usage: "produce log in JSON format: Logstash and Splunk friendly",
		},
	}

	app.Run(os.Args)

}

func before(c *cli.Context) error {
	// set debug log level
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	}
	// set log formatter to JSON
	if c.GlobalBool("json") {
		log.SetFormatter(&log.JSONFormatter{})
	}

	return nil
}

// start trigger manager server
func runServer(c *cli.Context) error {
	fmt.Println()
	fmt.Println(version.ASCIILogo)

	// bind dockerhub to hermes API endpoint
	var hub *dockerhub.DockerHub
	if c.Bool("dry-run") {
		hub = dockerhub.NewDockerHub(&HermesDryRun{})
	} else {
		// add http protocol, if missing
		hermesSvcName := c.String("hermes")
		if !strings.HasPrefix(hermesSvcName, "http://") {
			hermesSvcName = "http://" + hermesSvcName
		}
		hub = dockerhub.NewDockerHub(hermes.NewHermesEndpoint(hermesSvcName, c.String("token")))
	}

	// get public DNS name
	PublicDNS = c.String("dns")

	// setup gin router
	router := gin.Default()
	router.POST("/nomios/dockerhub", hub.HandleWebhook)
	router.POST("/dockerhub", hub.HandleWebhook)
	// event info route
	router.GET("/nomios/event-info/:uri", getEventInfo)
	router.GET("/event-info/:uri", getEventInfo)
	// status routes
	router.GET("/nomios/health", getHealth)
	router.GET("/health", getHealth)
	router.GET("/nomios/version", getVersion)
	router.GET("/version", getVersion)
	router.GET("/", getVersion)
	router.Run(fmt.Sprintf(":%d", c.Int("port")))
	return nil
}

func getEventInfo(c *gin.Context) {
	info, err := event.GetEventInfo(PublicDNS, c.Param("uri"), c.Param("secret"))
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func getHealth(c *gin.Context) {
	c.Status(http.StatusOK)
}

func getVersion(c *gin.Context) {
	c.String(http.StatusOK, version.HumanVersion)
}
