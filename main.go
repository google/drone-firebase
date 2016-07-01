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
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Firebase defines the types of data necessary for deployment.
type Firebase struct {
	Token     string `json:"token"`
	ProjectID string `json:"project_id"` // optional.
	Message   string `json:"message"`    // optional.
	Targets   string `json:"targets"`    // optional.
	DryRun    bool   `json:"dryrun"`     // optional.
	Debug     bool   `json:"debug"`      // optional.
}

// Workspace defines the types of data required from the workspace.
type Workspace struct {
	Path string `json:"path"`
}

var (
	buildCommit string
)

func main() {
	fmt.Printf("Firebase Plugin for Drone built from %s\n", buildCommit)

	flag.Parse()
	if len(flag.Args()) < 1 {
		fmt.Printf("Firebase: Too few arguments.\n")
		os.Exit(1)
	}

	vargs := flag.Args()[0]

	w := new(Workspace)
	f := new(Firebase)
	if err := parseJSON(strings.NewReader(vargs), w, f); err != nil {
		fmt.Printf("Firebase: Unable to parse invalid plugin input: %s\n", err)
		os.Exit(1)
	}

	if f.Debug {
		fmt.Printf("Workspace data: %+v\n", w)
		fmt.Printf("Firebase plugin data: %+v\n", f)
	}

	if err := doDeployment(w, f); err != nil {
		fmt.Printf("Firebase: Error in deployment: %s\n", err)
		os.Exit(1)
	}
}

func unmarshalData(key string, dict map[string]json.RawMessage, output interface{}) error {
	data, ok := dict[key]
	if !ok {
		return fmt.Errorf("'%s' does not exist in JSON dictionary", key)
	}
	err := json.Unmarshal(data, output)
	if err != nil {
		return fmt.Errorf("Unable to unmarshal '%s': %s", key, err)
	}
	return nil
}

// parseVargs parses the vargs key from the given reader which should provide a JSON dictionary.
func parseJSON(reader io.Reader, workspace *Workspace, firebase *Firebase) error {
	raw := map[string]json.RawMessage{}
	err := json.NewDecoder(reader).Decode(&raw)
	if err != nil {
		return err
	}

	if err := unmarshalData("workspace", raw, workspace); err != nil {
		return err
	}
	if err := unmarshalData("vargs", raw, firebase); err != nil {
		return err
	}

	firebase.Token = strings.TrimSpace(firebase.Token)
	firebase.Message = strings.TrimSpace(firebase.Message)

	if len(firebase.Token) == 0 {
		return fmt.Errorf("Firebase token can not be nil")
	}

	return nil
}

func doDeployment(w *Workspace, f *Firebase) error {
	fmt.Printf("Changing to path: %s\n", w.Path)
	os.Chdir(w.Path)

	if f.shouldSetProject() {
		use := f.buildUse()
		if err := execute(use, f.Debug, f.DryRun); err != nil {
			return err
		}
	}

	deploy := f.buildDeploy()
	if err := execute(deploy, f.Debug, f.DryRun); err != nil {
		return err
	}

	return nil
}

func (f *Firebase) shouldSetProject() bool {
	return f.ProjectID != ""
}

func getEnvironment(oldEnv []string, f *Firebase) []string {
	var env []string
	for _, v := range oldEnv {
		if !strings.HasPrefix(v, "DEBUG=") && !strings.HasPrefix(v, "FIREBASE_TOKEN=") {
			env = append(env, v)
		}
	}
	env = append(env, fmt.Sprintf("FIREBASE_TOKEN=%s", f.Token))
	if f.Debug {
		env = append(env, fmt.Sprintf("DEBUG=%s", "true"))
	}
	return env
}

// buildUse creates a command on the form:
// $ firebase use ...
func (f *Firebase) buildUse() *exec.Cmd {
	var args []string
	args = append(args, "use")

	if f.ProjectID != "" {
		args = append(args, f.ProjectID)
	}

	cmd := exec.Command("firebase", args...)
	cmd.Env = getEnvironment(os.Environ(), f)
	return cmd
}

// buildDeploy creates a command on the form:
// $ firebase deploy \
//   [--only ...] \
//   [--message ...]
func (f *Firebase) buildDeploy() *exec.Cmd {
	var args []string
	args = append(args, "deploy")

	if f.Targets != "" {
		args = append(args, "--only")
		args = append(args, f.Targets)
	}

	if f.Message != "" {
		args = append(args, "--message")
		args = append(args, fmt.Sprintf("\"%s\"", f.Message))
	}

	cmd := exec.Command("firebase", args...)
	cmd.Env = getEnvironment(os.Environ(), f)
	return cmd
}

// execute sets the stdout and stderr of the command to be the default, traces
// the command to be executed and returns the result of the command execution.
func execute(cmd *exec.Cmd, debug, dryRun bool) error {
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if dryRun || debug {
		fmt.Println("$", strings.Join(cmd.Args, " "))
	}
	if dryRun {
		return nil
	}
	return cmd.Run()
}
