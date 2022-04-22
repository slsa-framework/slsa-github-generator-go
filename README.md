# Generation of SLSA3+ provenance for Go binaries
This repository contains a reference implementation for generating non-forgeable [SLSA provenance](https://slsa.dev/) that meets the requirement for the [SLSA level 3 and above](https://slsa.dev/spec/v0.1/levels) for projects using the Go programming language.

This repository contains the code, examples and technical design for our blog post on [Non forgeable SLSA provenance using GitHub workflows](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

***Note: This is a beta release and we are looking for your feedback. The official 1.0 release should come out in the next few weeks*** 

________
[Generation of provenance](#generation)
- [Example provenance](#example-provenance)
- [Configuration file](#configuration-file)
- [Workflow inputs](#workflow-inputs)
- [Workflow Example](#workflow-example)

[Verification of provenance](#verification-of-provenance)
- [Inputs](#inputs)
- [Command line examples](#command-line-examples)

[Technical design](#technial-design)
- [Blog posts](#blog-posts)
- [Specifications](#specifications)
________

## Generation
To generate provenance for a golang binary, follow the steps below:

### Example provenance
An example of the provenance generated from this repo is below:
```json
{
  "_type": "https://in-toto.io/Statement/v0.1",
  "predicateType": "https://slsa.dev/provenance/v0.2",
  "subject": [
    {
      "name": "binary-linux-amd64",
      "digest": {
        "sha256": "0ae7e4fa71686538440012ee36a2634dbaa19df2dd16a466f52411fb348bbc4e"
      }
    }
  ],
  "predicate": {
    "builder": {
      "id": "https://github.com/slsa-framework/slsa-github-generator-go/.github/workflows/slsa3_builder.yml@main"
    },
    "buildType": "https://github.com/slsa-framework/slsa-github-generator-go@v1",
    "invocation": {
      "configSource": {
        "uri": "git+https://github.com/asraa/slsa-on-github-test@refs/heads/main.git",
        "digest": {
          "sha1": "11dba28bf106e98f9992daa56e3967be41a5f11d"
        },
        "entryPoint": "Test SLSA"
      },
      "parameters": {
        "version": 1,
        "event_name": "workflow_dispatch",
        "ref_type": "branch",
        "ref": "refs/heads/main",
        "base_ref": "",
        "head_ref": "",
        "actor": "asraa",
        "sha1": "11dba28bf106e98f9992daa56e3967be41a5f11d",
        "event_payload": ...
      },
      "environment": {
        "arch": "amd64",
        "github_event_name": "workflow_dispatch",
        "github_run_attempt": "1",
        "github_run_id": "1995071837",
        "github_run_number": "95",
        "os": "ubuntu"
      }
    },
    "buildConfig": {
      "version": 1,
      "steps": [
        {
          "command": [
            "/opt/hostedtoolcache/go/1.17.7/x64/bin/go",
            "build",
            "-mod=vendor",
            "-trimpath",
            "-tags=netgo",
            "-ldflags=-X main.gitVersion=v1.2.3 -X main.gitSomething=somthg",
            "-o",
            "binary-linux-amd64"
          ],
          "env": [
            "GOOS=linux",
            "GOARCH=amd64",
            "GO111MODULE=on",
            "CGO_ENABLED=0"
          ]
        }
      ]
    },
    "materials": [
      {
        "uri": "git+asraa/slsa-on-github-test.git",
        "digest": {
          "sha1": "11dba28bf106e98f9992daa56e3967be41a5f11d"
        }
      }
    ]
  }
}
```

### Configuration file

Define a configuration file called `.slsa-goreleaser.yml` in the root of your project:

```yml
version: 1
# List of env variables used during compilation.
env:
  - GO111MODULE=on
  - CGO_ENABLED=0

# Flags for the compiler.
flags:
  - -trimpath
  - -tags=netgo

goos: linux     # same values as GOOS env variable. 
goarch: amd64   # same values as GOARCH env variable. 

# Binary name.
# {{ .OS }} will be replaced by goos field in the config file.
# {{ .Arch }} will be replaced by goarch field in the config file.
binary: binary-{{ .OS }}-{{ .Arch }}

# (Optional) ldflags generated dynamically in the workflow, and set as the `env` input variables in the workflow. 
ldflags:
  - '{{ .Env.VERSION_LDFLAGS }}'
```

### Workflow inputs

The builder workflow [slsa-framework/slsa-github-generator-go/.github/workflows/slsa3_builder.yml](.github/workflows/slsa3_builder.yml) accepts the following inputs:

| Name | Required | Description |
| ------------ | -------- | ----------- |
| `go-version` | no | The go version for your project. This value is passed, unchanged, to the [actions/setup-go](https://github.com/actions/setup-go) action when setting up the environment |
| `env` | no | A list of environment variables, seperated by `,`: `VAR1: value, VAR2: value`. This is typically used to pass dynamically-generated values, such as `ldflags`. Note that only environment variables with names starting with `CGO_` or `GO` are accepted.|

### Workflow Example
Create a new workflow, say `.github/workflows/slsa-goreleaser.yml`:

```yaml
name: SLSA go releaser
on:
  workflow_dispatch:
  push:
    tags:
      - "*" 

permissions: read-all
      
jobs:
  # Generate ldflags dynamically.
  # Optional: only needed for ldflags.
  args:
    runs-on: ubuntu-latest
    outputs:
      ldflags: ${{ steps.ldflags.outputs.value }}
    steps:
      - id: checkout
        uses: actions/checkout@ec3a7ce113134d7a93b817d10a8272cb61118579 # v2.3.4
        with:
          fetch-depth: 0
      - id: ldflags
        run: |
          echo "::set-output name=value::$(./scripts/version-ldflags)"

  # Trusted builder.
  build:
    permissions:
      id-token: write
      contents: read
    needs: args
    uses: slsa-framework/slsa-github-generator-go/.github/workflows/slsa3_builder.yml@main # TODO: use hash upon release.
    with:
      go-version: 1.17
      # Optional: only needed if using ldflags.
      env: "VERSION_LDFLAGS:${{needs.args.outputs.ldflags}}"

  # Upload to GitHub release.
  upload:
    permissions:
      contents: write
    runs-on: ubuntu-latest
    needs: build
    steps:
      - uses: actions/download-artifact@fb598a63ae348fa914e94cd0ff38f362e927b741
        with:
          name: ${{ needs.build.outputs.go-binary-name }}
      - uses: actions/download-artifact@fb598a63ae348fa914e94cd0ff38f362e927b741
        with:
          name: ${{ needs.build.outputs.go-binary-name }}.intoto.jsonl
      - name: Release
        uses: softprops/action-gh-release@1e07f4398721186383de40550babbdf2b84acfc5
        if: startsWith(github.ref, 'refs/tags/')
        with:
          files: |
            ${{ needs.build.outputs.go-binary-name }}
            ${{ needs.build.outputs.go-binary-name }}.intoto.jsonl
```

## Verification of provenance
To verify the provenance, use the [github.com/slsa-framework/slsa-verifier](https://github.com/slsa-framework/slsa-verifier) project. 

### Inputs
```shell
$ git clone git@github.com:slsa-framework/slsa-verifier.git
$ go run . --help
    -binary string
    	path to a binary to verify
    -branch string
    	expected branch the binary was compiled from (default "main")
    -provenance string
    	path to a provenance file
    -source string
    	expected source repository that should have produced the binary, e.g. github.com/some/repo
    -tag string
    	[optional] expected tag the binary was compiled from
    -versioned-tag string
    	[optional] expected version the binary was compiled from. Uses semantic version to match the tag
```

### Command line examples
```shell
$ go run . --binary ~/Downloads/binary-linux-amd64 --provenance ~/Downloads/binary-linux-amd64.intoto.jsonl --source github.com/origin/repo

Verified against tlog entry 1544571
verified SLSA provenance produced at 
 {
        "caller": "origin/repo",
        "commit": "0dfcd24824432c4ce587f79c918eef8fc2c44d7b",
        "job_workflow_ref": "/slsa-framework/slsa-github-generator-go/.github/workflows/slsa3_builder.yml@refs/heads/main",
        "trigger": "workflow_dispatch",
        "issuer": "https://token.actions.githubusercontent.com"
}
successfully verified SLSA provenance
```

## Technical design

### Blog post
Find our blog post series [here](https://security.googleblog.com/2022/04/improving-software-supply-chain.html).

### Specifications
For a more in-depth technical dive, read the [SPECIFICATIONS.md](./SPECIFICATIONS.md).
