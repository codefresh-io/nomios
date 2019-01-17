package azure

import (
	"encoding/json"
	"fmt"
	"github.com/codefresh-io/nomios/pkg/hermes"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

// azure struct
type azure struct {
	hermesSvc hermes.Service
}

type webhookPayload struct {
	Action    string `json:action`
	Timestamp string `json:timestamp`
	Target    struct {
		Repository string `json:repository`
		Tag        string `json:tag`
		Name       string
	} `json:target`
	Request struct {
		Host string `json:host`
	} `json:request`
}

// NewAzure new azure handler
func NewAzure(svc hermes.Service) *azure {
	return &azure{svc}
}

func constructEventURI(payload *webhookPayload, account string) string {
	ns := strings.Split(payload.Request.Host, ".")
	uri := fmt.Sprintf("registry:azure:%s:%s:push", ns[0], payload.Target.Name)
	if account != "" {
		uri = fmt.Sprintf("%s:%s", uri, account)
	}
	return uri
}

// HandleWebhook handle azure webhook
func (d *azure) HandleWebhook(c *gin.Context) {
	log.Debug("Got azure webhook event")

	payload := webhookPayload{}

	if err := c.BindJSON(&payload); err != nil {
		log.WithError(err).Error("Failed to bind payload JSON to expected structure")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var s []string = strings.Split(payload.Target.Repository, "/")
	payload.Target.Name = strings.Join(s[1:len(s)], "")

	if payload.Action != "push" {
		log.Debug(fmt.Sprintf("Skip event %s", payload.Action))
		return
	}

	event := hermes.NewNormalizedEvent()
	eventURI := constructEventURI(&payload, c.Query("account"))
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		log.WithError(err).Error("Failed to covert webhook payload structure to JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// keep original JSON
	event.Original = string(payloadJSON)

	// get image push details
	event.Variables["event"] = payload.Action
	ns := strings.Split(payload.Request.Host, ".")
	event.Variables["namespace"] = ns[0]
	event.Variables["name"] = payload.Target.Name
	event.Variables["tag"] = payload.Target.Tag
	//event.Variables["pusher"] = payload.Artifactory.Webhook.Data.Event.ModifiedBy
	event.Variables["pushed_at"] = payload.Timestamp

	// get secret from URL query
	event.Secret = c.Query("secret")

	log.Debug("Event url " + eventURI)

	// invoke trigger
	err = d.hermesSvc.TriggerEvent(eventURI, event)
	if err != nil {
		log.WithError(err).Error("Failed to trigger event pipelines")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}
