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

package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// defaultGoImageTag is the golang base image tag used when the Go version
// cannot be determined from the source tree nor from the running binary.
//
// It is deliberately "latest" rather than a pinned version. This is the branch
// taken when nothing at all is known about the required toolchain, and a
// version hardcoded here would silently go stale in exactly the way that
// motivated deriving the tag in the first place. "latest" always exists and is
// never older than the newest released Go, so it cannot be too old to build
// the agent.
const defaultGoImageTag = "latest"

// BuilderImageTag returns the tag of the golang base image to use for a remote
// (in-container) build of the user's agent, along with a short human-readable
// description of where that tag came from.
//
// A build container whose Go toolchain is older than the "go" directive of the
// agent being built fails outright: the official golang images pin
// GOTOOLCHAIN=local, so the go command is not allowed to upgrade itself. The
// tag is therefore derived from the agent's own go.mod, in this order:
//
//  1. override, if non-empty (the --go_version flag), used verbatim
//  2. the "toolchain" directive of moduleDir/go.mod
//  3. the "go" directive of moduleDir/go.mod
//  4. the Go version this binary was built with
//  5. defaultGoImageTag
//
// moduleDir is the directory that becomes the root of the uploaded archive.
// Only that exact directory is consulted, deliberately: it holds the only
// go.mod the remote build can see, so honouring an enclosing module further up
// the filesystem would derive a version from a module that is not the one being
// deployed. ModuleDirHasGoMod reports whether that file is actually there.
//
// Nothing here needs updating when a new version of ADK raises its own Go
// requirement: the agent's go.mod transitively requires ADK, so "go mod tidy"
// raises the agent's own "go" directive to match, and steps 2 and 3 pick that
// up on the next deployment.
//
// Versions from go.mod and from the running binary are reduced to their
// "<major>.<minor>" form, which is a rolling tag that always exists on Docker
// Hub and always resolves to the newest patch release of that minor version.
// Callers are expected to pair the result with ENV GOTOOLCHAIN=auto so that a
// go.mod requiring a patch release newer than the image ships still builds.
//
// BuilderImageTag never fails: an unreadable or malformed go.mod degrades to
// the next source rather than blocking a deployment.
func BuilderImageTag(override, moduleDir string) (tag, origin string) {
	if t := strings.TrimSpace(override); t != "" {
		return t, "--go_version flag"
	}

	goMod := filepath.Join(moduleDir, "go.mod")
	if directive, version := goModVersion(goMod); version != "" {
		return version, fmt.Sprintf("%s directive in %s", directive, goMod)
	}

	if v := majorMinor(runtime.Version()); v != "" {
		return v, "Go toolchain of the adkgo binary"
	}

	return defaultGoImageTag, "built-in default"
}

// ModuleDirHasGoMod reports whether dir directly contains a go.mod file.
//
// The remote build runs "go build" at the root of the uploaded archive, so a
// missing go.mod there means the build cannot succeed no matter which builder
// image is used. Callers use this to warn before uploading rather than letting
// the managed build fail with an opaque error.
func ModuleDirHasGoMod(dir string) bool {
	fi, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil && !fi.IsDir()
}

// goModVersion reports the Go version requested by the go.mod at path, reduced
// to "<major>.<minor>", and the name of the directive it came from.
//
// A "toolchain" directive wins over a "go" directive, mirroring the precedence
// the go command applies when choosing which toolchain to run. Both return ""
// if the file cannot be read or declares nothing usable.
func goModVersion(path string) (directive, version string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", ""
	}

	var goDirective, toolchainDirective string
	for line := range strings.SplitSeq(string(data), "\n") {
		if i := strings.Index(line, "//"); i >= 0 {
			line = line[:i]
		}
		// Only top-level directives matter. Lines inside a require/replace
		// block never have "go" or "toolchain" as their first field, since a
		// module path cannot be a bare keyword.
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "go":
			if goDirective == "" {
				goDirective = fields[1]
			}
		case "toolchain":
			if toolchainDirective == "" {
				toolchainDirective = fields[1]
			}
		}
	}

	// "toolchain default" carries no version, so fall through to "go" for it.
	if v := majorMinor(toolchainDirective); v != "" {
		return "toolchain", v
	}
	if v := majorMinor(goDirective); v != "" {
		return "go", v
	}
	return "", ""
}

// majorMinor reduces a Go version to the "<major>.<minor>" form used by the
// rolling golang image tags: "1.26.5", "go1.26.5", "1.26" and "1.26rc1" all
// become "1.26". It returns "" for anything it does not recognize, including
// the "devel ..." version reported by unreleased toolchains.
func majorMinor(v string) string {
	v = strings.TrimPrefix(strings.TrimSpace(v), "go")
	major, rest, ok := strings.Cut(v, ".")
	if !ok || major == "" || leadingDigits(major) != major {
		return ""
	}
	// Drop the patch component, then any pre-release suffix ("26rc1" -> "26").
	minor, _, _ := strings.Cut(rest, ".")
	minor = leadingDigits(minor)
	if minor == "" {
		return ""
	}
	return major + "." + minor
}

// leadingDigits returns the longest prefix of s consisting of ASCII digits.
func leadingDigits(s string) string {
	for i := range len(s) {
		if s[i] < '0' || s[i] > '9' {
			return s[:i]
		}
	}
	return s
}
