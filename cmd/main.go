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
					Name:   "hermes",
					Usage:  "Codefresh Hermes service",
					Value:  "http://local.codefresh.io:9011",
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
					Name:   "port",
					Usage:  "TCP port for the dockerhub provider server",
					Value:  10001,
					EnvVar: "PORT",
				},
				cli.BoolFlag{
					Name:  "dry-run",
					Usage: "do not execute commands, just log",
				},
			},
			Usage: "start nomios webhook handler server",
			Description: `Run DockerHub WebHook handler server. Process and send normalized event payload to the Codefresh Hermes trigger manager service to invoke associated Codefresh pipelines.
			
		Event URI Pattern: registry:dockerhub:{{namespace}}:{{name}}:push`,
			Action: runServer,
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "log-level, l",
			Usage:  "set log level (debug, info, warning(*), error, fatal, panic)",
			Value:  "warning",
			EnvVar: "LOG_LEVEL",
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
	switch level := c.GlobalString("log-level"); level {
	case "debug", "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "info", "INFO":
		log.SetLevel(log.InfoLevel)
	case "warning", "WARNING":
		log.SetLevel(log.WarnLevel)
	case "error", "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "fatal", "FATAL":
		log.SetLevel(log.FatalLevel)
	case "panic", "PANIC":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.WarnLevel)
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
	router := gin.New()
	router.Use(gin.Recovery())
	router.POST("/nomios/dockerhub", gin.Logger(), hub.HandleWebhook)
	router.POST("/dockerhub", gin.Logger(), hub.HandleWebhook)
	// event info route
	router.GET("/nomios/event/:uri/:secret", gin.Logger(), getEventInfo)
	router.GET("/event/:uri/:secret", gin.Logger(), getEventInfo)
	// subscribe/unsubscribe route
	router.POST("/nomios/event/:uri/:secret/:credentials", gin.Logger(), subscribeToEvent)
	router.POST("/event/:uri/:secret/:credentials", gin.Logger(), subscribeToEvent)
	router.DELETE("/nomios/event/:uri/:credentials", gin.Logger(), unsubscribeFromEvent)
	router.DELETE("/event/:uri/:credentials", gin.Logger(), unsubscribeFromEvent)
	// status routes
	router.GET("/nomios/health", getHealth)
	router.GET("/health", getHealth)
	router.GET("/nomios/version", getVersion)
	router.GET("/version", getVersion)
	router.GET("/", getVersion)

	port := c.Int("port")
	log.WithField("port", port).Debug("starting nomios server")
	return router.Run(fmt.Sprintf(":%d", port))
}

func getEventInfo(c *gin.Context) {
	info, err := event.GetEventInfo(PublicDNS, c.Param("uri"), c.Param("secret"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func subscribeToEvent(c *gin.Context) {
	// info, err := event.Subscribe(PublicDNS, c.Param("uri"), c.Param("secret"), c.Param("credentials"))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// c.JSON(http.StatusOK, info)
	c.Status(http.StatusNotImplemented)
}

func unsubscribeFromEvent(c *gin.Context) {
	// info, err := event.Unsubscribe(PublicDNS, c.Param("uri"), c.Param("credentials"))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	// 	return
	// }
	// c.JSON(http.StatusOK, info)
	c.Status(http.StatusNotImplemented)
}

func getHealth(c *gin.Context) {
	c.Status(http.StatusOK)
}

func getVersion(c *gin.Context) {
	c.String(http.StatusOK, version.HumanVersion)
}
