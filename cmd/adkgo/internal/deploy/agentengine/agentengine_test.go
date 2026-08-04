// Copyright 2026 Google LLC
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

package agentengine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeDockerfile renders the Dockerfile for the given builder image tag and
// returns its contents.
func writeDockerfile(t *testing.T, goImageTag string) string {
	t.Helper()

	// prepareDockerfile reads the server port from the package-level flags var.
	original := flags
	t.Cleanup(func() { flags = original })
	flags.agentEngine.serverPort = 8080

	f := &deployAgentEngineFlags{}
	f.build.dockerfileBuildPath = filepath.Join(t.TempDir(), "Dockerfile")
	f.build.execFile = "agent"
	f.build.goImageTag = goImageTag
	f.source.origEntryPointPath = "./main.go"

	if err := f.prepareDockerfile(); err != nil {
		t.Fatalf("prepareDockerfile() failed: %v", err)
	}
	got, err := os.ReadFile(f.build.dockerfileBuildPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) failed: %v", f.build.dockerfileBuildPath, err)
	}
	return string(got)
}

func TestPrepareDockerfileUsesResolvedGoImageTag(t *testing.T) {
	for _, tc := range []struct {
		name string
		tag  string
		want string
	}{
		{name: "major minor", tag: "1.26", want: "FROM golang:1.26 AS builder"},
		{name: "explicit patch", tag: "1.26.5", want: "FROM golang:1.26.5 AS builder"},
		{name: "variant tag", tag: "1.27-alpine", want: "FROM golang:1.27-alpine AS builder"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := writeDockerfile(t, tc.tag); !strings.Contains(got, tc.want) {
				t.Errorf("Dockerfile does not contain %q:\n%s", tc.want, got)
			}
		})
	}
}

func TestPrepareDockerfileAllowsToolchainUpgrade(t *testing.T) {
	// The official golang images pin GOTOOLCHAIN=local, which turns a go.mod
	// asking for a newer toolchain than the image ships into a build failure.
	// See https://github.com/google/adk-go/issues/1196.
	got := writeDockerfile(t, "1.26")
	if !strings.Contains(got, "ENV GOTOOLCHAIN=auto") {
		t.Errorf("Dockerfile does not set GOTOOLCHAIN=auto:\n%s", got)
	}
}

func TestPrepareDockerfileUsesUppercaseAs(t *testing.T) {
	// Lowercase "as" makes BuildKit emit a FromAsCasing warning.
	got := writeDockerfile(t, "1.26")
	if strings.Contains(got, " as builder") {
		t.Errorf("Dockerfile uses lowercase \"as builder\":\n%s", got)
	}
}

func TestPrepareDockerfileBuildsEntryPoint(t *testing.T) {
	got := writeDockerfile(t, "1.26")
	for _, want := range []string{
		"CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags \"-s -w\" -o agent ./main.go",
		"COPY --from=builder /app/agent",
		"EXPOSE 8080",
		`CMD ["/app/agent", "web", "-port", "8080", "agentengine"]`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("Dockerfile does not contain %q:\n%s", want, got)
		}
	}
}
