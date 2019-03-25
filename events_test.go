package sentry

import (
	"fmt"
	"testing"

	"github.com/getsentry/raven-go"
	"github.com/quantonganh/go-sentry-api/datatype"
)

func TestEventsResource(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	org, err := client.GetOrganization(getDefaultOrg())
	if err != nil {
		t.Fatal(err)
	}

	team, cleanup := createTeamHelper(t)
	defer cleanup()

	project, cleanupproj := createProjectHelper(t, team)
	defer cleanupproj()

	dsnkey, err := client.CreateClientKey(org, project, "testing key")
	if err != nil {
		t.Fatal(err)
	}

	ravenClient, err := raven.New(dsnkey.DSN.Secret)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		ravenClient.CaptureMessageAndWait(fmt.Sprintf("Testing message %d", i), map[string]string{
			"server": fmt.Sprintf("%d-app-node", i),
		}, &raven.Message{
			Message: fmt.Sprintf("Testing some stuff out Brah %d", i),
		})
	}

	issues, _, err := client.GetIssues(org, project, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(issues) == 0 {
		t.Fatalf("no issues found for project %s", project.Name)
	}

	t.Run("Read out all events for a issue", func(t *testing.T) {

		for _, issue := range issues {

			_, err = client.GetLatestEvent(issue)
			if err != nil {
				t.Error(err)
			}

			_, err = client.GetOldestEvent(issue)
			if err != nil {
				t.Error(err)
			}

		}

		t.Run("Convert all entries to proper types if we can", func(t *testing.T) {

			oldest, err := client.GetOldestEvent(issues[0])
			if err != nil {
				t.Error(err)
			}

			for _, entry := range *oldest.Entries {
				name, inter, err := entry.GetInterface()
				if err != nil {
					t.Error(err)
				}
				t.Logf("Name: %s, Interface: %v, Error: %v", name, inter, err)

				if name == "message" {
					t.Logf("Messages: %v", *inter.(*datatype.Message).Message)
				}
			}
		})

	})

}
