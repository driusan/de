package redmine

import (
	//"net/http"
	"bytes"
	//"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

//type viewportStub struct{}

// Tests that the plugin parses the output of a request to URL/projects.json
// This is used for the Redmine (no arguments) variation of the plugin.
func TestParseRedmineProjectsJSON(t *testing.T) {
	// Use a string form of the output from http://www.redmine.org/projects.json for
	// a test
	var rpTest string = `{"projects":[{"id":1,"name":"Redmine","identifier":"redmine","description":"Redmine is a flexible project management web application written using Ruby on Rails framework.","status":1,"created_on":"2007-09-29T10:03:04Z","updated_on":"2009-03-15T11:35:11Z"}],"total_count":1,"offset":0,"limit":25}`
	projects, err := parseRedmineProjectJSON(strings.NewReader(rpTest))
	if err != nil {
		t.Errorf("Error while parsing basic Redmine projects.json output %v\n", err)
	}
	if len(projects.Projects) != 1 {
		t.Errorf("Unexpected number of projects parsed. Got %d expected 1\n", len(projects.Projects))
	}
	if projects.Projects[0].Name != "Redmine" {
		t.Errorf("Unxpected name for Project 1. Expected Redmine got %s\n", projects.Projects[0].Name)
	}
	if projects.Projects[0].Identifier != "redmine" {
		t.Errorf("Unxpected name for Project 1. Expected redmine got %s\n", projects.Projects[0].Identifier)
	}
	if projects.Projects[0].Id != 1 {
		t.Errorf("Unxpected name for Project 1. Expected redmine got %d\n", projects.Projects[0].Id)
	}
}

// Tests that the plugin parses the output of a request to URL/projects/identifier/issues.json
// This is used for the Redmine:ProjectName variation of the plugin.
func TestParseRedmineIssuesJSON(t *testing.T) {
	// A hardcoded string that came from http://www.redmine.org/projects/redmine/issues.json at
	// some point of time as a real world test case.
	issuesTest, err := ioutil.ReadFile("redmineissues_test.json")
	if err != nil {
		t.Fatal("Could not load JSON for issues test from redmineissues_test.json\n")
	}
	issues, err := parseRedmineIssueList(bytes.NewReader(issuesTest))
	if err != nil {
		t.Errorf("Error while parsing basic Redmine projects.json output %v\n", err)
	}
	if len(issues.Issues) != 25 {
		t.Errorf("Unexpected number of issues parsed. Expected 25 got %d\n", len(issues.Issues))

	}
	if issues.Issues[0].Id != 22943 {
		t.Errorf("Incorrect Id for first issue. Expect 22943 but got %d\n", issues.Issues[0].Id)
	}
	if issues.Issues[0].Subject != "Transfer " {
		t.Errorf("Unexpected subject for first issue. Expected \"Transfer \" got \"%s\"", issues.Issues[0].Subject)
	}
}

// This tests that the plugin parses the output of a request to URL/issues/issuenum.json
// correctly.
// This is used for the Redmine:IssueNum variation of the plugin
func TestParseRedmineIssueJSON(t *testing.T) {
	issueJSON := `{"issue":{"id":22944,"project":{"id":1,"name":"Redmine"},"tracker":{"id":2,"name":"Feature"},"status":{"id":1,"name":"New"},"priority":{"id":5,"name":"High"},"author":{"id":152527,"name":"Mauricio Palumbo"},"category":{"id":6,"name":"News"},"subject":"EDI Faturamento FIAT","description":"Alterar no EDI de faturamento para que as faturas que cont\u00e9m mais de um pagador sejam gravadas considerando o n\u00famero que diferencia os pagadores. Cada fatura que diferencia um pagador deve corresponder a uma linha do EDI para corre\u00e7\u00e3o das emiss\u00f5es de faturamento","done_ratio":0,"custom_fields":[{"id":2,"name":"Resolution","value":""}],"created_on":"2016-05-31T17:13:46Z","updated_on":"2016-05-31T17:13:46Z"}}`

	issue, err := parseRedmineIssue(strings.NewReader(issueJSON))
	if err != nil {
		t.Fatal("Could not load JSON for test issue.")
	}
	if issue.Id != 22944 {
		t.Errorf("Unexpected Id for issue. Expected 22944 got %d", issue.Id)
	}
	if issue.Subject != "EDI Faturamento FIAT" {
		t.Errorf("Incorrect title for parsed issue.")
	}
	if issue.Description != "Alterar no EDI de faturamento para que as faturas que cont\u00e9m mais de um pagador sejam gravadas considerando o n\u00famero que diferencia os pagadores. Cada fatura que diferencia um pagador deve corresponder a uma linha do EDI para corre\u00e7\u00e3o das emiss\u00f5es de faturamento" {
		t.Errorf("Incorrect description for parsed issue.")
	}

}
