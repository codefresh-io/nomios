package event

import (
	"fmt"
	"regexp"
	"strings"
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

const validURI = `^index\.docker\.io:[a-z0-9_-]+:[a-z0-9_-]+:push$`

// compiled validator regexp
var validator, _ = regexp.Compile(validURI)

// GetEventInfo get extended info from uri
func GetEventInfo(publicDNS string, uri string, secret string) (*Info, error) {
	if !validator.MatchString(uri) {
		return nil, fmt.Errorf("unexpected event uri: %s", uri)
	}
	// split uri
	s := strings.Split(uri, ":")
	// dockerhub repository 1st
	repo := s[1]
	// dockerhub image 2nd
	image := s[2]

	// format info
	info := new(Info)
	info.Description = fmt.Sprintf("Docker Hub %s/%s push event", repo, image)
	info.Endpoint = fmt.Sprintf("%s/nomios/dockerhub?secret=%s", publicDNS, secret)
	info.Status = "active"
	info.Help = fmt.Sprintf(`Docker Hub webhooks fire when an image is built in, pushed or a new tag is added to, your repository.

Configure Docker Hub webhooks on https://hub.docker.com/r/%s/%s/~/settings/webhooks/

Add following Codefresh Docker Hub webhook endpoint %s`, repo, image, info.Endpoint)

	// return info
	return info, nil
}
