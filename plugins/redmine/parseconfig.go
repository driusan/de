package redmine

import (
	"io/ioutil"
	"strings"

	"github.com/driusan/de/demodel"
)

func parseRedminePluginConfig() {
	config, err := ioutil.ReadFile(demodel.ConfigHome() + "/redmine.ini")
	if err != nil {
		return
	}
	lines := strings.Split(string(config), "\n")
	for _, line := range lines {
		val := strings.SplitN(line, "=", 2)
		if len(val) != 2 {
			continue
		}
		switch strings.TrimSpace(val[0]) {
		case "url":
			configurl = strings.TrimSpace(val[1])
		case "apikey":
			apikey = strings.TrimSpace(val[1])
		}
	}
}
