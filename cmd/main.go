package main

import (
	"fmt"
	"os"

	"github.com/codefresh-io/dockerhub-provider/pkg/dockerhub"
	"github.com/codefresh-io/dockerhub-provider/pkg/hermes"
	"github.com/codefresh-io/hermes/pkg/version"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "dockerhub-provider"
	app.Authors = []cli.Author{{Name: "Alexei Ledenev", Email: "alexei@codefresh.io"}}
	app.Version = version.HumanVersion
	app.EnableBashCompletion = true
	app.Usage = "handle DockerHub webhook payload"
	app.UsageText = fmt.Sprintf(`Run DockerHub WebHook handler server.
%s
dockerhub-provider respects following environment variables:
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
			},
			Usage:       "start dockerhub-provider webhook handler server",
			Description: "Run DockerHub WebHook handler server. Proccess and send normalized event payload to the Codefresh Hermes trigger manager service to invoke associated Codefresh pipelines.",
			Action:      runServer,
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
	fmt.Println(version.ASCIILogo)

	// bind dockerhub to hermes API endpoint
	dockerhub := dockerhub.NewDockerHub(hermes.NewHermesEndpoint(c.String("hermes"), c.String("hermes")))

	// setup gin router
	router := gin.Default()
	router.POST("/dockerhub", dockerhub.HandleWebhook)
	router.Run()
	return nil
}
