// Copyright 2016 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

var parsingTestData = []struct {
	description  string
	json         string
	expWorkspace Workspace
	expFirebase  Firebase
	expError     bool
}{
	{
		"All data has been set",
		`{
			"system": {
				"link": "http://drone.mycompany.com"
			},
			"repo": {
				"owner": "octocat",
				"name": "hello-world",
				"full_name": "octocat/hello-world",
				"link_url": "https://github.com/octocat/hello-world",
				"clone_url": "https://github.com/octocat/hello-world.git"
			},
			"build": {
				"number": 1,
				"event": "push",
				"branch": "master",
				"commit": "436b7a6e2abaddfd35740527353e78a227ddcb2c",
				"ref": "refs/heads/master",
				"author": "octocat",
				"author_email": "octocat@github.com"
			},
			"workspace": {
				"root": "/drone/src",
				"path": "/drone/src/github.com/octocat/hello-world",
				"keys": {
					"private": "-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQC..."
				}
			},
			"vargs": {
				"token": "1/deadbeefforrealz-4",
				"project_id": "staging",
				"message": "Autodeploy of commit $$COMMIT",
				"targets": "hosting",
				"dryrun": true,
				"debug": true
			}
		}`,
		Workspace{
			Path: "/drone/src/github.com/octocat/hello-world",
		},
		Firebase{
			Token:     "1/deadbeefforrealz-4",
			ProjectID: "staging",
			Message:   "Autodeploy of commit $$COMMIT",
			Targets:   "hosting",
			DryRun:    true,
			Debug:     true,
		},
		false,
	},
	{
		"Token and message has whitespace",
		`{
			"workspace": {},
			"vargs": {
				"token": "    foo    ",
				"message": "    bar    "
			}
		}`,
		Workspace{},
		Firebase{
			Token:   "foo",
			Message: "bar",
		},
		false,
	},
	{
		"No token specified",
		`{
				"workspace": {},
				"vargs": {}
			}`,
		Workspace{},
		Firebase{},
		true,
	},
}

func TestParseJSONAllData(t *testing.T) {
	for i, data := range parsingTestData {
		var (
			actualWorkspace Workspace
			actualFirebase  Firebase
		)
		err := parseJSON(strings.NewReader(data.json), &actualWorkspace, &actualFirebase)
		hasError := err != nil
		if data.expError != hasError {
			t.Fatalf("Case %d (%s): Expected %t, got %t", i, data.description, data.expError, hasError)
		}
		if !reflect.DeepEqual(data.expWorkspace, actualWorkspace) {
			t.Fatalf("Case %d (%s):: Expected %+v, got %+v", i, data.description, data.expWorkspace, actualWorkspace)
		}
		if !reflect.DeepEqual(data.expFirebase, actualFirebase) {
			t.Fatalf("Case %d (%s):: Expected %+v, got %+v", i, data.description, data.expFirebase, actualFirebase)
		}
	}
}

func contains(needle string, haystack []string) bool {
	for _, item := range haystack {
		if item == needle {
			return true
		}
	}
	return false
}

func verifyToken(f Firebase, env []string) bool {
	tokenString := fmt.Sprintf("FIREBASE_TOKEN=%s", f.Token)
	return contains(tokenString, env)
}

func verifyDebugEnv(f Firebase, env []string, t *testing.T) {
	debugString := "DEBUG=true"
	numDebugEnvs := 0
	for _, v := range env {
		if strings.HasPrefix(v, "DEBUG=") {
			numDebugEnvs++
		}
	}
	if f.Debug {
		if !(contains(debugString, env) && numDebugEnvs == 1) {
			t.Fatalf("Expected environment to contain DEBUG=true but it is %s", env)
		}
	} else {
		if numDebugEnvs != 0 {
			t.Fatalf("Expected environment to not contain any DEBUG= but it is %s", env)
		}
	}
}

func TestGetEnvironmentRemovesOldVariables(t *testing.T) {
	oldEnv := []string{"DEBUG=true", "FIREBASE_TOKEN=foo"}
	f := Firebase{Token: "bar", Debug: false}
	env := getEnvironment(oldEnv, &f)
	expectedLen := 1
	if len(env) != expectedLen {
		t.Fatalf("Wrong lenth for environment: Expected %d, got %d. Environment: %v", expectedLen, len(env), env)
	}
	verifyDebugEnv(f, env, t)
	if !verifyToken(f, env) {
		t.Fatalf("Expected environment to contain %s but it is %s", "FIREBASE_TOKEN=bar", env)
	}
}

var testdata = []struct {
	f            Firebase
	expShouldSet bool
	expDebug     string
	expToken     string
	expUse       []string
	expDeploy    []string
}{
	{
		Firebase{
			Token:     "",
			ProjectID: "",
			Message:   "",
			Targets:   "",
			Debug:     false,
		},
		false,
		"DEBUG=false",
		"FIREBASE_TOKEN=",
		[]string{
			"firebase",
			"use",
		},
		[]string{
			"firebase",
			"deploy",
		},
	},
	{
		Firebase{
			Token:     "",
			ProjectID: "my-project-id",
			Message:   "",
			Targets:   "",
			Debug:     false,
		},
		true,
		"DEBUG=false",
		"FIREBASE_TOKEN=",
		[]string{
			"firebase",
			"use",
			"my-project-id",
		},
		[]string{
			"firebase",
			"deploy",
		},
	},
	{
		Firebase{
			Token:     "1/2/3",
			ProjectID: "",
			Message:   "",
			Targets:   "",
			Debug:     false,
		},
		false,
		"DEBUG=false",
		"FIREBASE_TOKEN=1/2/3",
		[]string{
			"firebase",
			"use",
		},
		[]string{
			"firebase",
			"deploy",
		},
	},
	{
		Firebase{
			Token:     "",
			ProjectID: "",
			Message:   "my cool message",
			Targets:   "",
			Debug:     false,
		},
		false,
		"DEBUG=false",
		"FIREBASE_TOKEN=",
		[]string{
			"firebase",
			"use",
		},
		[]string{
			"firebase",
			"deploy",
			"--message",
			"\"my cool message\"",
		},
	},
	{
		Firebase{
			Token:     "",
			ProjectID: "",
			Message:   "",
			Targets:   "storage,hosting",
			Debug:     false,
		},
		false,
		"DEBUG=false",
		"FIREBASE_TOKEN=",
		[]string{
			"firebase",
			"use",
		},
		[]string{
			"firebase",
			"deploy",
			"--only",
			"storage,hosting",
		},
	},
	{
		Firebase{
			Token:     "1/2/3",
			ProjectID: "my-cool-project",
			Message:   "my cool message",
			Targets:   "storage,hosting",
			Debug:     true,
		},
		true,
		"DEBUG=true",
		"FIREBASE_TOKEN=1/2/3",
		[]string{
			"firebase",
			"use",
			"my-cool-project",
		},
		[]string{
			"firebase",
			"deploy",
			"--only",
			"storage,hosting",
			"--message",
			"\"my cool message\"",
		},
	},
}

func TestShouldSet(t *testing.T) {
	for i, data := range testdata {
		c := data.f.shouldSetProject()
		if c != data.expShouldSet {
			t.Fatalf("Case %d: Expected %t, got %t", i, data.expShouldSet, c)
		}
	}
}

func TestBuildUse(t *testing.T) {
	for i, data := range testdata {
		c := data.f.buildUse()
		if len(c.Args) != len(data.expUse) {
			t.Fatalf("Case %d: Expected %d, got %d", i, len(data.expUse), len(c.Args))
		}
		for j := range c.Args {
			if c.Args[j] != data.expUse[j] {
				t.Fatalf("Case %d:\nExpected:\n\t%s\nGot:\n\t%s", j, strings.Join(data.expUse, " "), strings.Join(c.Args, " "))
			}
		}
		verifyDebugEnv(data.f, c.Env, t)
		if !verifyToken(data.f, c.Env) {
			t.Fatalf("Case %d: Expected environment to contain %s but it is %s", i, data.expToken, c.Env)
		}
	}
}

func TestBuildDeploy(t *testing.T) {
	for i, data := range testdata {
		c := data.f.buildDeploy()
		if len(c.Args) != len(data.expDeploy) {
			t.Fatalf("Case %d: Expected %d, got %d", i, len(data.expDeploy), len(c.Args))
		}
		for j := range c.Args {
			if c.Args[j] != data.expDeploy[j] {
				t.Fatalf("Case %d:\nExpected:\n\t%s\nGot:\n\t%s", j, strings.Join(data.expDeploy, " "), strings.Join(c.Args, " "))
			}
		}
		verifyDebugEnv(data.f, c.Env, t)
		if !verifyToken(data.f, c.Env) {
			t.Fatalf("Case %d: Expected environment to contain %s but it is %s", i, data.expToken, c.Env)
		}
	}
}
