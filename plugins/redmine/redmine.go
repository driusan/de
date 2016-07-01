package redmine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/driusan/de/actions"
	"github.com/driusan/de/demodel"
	"io"
	"net/http"
	"strconv"
	"strings"
)

var configurl, apikey string

func init() {
	// sets up the configurl, configuser and configpassword package
	// variables based on ~/.de/redmine.ini
	parseRedminePluginConfig()

	// Look up an issue on Redmine
	actions.RegisterAction("Redmine", Redmine)
}

type redmineIssueList struct {
	Issues []redmineIssue `json:"issues"`
}
type redmineField struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
type redmineIssue struct {
	Id          int `json:"id"`
	Subject     string
	Description string
	Priority    redmineField `json:"priority"`
	Status      redmineField `json:"status"`
	Author      redmineField `json:"author"`
	AssignedTo  redmineField `json:"assigned_to"`

	// CustomFields []redmineField `json:
}

/*
func (ri redmineIssue) GetFieldByName(name string) string {
	for _, f := range ri.Fields {
		if f.Name == name {
			return f.Name
		}
	}
	return ""
}
func (ri redmineIssue) GetFieldById(id int) string {
	for _, f := range ri.Fields {
		if f.Id == id {
			return f.Name
		}
	}
	return ""
}*/

type redmineProjects struct {
	Projects []redmineProject `json:"projects"`
}
type redmineProject struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description"`
}

func parseRedmineIssueList(r io.Reader) (redmineIssueList, error) {
	decoder := json.NewDecoder(r)
	var il redmineIssueList
	for decoder.More() {
		err := decoder.Decode(&il)
		if err != nil {
			return il, err
		}
	}
	return il, nil
}
func parseRedmineProjectJSON(r io.Reader) (redmineProjects, error) {
	decoder := json.NewDecoder(r)
	//decoder := json.NewDecoder(resp.Body)
	var rps redmineProjects
	for decoder.More() {
		err := decoder.Decode(&rps)
		if err != nil {
			return rps, err
		}
	}
	return rps, nil
}

// not a real type, only used because the JSON returned by the Redmine
// API is of the form {"issue" : redmineIssue } instead of directly
// a redmine issue.
type issueHolder struct {
	Issue redmineIssue `json:"issue"`
}

func parseRedmineIssue(r io.Reader) (redmineIssue, error) {
	// temporary struct because the JSON is of the form issue : redmineIssue
	// and not directly just a redmineIssue.
	var ih issueHolder

	decoder := json.NewDecoder(r)
	for decoder.More() {
		err := decoder.Decode(&ih)
		if err != nil {
			return ih.Issue, err
		}
	}
	return ih.Issue, nil
}

// Redmine looks up an issue on Redmine. It can be called with or without arguments.
// The default Redmine command will list projects on the server.
// Redmine:string will look up the issues associated with the project identified
// by string
// Redmine:integer will look up a particular issue number on the server.
// The server that is connected to will be the one set up (with the credentials from)
// ~/.de/redmine.ini
// which should have the format:
// url=URL
// username=user
// password=password
// Where username and password will be sent to the Redmine API with HTTP Basic authentication
// If they're empty, it will attempt to use the API anonymously, and fail if the redmine
// server is private.
func Redmine(args string, buff *demodel.CharBuffer, v demodel.Viewport) {
	if v == nil {
		return
	}
	if configurl == "" {
		buff.AppendTag("\nMissing configuration for Redmine plugin.")
		return
	}
	if args == "" {
		// Basic usage of the plugin, just get a list of projects.
		url := fmt.Sprintf("%s/projects.json?key=%s", configurl, apikey)

		resp, err := http.Get(url)
		if err != nil {
			buff.AppendTag("\n")
			buff.AppendTag(err.Error())
			return
		}
		if err != nil {
			buff.AppendTag(err.Error())
			return
		}
		defer resp.Body.Close()
		projects, err := parseRedmineProjectJSON(resp.Body)
		if err != nil {
			buff.AppendTag(err.Error())
			return
		}
		var newBuffer bytes.Buffer
		for _, project := range projects.Projects {
			fmt.Fprintf(&newBuffer, "Redmine:%s %s\n", project.Identifier, project.Description)
		}
		buff.Filename = ""
		buff.Buffer = newBuffer.Bytes()
		buff.Dot.Start = 0
		buff.Dot.End = 0
		v.ResetLocation()
		v.Rerender()
		return
	}

	cmd := args
	var params []string
	if i := strings.Index(args, ","); i >= 0 {
		cmd = args[0:i]
		params = strings.Split(args[i:], ",")
	}
	if issuenum, err := strconv.Atoi(cmd); err == nil {
		// Looking up a specific issue for id issuenum
		url := fmt.Sprintf("%s/issues/%d.json?key=%s", configurl, issuenum, apikey)
		for _, s := range params {
			url += "&" + s
		}
		resp, err := http.Get(url)
		if err != nil {
			buff.AppendTag("\n")
			buff.AppendTag(err.Error())
			return
		}
		defer resp.Body.Close()
		var newBuffer bytes.Buffer
		issue, err := parseRedmineIssue(resp.Body)
		if err != nil {
			buff.AppendTag("\n")
			buff.AppendTag(err.Error())
			return
		}
		fmt.Fprintf(&newBuffer, "Full URL: %s\nRedmine:%d\nAuthor:%s\nAssignee:%s\nPriority: %s\nStatus:%s\nSubject:%s\n\nDescription:\n%s\n", strings.TrimSuffix(url, ".json"+"?apikey="+apikey)+"/", issue.Id, issue.Author.Name, issue.AssignedTo.Name, issue.Priority.Name, issue.Status.Name, issue.Subject, strings.Replace(issue.Description, "\r", "", -1))
		buff.Filename = ""
		buff.Buffer = newBuffer.Bytes()

	} else {
		// Couldn't convert to args to an integer, so args is actually a project name.
		// Look up the issues for that project, and add any , separated parameters to
		// the URL request.
		url := fmt.Sprintf("%s/projects/%s/issues.json?key=%s", configurl, cmd, apikey)
		for _, s := range params {
			url += "&" + s
		}

		resp, err := http.Get(url)
		if err != nil {
			buff.AppendTag("\n")
			buff.AppendTag(err.Error())
			return
		}
		defer resp.Body.Close()
		issues, err := parseRedmineIssueList(resp.Body)
		if err != nil {
			buff.AppendTag("\n")
			buff.AppendTag(err.Error())
			return
		}
		var newBuffer bytes.Buffer
		// include the command that was used at the top, so it can just be modified to get
		// somewhere else.
		fmt.Fprintf(&newBuffer, "Redmine:%s\n", args)
		if strings.Contains(args, "assigned_to") {
			fmt.Fprintf(&newBuffer, "Redmine:%s,assigned_to_id=me\n", args)
		}
		if strings.Contains(args, "limit") {
			fmt.Fprintf(&newBuffer, "Redmine:%s,limit=50,offset=0\n", args)
		}
		if strings.Contains(args, "sort") {
			fmt.Fprintf(&newBuffer, "Redmine:%s,sort=priority\n", args)
		}
		fmt.Fprintf(&newBuffer, "\n")
		for _, issue := range issues.Issues {
			fmt.Fprintf(&newBuffer, "Redmine:%d (%s)\n", issue.Id, issue.Subject)
		}
		buff.Filename = ""
		buff.Buffer = newBuffer.Bytes()

	}

	v.ResetLocation()
	v.Rerender()
}
