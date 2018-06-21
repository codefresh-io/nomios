package event

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

type (
	// Info event info - EVERYTHING for specific event (eventURI)
	Info struct {
		// Endpoint URL
		Endpoint string `json:"endpoint,omitempty"`
		// Description human readable text
		Description string `json:"description,omitempty"`
		// Status current event handler status (active, error, not active)
		Status string `json:"status,omitempty"`
		// Help test
		Help string `json:"help,omitempty"`
	}
)

const validURI = `^registry:dockerhub:[a-z0-9_-]+:[a-z0-9_-]+:push(:[[:xdigit:]]{12})$`

// compiled validator regexp
var validator, _ = regexp.Compile(validURI)

// GetEventInfo get extended info from uri
func GetEventInfo(publicDNS string, uri string, secret string) (*Info, error) {
	log.WithFields(log.Fields{
		"event-uri": uri,
		"validator": validURI,
	}).Debug("get trigger-event info")
	if !validator.MatchString(uri) {
		log.Error("failed to match event URI")
		return nil, fmt.Errorf("unexpected event uri: %s", uri)
	}
	// split uri
	s := strings.Split(uri, ":")
	// dockerhub repository 1st
	repo := s[2]
	// dockerhub image 2nd
	image := s[3]
	// get account hash (may not exist)
	var account string
	if len(s) == 6 {
		account = s[5]
	}

	// format info
	info := new(Info)
	info.Description = fmt.Sprintf("Docker Hub %s/%s push event", repo, image)
	// handle endpoint url
	u, err := url.Parse(publicDNS)
	if err != nil {
		log.WithError(err).WithField("dns", publicDNS).Warn("failed to parse public dns")
	} else {
		q := u.Query()
		q.Set("secret", secret)
		if account != "" {
			q.Set("account", account)
		}
		u.Path = "nomios/dockerhub" + u.Path
		u.RawQuery = q.Encode()
		info.Endpoint = u.String()
	}
	info.Status = "active"
	info.Help = fmt.Sprintf(`Docker Hub webhooks fire when an image is built in, pushed or a new tag is added to, your repository.

Configure Docker Hub webhooks on https://hub.docker.com/r/%s/%s/~/settings/webhooks/

Add following Codefresh Docker Hub webhook endpoint %s`, repo, image, info.Endpoint)

	// return info
	return info, nil
}

// Subscribe to event in DockerHub
func Subscribe(publicDNS, uri, eventType, kind, secret string, values, credentials map[string]string) (*Info, error) {
	return nil, nil
}

// Unsubscribe from event in DockerHub
func Unsubscribe(publicDNS, uri, credentials string) (*Info, error) {
	return nil, nil
}
