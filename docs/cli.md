# CLI Reference

The Spidey CLI offers standard commands and spider-themed aliases.

## init (Alias: hatch)
```bash
spidey init [project-name]
```
Initializes a new Go module, downloads dependencies, and generates the required workspace folders (`api`, `pages`, `components`, `public`) and boilerplate templates.

## dev (Alias: weave)
```bash
spidey dev
```
Starts the development environment. It launches a file watcher, transpiles pages on the fly, starts a live reload server, and automatically recompiles and restarts your Go backend when changes are detected.

## build (Alias: wrap)
```bash
spidey build
```
Transpiles all `.spidey` pages into Go code, bundles frontend assets using esbuild, and compiles the final production binary to the configured output directory (`bin/server` by default).

## export (Alias: shed)
```bash
spidey export
```
Transpiles pages and runs a temporary server to crawl and export all non-dynamic routes as static HTML files. The output is placed in the `out/` directory, making it ready for static hosting.

## version (Aliases: -v, --version)
```bash
spidey version
```
Outputs the currently installed version of the Spidey CLI. This version number is automatically embedded based on the Git Tag you downloaded via `go install`.

## update
```bash
spidey update
```
Automatically downloads and installs the latest version of the Spidey CLI directly from GitHub. Spidey also runs a fast, non-blocking check in the background when running other commands to notify you if a new version is available!
