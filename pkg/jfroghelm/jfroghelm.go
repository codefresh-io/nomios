package jfroghelm

import (
	"encoding/json"
	"fmt"
	"github.com/codefresh-io/nomios/pkg/hermes"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// JFrog struct
type JFrogHelm struct {
	hermesSvc hermes.Service
}

type webhookPayload struct {
	Artifactory struct {
		Webhook struct {
			Event string `json:event`
			Data  struct {
				ModifiedBy string `json:modifiedBy`
				Created    int64  `json:created`
				RepoPath   struct {
					RepoKey string `json:repoKey`
					Name    string `json:name`
				} `json:repoPath`
			} `json:data`
		} `json:"webhook"`
	} `json:"artifactory"`
}

// NewJFrog new jfrog handler
func NewJFrog(svc hermes.Service) *JFrogHelm {
	return &JFrogHelm{svc}
}

func constructEventURI(payload *webhookPayload, account string) string {
	uri := fmt.Sprintf("helm:jfrog:%s:%s:push", payload.Artifactory.Webhook.Data.RepoPath.RepoKey, payload.Artifactory.Webhook.Data.RepoPath.Name)
	if account != "" {
		uri = fmt.Sprintf("%s:%s", uri, account)
	}
	return uri
}

// HandleWebhook handle JFrog webhook
func (d *JFrogHelm) HandleWebhook(c *gin.Context) {
	log.Info("Got JFrog Helm webhook event")

	buf := make([]byte, 20024)
	num, _ := c.Request.Body.Read(buf)
	reqBody := string(buf[0:num])

	log.Info("Helm payload " + reqBody)

	payload := webhookPayload{}
	if err := c.BindJSON(&payload); err != nil {
		log.WithError(err).Error("Failed to bind payload JSON to expected structure")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if payload.Artifactory.Webhook.Event != "storage.afterCreate" {
		log.Debug(fmt.Sprintf("Skip event %s", payload.Artifactory.Webhook.Event))
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
	event.Variables["event"] = payload.Artifactory.Webhook.Event
	event.Variables["namespace"] = payload.Artifactory.Webhook.Data.RepoPath.RepoKey
	event.Variables["name"] = payload.Artifactory.Webhook.Data.RepoPath.Name
	event.Variables["pusher"] = payload.Artifactory.Webhook.Data.ModifiedBy
	event.Variables["pushed_at"] = time.Unix(int64(payload.Artifactory.Webhook.Data.Created/1000), 0).Format(time.RFC3339)

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
