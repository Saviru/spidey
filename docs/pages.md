# Pages and Layouts

Spidey uses file-based routing. Any `.spidey` file inside the `pages/` directory automatically becomes a route mapped to its filename.

## Hybrid Routing (File & Folder Based)

Spidey supports both file-based and folder-based routing seamlessly:
*   **File-based:** `pages/about.spidey` becomes `/about`.
*   **Folder-based:** `pages/about/index.spidey` also becomes `/about`.

This gives you the flexibility to use a simple flat file structure for small pages, and transition to a folder-based structure when routes get more complex.

## Private Files & Colocation

When using folder-based routing, you might want to put components or helper files next to your page without them becoming public URLs. 

Spidey ignores any file or folder in the `pages/` directory that starts with an underscore (`_`).

```text
pages/
├── index.spidey              -> /
└── dashboard/                -> /dashboard
    ├── index.spidey          
    └── _components/          -> Ignored by the router!
        ├── Sidebar.spidey    
        └── Chart.spidey      
```
These ignored `.spidey` files are still compiled as components and can be used inside your pages (e.g., `<_components/Sidebar />` or `{{ template "_components/Sidebar" . }}`).

## Basic Page and Go Frontmatter

Spidey pages are not just static HTML; they are executed natively by Go. You can define Go data structures and backend logic at the top of your file using `---go ... ---` frontmatter.

```html
---go
type PageData struct {
    Message string
}
---
<div>
    <h1>Hello, {{.Message}}!</h1>
</div>
```

The code inside the frontmatter block is extracted and injected into the transpiled server file. This means you can define structs, import packages, or write helper functions specific to that route.

## Dynamic Routes

Create files with brackets, like `pages/users/[id].spidey`. Spidey's transpiler automatically extracts URL parameters and passes them into the page's data map.

Inside `pages/users/[id].spidey`, you can access the parameter by capitalizing the bracketed name:

```html
<h1>User Profile</h1>
<p>Viewing data for User ID: {{.Id}}</p>
```

If you have multiple parameters (e.g., `pages/users/[id]/posts/[postId].spidey`), they are accessible as `{{.Id}}` and `{{.PostId}}`.

## Layouts

Spidey supports deeply nested layouts. You can create a `layout.spidey` file inside any directory in `pages/`. 

The inner content of the current directory will be injected wherever `{{template "content" .}}` is defined in the layout.

The root layout of the entire application is `app.spidey` at the project root. This is where you should define your `<html>`, `<head>`, and `<body>` tags. Spidey automatically injects livereload scripts, the CSS bundle, and the client engine bootstrapper into `app.spidey`.

## Client-Side Interactivity (S-Tags)

Spidey provides built-in attributes (S-Tags) for server-side reactivity, heavily inspired by HTMX, allowing you to build dynamic UIs without writing JavaScript:

- `s-get="url"`: Fetches HTML from the URL on click (or specified trigger).
- `s-post="url"`: Submits a form via POST (prevents default reload).
- `s-target="selector"`: The CSS selector of the element to replace with the response HTML.
- `s-swap="style"`: How to swap the content (`innerHTML` by default, or `outerHTML`).
- `s-trigger="event"`: Overrides the default trigger (`click` or `submit`). Supports modifiers:
  - Custom events: `s-trigger="keyup"`, `s-trigger="change"`, etc.
  - Debouncing: `s-trigger="keyup delay:500ms"` (waits 500ms after the last keystroke).
  - Polling: `s-trigger="every:5s"` (fetches data every 5 seconds).
  - Lazy Loading: `s-trigger="intersect"` (fetches when the element scrolls into view).
- `@transition="name"` or `s-transition="name"`: Uses the native browser View Transitions API to seamlessly animate the DOM swap. 
  - **Built-in Animations**: Set the name to `spidey-fade`, `spidey-slide-up`, `spidey-slide-down`, or `spidey-scale` for ready-to-use smooth animations! If set to `true`, it defaults to `spidey-fade`.
  - **Custom Animations**: If given a custom name (e.g. `slide-fade`), you can write your own CSS animations using `::view-transition-old(slide-fade)` and `::view-transition-new(slide-fade)`. *Tip: To prevent text from "vibrating" during custom animations, add `mix-blend-mode: normal;` to your pseudo-elements!*
- `@prefetch="event"` or `s-prefetch="event"`: Preloads the HTML response before the user clicks. Only applies to `s-get` requests.
  - `@prefetch="hover"` (Default): Starts fetching when the user hovers over the element.
  - `@prefetch="intersect"`: Starts fetching as soon as the element scrolls into the viewport.


Example:
```html
<form s-post="/api/submit" s-target="#result">
    <input name="email" type="email" />
    <button type="submit">Send</button>
</form>
<div id="result"></div>
```

## AOT Events

You can write inline JavaScript events using the `@` syntax. Spidey's Ahead-Of-Time (AOT) compiler extracts these, assigns a unique ID, and bundles them into vanilla JavaScript listeners.

```html
<button @click="console.log('Clicked!');">Click Me</button>
```
During build, this is converted to something like `<button id="s-1a2b3c4d">` and the Javascript logic is securely shipped in `spidey-aot.js`.
