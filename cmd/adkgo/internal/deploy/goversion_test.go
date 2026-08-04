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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMajorMinor(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{name: "patch release", in: "1.26.5", want: "1.26"},
		{name: "major minor only", in: "1.21", want: "1.21"},
		{name: "go prefix", in: "go1.26.5", want: "1.26"},
		{name: "release candidate", in: "1.26rc1", want: "1.26"},
		{name: "beta", in: "go1.27beta2", want: "1.27"},
		{name: "surrounding space", in: "  1.26.5\r", want: "1.26"},
		{name: "two digit minor", in: "1.9.7", want: "1.9"},
		{name: "devel toolchain", in: "devel go1.27-abc123", want: ""},
		{name: "toolchain default", in: "default", want: ""},
		{name: "empty", in: "", want: ""},
		{name: "no minor", in: "1.", want: ""},
		{name: "no separator", in: "126", want: ""},
		{name: "non numeric major", in: "x.26", want: ""},
		{name: "non numeric minor", in: "1.x", want: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := majorMinor(tc.in); got != tc.want {
				t.Errorf("majorMinor(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestGoModVersion(t *testing.T) {
	for _, tc := range []struct {
		name          string
		content       string
		wantDirective string
		wantVersion   string
	}{
		{
			name:          "go directive with patch",
			content:       "module example.com/agent\n\ngo 1.26.5\n",
			wantDirective: "go",
			wantVersion:   "1.26",
		},
		{
			name:          "go directive without patch",
			content:       "module example.com/agent\n\ngo 1.21\n",
			wantDirective: "go",
			wantVersion:   "1.21",
		},
		{
			name:          "toolchain wins over go",
			content:       "module example.com/agent\n\ngo 1.26.5\n\ntoolchain go1.27.0\n",
			wantDirective: "toolchain",
			wantVersion:   "1.27",
		},
		{
			name:          "toolchain default falls back to go",
			content:       "module example.com/agent\n\ngo 1.26.5\n\ntoolchain default\n",
			wantDirective: "go",
			wantVersion:   "1.26",
		},
		{
			name: "require block is not mistaken for a directive",
			content: "module example.com/agent\n\ngo 1.26.5\n\n" +
				"require (\n\tgolang.org/x/mod v0.37.0 // indirect\n\tgo.uber.org/zap v1.27.0\n)\n",
			wantDirective: "go",
			wantVersion:   "1.26",
		},
		{
			name:          "trailing comment is stripped",
			content:       "module example.com/agent\n\ngo 1.26.5 // pinned by the platform team\n",
			wantDirective: "go",
			wantVersion:   "1.26",
		},
		{
			name:          "commented out directive is ignored",
			content:       "module example.com/agent\n\n// go 1.26.5\n",
			wantDirective: "",
			wantVersion:   "",
		},
		{
			name:          "crlf line endings",
			content:       "module example.com/agent\r\n\r\ngo 1.26.5\r\n",
			wantDirective: "go",
			wantVersion:   "1.26",
		},
		{
			name:          "no go directive",
			content:       "module example.com/agent\n",
			wantDirective: "",
			wantVersion:   "",
		},
		{
			name:          "malformed",
			content:       "this is not a go.mod at all\ngo\n",
			wantDirective: "",
			wantVersion:   "",
		},
		{
			name:          "empty file",
			content:       "",
			wantDirective: "",
			wantVersion:   "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "go.mod")
			if err := os.WriteFile(path, []byte(tc.content), 0o600); err != nil {
				t.Fatalf("WriteFile(%q) failed: %v", path, err)
			}
			directive, version := goModVersion(path)
			if directive != tc.wantDirective || version != tc.wantVersion {
				t.Errorf("goModVersion() = (%q, %q), want (%q, %q)", directive, version, tc.wantDirective, tc.wantVersion)
			}
		})
	}
}

func TestGoModVersionMissingFile(t *testing.T) {
	directive, version := goModVersion(filepath.Join(t.TempDir(), "absent", "go.mod"))
	if directive != "" || version != "" {
		t.Errorf("goModVersion() on a missing file = (%q, %q), want two empty strings", directive, version)
	}
}

func TestModuleDirHasGoMod(t *testing.T) {
	withGoMod := t.TempDir()
	if err := os.WriteFile(filepath.Join(withGoMod, "go.mod"), []byte("module example.com/agent\n"), 0o600); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
	goModIsDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(goModIsDir, "go.mod"), 0o750); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	for _, tc := range []struct {
		name string
		dir  string
		want bool
	}{
		{name: "go.mod present", dir: withGoMod, want: true},
		{name: "no go.mod", dir: t.TempDir(), want: false},
		{name: "go.mod is a directory", dir: goModIsDir, want: false},
		{name: "dir does not exist", dir: filepath.Join(t.TempDir(), "absent"), want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ModuleDirHasGoMod(tc.dir); got != tc.want {
				t.Errorf("ModuleDirHasGoMod(%q) = %v, want %v", tc.dir, got, tc.want)
			}
		})
	}
}

func TestBuilderImageTag(t *testing.T) {
	sourceDir := t.TempDir()
	goMod := filepath.Join(sourceDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/agent\n\ngo 1.26.5\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) failed: %v", goMod, err)
	}

	t.Run("go.mod drives the tag", func(t *testing.T) {
		tag, origin := BuilderImageTag("", sourceDir)
		if tag != "1.26" {
			t.Errorf("BuilderImageTag() tag = %q, want %q", tag, "1.26")
		}
		if origin == "" {
			t.Error("BuilderImageTag() origin is empty, want a description of the source")
		}
	})

	t.Run("override wins over go.mod", func(t *testing.T) {
		tag, origin := BuilderImageTag("1.27-alpine", sourceDir)
		if tag != "1.27-alpine" {
			t.Errorf("BuilderImageTag() tag = %q, want %q", tag, "1.27-alpine")
		}
		if origin != "--go_version flag" {
			t.Errorf("BuilderImageTag() origin = %q, want %q", origin, "--go_version flag")
		}
	})

	t.Run("blank override is ignored", func(t *testing.T) {
		tag, _ := BuilderImageTag("   ", sourceDir)
		if tag != "1.26" {
			t.Errorf("BuilderImageTag() tag = %q, want %q", tag, "1.26")
		}
	})

	t.Run("ignores an enclosing module outside the archive root", func(t *testing.T) {
		// Only sourceDir's contents are uploaded, so a go.mod above it is not
		// visible to the remote build. Deriving a version from it would report
		// a toolchain for a module that is not the one being deployed.
		nested := filepath.Join(sourceDir, "cmd", "agent")
		if err := os.MkdirAll(nested, 0o750); err != nil {
			t.Fatalf("MkdirAll(%q) failed: %v", nested, err)
		}
		tag, origin := BuilderImageTag("", nested)
		if strings.Contains(origin, "go.mod") {
			t.Errorf("BuilderImageTag() used %s for a directory with no go.mod of its own; tag = %q", origin, tag)
		}
	})

	t.Run("falls back to a usable tag without a go.mod", func(t *testing.T) {
		// Nothing above a temp dir declares a module, so this exercises the
		// runtime/default fallbacks. Assert the shape rather than an exact
		// value, which depends on the toolchain running the test.
		tag, origin := BuilderImageTag("", t.TempDir())
		if majorMinor(tag) != tag && tag != defaultGoImageTag {
			t.Errorf("BuilderImageTag() tag = %q, want a <major>.<minor> version or %q", tag, defaultGoImageTag)
		}
		if origin == "" {
			t.Error("BuilderImageTag() origin is empty, want a description of the source")
		}
	})
}

func TestDefaultGoImageTagIsNotPinnedToAVersion(t *testing.T) {
	// The last-resort default is only reached when the required Go version is
	// entirely unknown. Pinning a version here would rot in exactly the way the
	// hardcoded golang:1.25 did, so it has to stay a rolling tag.
	// See https://github.com/google/adk-go/issues/1196.
	if v := majorMinor(defaultGoImageTag); v != "" {
		t.Errorf("defaultGoImageTag = %q, which pins Go %s; want a rolling tag that cannot go stale", defaultGoImageTag, v)
	}
}
