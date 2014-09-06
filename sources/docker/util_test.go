package docker

import (
	docker "github.com/fsouza/go-dockerclient"
	"testing"
)

func TestGetPort(t *testing.T) {
	if p, _ := getPort(docker.Env{"WEB_PORT=8888"}); p != 8888 {
		t.Fail()
	}
}

func TestGetPortInvalid(t *testing.T) {
	e := docker.Env{"WEB_PORT=foobar"}
	if _, err := getPort(e); err != InvalidPortError {
		t.Fail()
	}
}

func TestGetPortNone(t *testing.T) {
	if p, _ := getPort(docker.Env{}); p != 80 {
		t.Fail()
	}
}

func TestGetHostnames(t *testing.T) {
	e := docker.Env{"WEB_HOSTNAME=foo|bar"}
	if h := getHostnames(e); len(h) != 2 || h[0] != "foo" || h[1] != "bar" {
		t.Fail()
	}
}

func TestGetHostnamesOne(t *testing.T) {
	e := docker.Env{"WEB_HOSTNAME=foo"}

	if h := getHostnames(e); len(h) != 1 || h[0] != "foo" {
		t.Fail()
	}
}

func TestGetHostnamesNone(t *testing.T) {
	e := docker.Env{}
	if h := getHostnames(e); len(h) != 0 {
		t.Fail()
	}
}
