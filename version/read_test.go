// Copyright 2020 The Go Authors. All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package version

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type goVersion struct {
	name          string
	versionPrefix string
	buildImage    string
	boringcrypto  bool
}

type buildType struct {
	name  string
	flags []string
}

var goVersions = []goVersion{
	{name: "go1.9.7b4", versionPrefix: "go1.9.7b4", buildImage: "goboring/golang:1.9.7b4", boringcrypto: true},
	{name: "go1.10.8b4", versionPrefix: "go1.10.8b4", buildImage: "goboring/golang:1.10.8b4", boringcrypto: true},
	{name: "go1.11.13b4", versionPrefix: "go1.11.13b4", buildImage: "goboring/golang:1.11.13b4", boringcrypto: true},
	{name: "go1.12.17b4", versionPrefix: "go1.12.17b4", buildImage: "goboring/golang:1.12.17b4", boringcrypto: true},
	{name: "go1.13.14b4", versionPrefix: "go1.13.14b4", buildImage: "goboring/golang:1.13.14b4", boringcrypto: true},
	{name: "go1.14.6b4", versionPrefix: "go1.14.6b4", buildImage: "goboring/golang:1.14.6b4", boringcrypto: true},
	{name: "go1.15b5", versionPrefix: "go1.15b5", buildImage: "goboring/golang:1.15b5", boringcrypto: true},

	{name: "go1.9", versionPrefix: "go1.9", buildImage: "golang:1.9"},
	{name: "go1.10", versionPrefix: "go1.10", buildImage: "golang:1.10"},
	{name: "go1.11", versionPrefix: "go1.11", buildImage: "golang:1.11"},
	{name: "go1.12", versionPrefix: "go1.12", buildImage: "golang:1.12"},
	{name: "go1.13", versionPrefix: "go1.13", buildImage: "golang:1.13"},
	{name: "go1.14", versionPrefix: "go1.14", buildImage: "golang:1.14"},
	{name: "go1.15", versionPrefix: "go1.15", buildImage: "golang:1.15"},
}

var buildTypes = []buildType{
	{name: "default", flags: []string{}},
	{name: "stripped", flags: []string{"-ldflags=-s -w -buildid="}},
}

var mainCrypto = []byte(`
package main
import _ "crypto/rand"
func main(){}
`)

func TestReadExe(t *testing.T) {
	for _, v := range goVersions {
		for _, b := range buildTypes {
			testName := fmt.Sprintf("%s_%s", v.name, b.name)
			t.Run(testName, func(t *testing.T) {
				testFile := filepath.Join("testdata", testName)

				// build if needed, with specified go version and build options
				if _, err := os.Stat(testFile); os.IsNotExist(err) {
					if testing.Short() {
						t.Skipf("skipping: %q missing, build required", testFile)
					}
					t.Logf("%q missing, building (set POPULATE_TESTDATA=true to save to testdata)...", testFile)

					tmpDir, err := ioutil.TempDir("", testName)
					if err != nil {
						t.Fatal(err)
					}
					defer os.RemoveAll(tmpDir)

					if err := ioutil.WriteFile(filepath.Join(tmpDir, "main.go"), mainCrypto, os.FileMode(0755)); err != nil {
						t.Fatal(err)
					}
					cmd := []string{"docker", "run", "-v", fmt.Sprintf("%s:/workspace", tmpDir), v.buildImage}
					cmd = append(cmd, "go", "build", "-o", "/workspace/main")
					cmd = append(cmd, b.flags...)
					cmd = append(cmd, "/workspace/main.go")
					output, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
					if len(output) > 0 {
						t.Log(string(output))
					}
					if err != nil {
						t.Fatal(err)
					}

					builtFile := filepath.Join(tmpDir, "main")
					if os.Getenv("POPULATE_TESTDATA") == "true" {
						if err := os.MkdirAll("testdata", os.FileMode(0755)); err != nil {
							t.Fatal(err)
						}
						if err := os.Rename(builtFile, testFile); err != nil {
							t.Fatal(err)
						}
					} else {
						testFile = builtFile
					}
				}

				readVersion, err := ReadExe(testFile)
				if err != nil {
					t.Fatalf("error reading info: %v", err)
				}
				if e, a := v.versionPrefix, readVersion.Release; !strings.HasPrefix(a, e) {
					t.Errorf("expected version prefix %q, got %q", e, a)
				}
				if e, a := v.boringcrypto, readVersion.BoringCrypto; e != a {
					t.Errorf("expected boringcrypto %v, got %v", e, a)
				}
			})
		}
	}
}
