package redmine

import (
	"io/ioutil"
	"os/user"
	"strings"
	//	"fmt"
)

func parseRedminePluginConfig() {
	u, err := user.Current()
	if err != nil {
		return
	}

	config, err := ioutil.ReadFile(u.HomeDir + "/.de/redmine.ini")
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
