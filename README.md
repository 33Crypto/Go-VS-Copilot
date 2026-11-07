# Go-VS-Copilot

The Windows Azure Table Storage repository-pattern sample that used to live in this tree has been removed. That sample is still in [33Crypto/WindowsAzureRepository](https://github.com/33Crypto/WindowsAzureRepository). This repository was not used to delete or change that project.

What remains here is a handful of short text files, one shell script, one workflow file outside the Actions directory, and an empty Word document. There is no application to build or run.

## Files

- `.github/cosx-erb.yml` — GitHub Actions starter workflow named CI. It checks out the repository and prints "Hello, world!". It sits in `.github/` rather than `.github/workflows/`, so Actions does not run it. The trigger branches are `master`.
- `scripts/all-cmd.sh` — A one-line script, `echo hello`. It has no shebang, is not marked executable, and is not called by the workflow file.
- `ci.go` — Not a Go source file. It is five lines of notes (`#Create`, `#Add_Update(gh-os)`, `#Review_Commit`, `Commit.`). Nothing in this tree executes them. `gh-os` is not a file here.
- `go.os` — The same kind of note, naming `Index.os.tar.zst`. That path is not in this tree. The file that does exist is `Index.cs.tar.zst`.
- `Index.cs.tar.zst` — Not a zstandard archive and not a tar. The entire file is the line `#Create Index.cs.tar.zst`.
- `cosx-env.yml` — GitHub's FUNDING.yml template. Every sponsorship platform is left as a comment. It is not an environment configuration, and it is not at `.github/FUNDING.yml`.
- `cosx-vscode-codeql-tanx-triage.pro*` — Stock CodeQL workflow whose language matrix is `csharp`. The asterisk is part of the filename. The file is in the repository root, so Actions does not run it. This tree no longer contains a C# project for that workflow to build. The actions it references (`actions/checkout@v2` and `github/codeql-action` v1) are the original starter pins.
- `LOCK33SecDevOpsPyQSL4A.33WebDocxJSEditPro°.docx` — A Word 2010 document with no body text (a `_GoBack` bookmark only). The file on disk is that same document package written twice, back to back. It has no macros, embedded objects, or hyperlinks. Document properties name the creator as `huagen_DEV` (created 2019-04-04, modified 2020-08-03).

## Other branches

This README describes the files on `main`. The other branches on this repository (`patch-1`, `patch-2`, `patch-3`, and `33Crypto-go-cosx-pro(#-0001)`) still contain the Azure sample, along with a few extra stub files that are not on `main`.
