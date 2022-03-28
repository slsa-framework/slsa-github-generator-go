// Copyright The GOSST team.
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
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/gossts/slsa-go/builder/pkg"
)

func usage(p string) {
	panic(fmt.Sprintf(`Usage: 
	 %s build [--dry] slsa-releaser.yml
	 %s provenance --binary-name $NAME --digest $DIGEST --command $COMMAND --env $ENV`, p, p))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	// Build command.
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	buildDry := buildCmd.Bool("dry", false, "dry run of the build without invoking compiler")

	// Provenance command.
	provenanceCmd := flag.NewFlagSet("provenance", flag.ExitOnError)
	provenanceName := provenanceCmd.String("binary-name", "", "untrusted binary name of the artifact built")
	provenanceDigest := provenanceCmd.String("digest", "", "sha256 digest of the untrusted binary")
	provenanceCommand := provenanceCmd.String("command", "", "command used to compile the binary")
	provenanceEnv := provenanceCmd.String("env", "", "env variables used to compile the binary")

	// Expect a sub-command.
	if len(os.Args) < 2 {
		usage(os.Args[0])
	}

	switch os.Args[1] {
	case buildCmd.Name():
		buildCmd.Parse(os.Args[2:])
		if len(buildCmd.Args()) < 1 {
			usage(os.Args[0])
		}

		goc, err := exec.LookPath("go")
		check(err)

		cfg, err := pkg.ConfigFromFile(buildCmd.Args()[0])
		check(err)
		fmt.Println(cfg)

		gobuild := pkg.GoBuildNew(goc, cfg)

		// Set env variables encoded as arguments.
		err = gobuild.SetArgEnvVariables(buildCmd.Args()[1])
		check(err)

		err = gobuild.Run(*buildDry)
		check(err)
	case provenanceCmd.Name():
		provenanceCmd.Parse(os.Args[2:])
		// Note: *provenanceEnv may be empty.
		if *provenanceName == "" || *provenanceDigest == "" ||
			*provenanceCommand == "" {
			usage(os.Args[0])
		}

		githubContext, ok := os.LookupEnv("GITHUB_CONTEXT")
		if !ok {
			panic(errors.New("environment variable GITHUB_CONTEXT not present"))
		}

		attBytes, err := pkg.GenerateProvenance(*provenanceName, *provenanceDigest,
			githubContext, *provenanceCommand, *provenanceEnv)
		check(err)

		filename := fmt.Sprintf("%s.intoto.jsonl", *provenanceName)
		err = ioutil.WriteFile(filename, attBytes, 0600)
		check(err)

		fmt.Printf("::set-output name=signed-provenance-name::%s\n", filename)
	default:
		fmt.Println("expected 'build' or 'provenance' subcommands")
		os.Exit(1)
	}
}
