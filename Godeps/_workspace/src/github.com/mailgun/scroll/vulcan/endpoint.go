package vulcan

import (
	"fmt"
	"os"

	"github.com/mailgun/iptools"
)

type Endpoint struct {
	ID   string
	Name string
	URL  string
}

func NewEndpoint(name, listenIP string, listenPort int) (*Endpoint, error) {
	id, err := makeEndpointID(listenPort)
	if err != nil {
		return nil, fmt.Errorf("failed to make endpoint ID: %v", err)
	}
	url, err := makeEndpointURL(listenIP, listenPort)
	if err != nil {
		return nil, fmt.Errorf("failed to make endpoint URL: %v", err)
	}
	return &Endpoint{
		ID:   id,
		Name: name,
		URL:  url,
	}, nil
}

func (e *Endpoint) String() string {
	return fmt.Sprintf("Endpoint(ID=%v, Name=%v, URL=%v)", e.ID, e.Name, e.URL)
}

// Construct an endpoint ID in the format of <hostname>_<port>.
func makeEndpointID(listenPort int) (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v_%v", hostname, listenPort), nil
}

// Construct an endpoint URL by determining the private IP address of the host.
func makeEndpointURL(listenIP string, listenPort int) (string, error) {
	// if an app is listening on a specific IP, use it
	if listenIP != "0.0.0.0" {
		return fmt.Sprintf("http://%v:%v", listenIP, listenPort), nil
	}

	// otherwise find a private IP
	privateIPs, err := iptools.GetPrivateHostIPs()
	if err != nil {
		return "", fmt.Errorf("failed to obtain host's private IPs: %v", err)
	}

	if len(privateIPs) == 0 {
		return "", fmt.Errorf("no host's private IPs are found")
	}

	return fmt.Sprintf("http://%v:%v", privateIPs[0], listenPort), nil
}
