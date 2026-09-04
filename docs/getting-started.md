# Getting Started

## Installation

You need Go installed on your machine. Install the Spidey CLI globally:

```bash
go install github.com/yourusername/spidey/cmd/spidey@latest
```

## Creating a Project

Run the initialization command:

```bash
spidey init [project-name]
```

This will create a new directory (or initialize the current one) with the following structure:

- `api/`: Your backend Go API routes and main.go entry point.
- `pages/`: Your `.spidey` files for frontend pages and routing.
- `components/`: Your reusable `.spidey` components.
- `public/`: Static assets (images, fonts, etc.).
- `spidey.config.json`: Configuration file.
- `app.spidey`: The root layout for your application.

## Configuration

The generated `spidey.config.json` allows you to customize the framework's behavior:

```json
{
    "port": 3000,
    "directories": {
        "publicDir": "public",
        "outputDir": "bin/server"
    },
    "wsAllowedOrigins": []
}
```
- **`port`**: The port your server and development watcher will run on.
- **`publicDir`**: The folder Spidey uses to output generated CSS, JS islands, and serve static files.
- **`outputDir`**: The filepath where the final `go build` executable binary will be saved.
- **`wsAllowedOrigins`**: A list of allowed cross-origin domains for WebSockets (leave empty to enforce Same-Origin Policy).

## Development

Start the development server with live reload:

```bash
spidey dev
```

This starts the server on the configured port. Spidey watches your `pages/`, `components/`, and `app.spidey` files. When you save a file, it intelligently recompiles the Go templates, builds the binary, restarts the server, and uses a Server-Sent Event (SSE) to live-reload your browser!

## Building for Production

Compile your application into a single executable binary:

```bash
spidey build
```

The output will be placed in `bin/server` (or `bin/server.exe` on Windows). You can simply execute this binary to run your production server.

## Static Export (SSG)

Export your application as a purely static HTML/CSS/JS site:

```bash
spidey export
```

This command transpiles the pages, compiles a temporary server, and automatically crawls your non-dynamic routes. It generates an `out/` directory containing the purely static files, which can be deployed to any static host (like GitHub Pages or Vercel).
