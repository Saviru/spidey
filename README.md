# Spidey - Let's weave the web

Spidey is a blazing-fast, modern full-stack meta-framework written in Go. It enables seamless frontend and backend development with hybrid routing, server-side reactivity (S-Tags), scoped CSS, Islands architecture (partial hydration), API generation, and Static Site Generation (SSG).

## Features

*   **Hybrid Routing:** Drop a `.spidey` file in `pages/` for an instant route, or use `pages/folder/index.spidey` for complex nested routes and layouts.
*   **Colocated Components:** Prefix files or folders with an underscore (`_`) in your `pages/` directory to safely colocate private components right next to your routes.
*   **Islands Architecture:** Render everything as fast, static HTML by default. Opt-in to client-side JS interactivity only where needed using the `<Widget client:load />` directive.
*   **Zero-JS Server Reactivity (S-Tags):** Build highly interactive forms, buttons, and UI updates without writing JavaScript using built-in `s-post`, `s-get`, and `s-target` attributes (inspired by HTMX).
*   **AOT JavaScript Compilation:** Write inline JS (e.g., `@click="alert('hi')"`) directly in your HTML; Spidey extracts, hashes, and securely bundles it Ahead-Of-Time.
*   **Native Go Frontmatter:** Write pure Go backend logic, data structs, and database queries right at the top of your page files inside `---go ... ---` blocks.
*   **Scoped CSS & Modules:** Keep your styles fully encapsulated. Use `<style>` for component-scoped CSS or `<style module>` for strictly hashed, collision-free CSS classes.
*   **Deeply Nested Layouts:** Drop a `layout.spidey` file anywhere in your routing tree to instantly wrap all child routes with persistent UI (like navbars and sidebars).
*   **API Routes with Magic Comments:** Generate backend endpoints effortlessly. Use `//spidey:route GET /users` and `//spidey:middleware AuthCheck` to auto-register standard Go handlers without centralized routing files.
*   **Ergonomic Context Engine:** A powerful `router.Context` (inspired by Gin/Fiber) provides robust helpers for request parsing, automatic JSON struct binding and validation, and multiple JSON response types (SecureJSON, PureJSON, JSONP).
*   **Middleware & Standard Go Interop:** Seamlessly stack route-specific or group middlewares. Spidey natively wraps and supports standard Go middlewares (`func(http.Handler) http.Handler`) out of the box.
*   **Built-in Reverse Proxy:** Easily route and forward traffic to other backend microservices using the native `.Proxy("/path", "url")` method.
*   **Static Site Generation (SSG) & SSR:** Run as a blazing fast Server-Side Rendered Go binary, or export your entire app to static files via the CLI for CDN deployment.

## Installation

```bash
go install github.com/saviru/spidey/cmd/spidey@latest
```

## Quick Start

```bash
spidey init my-app
spidey dev
```

## Documentation

Full documentation is available in the docs/ directory:

- [Getting Started](./docs/getting-started.md)
- [CLI Reference](./docs/cli.md)
- [Pages and Layouts](./docs/pages.md)
- [Components and Styling](./docs/components.md)
- [API Routes](./docs/api-routes.md)
- [Router API](./docs/router.md)
- [Architecture](./docs/architecture.md)

## License

MIT License.

```Copyright (c) 2025 Saviru Kashmira Atapattu```

<br>
<hr>
<p align="center">Made with ❤️ for the web development community</p>