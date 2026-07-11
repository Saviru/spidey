# Components and Styling

Components are `.spidey` files placed inside the `components/` directory. They allow you to create reusable UI pieces.

## Using Components

To use a component (e.g., `components/Button.spidey`), include it in your pages or other components using XML-like syntax:

```html
<Button title="Click Me" />
```

Spidey automatically resolves this to the `components/Button.spidey` file and renders it server-side.

## Partial Hydration (Islands Architecture)

By default, all components are rendered on the server to purely static HTML. If you need a component to be interactive on the client side (e.g., a React/Vanilla JS widget), you can use the `client:load` directive:

```html
<InteractiveWidget client:load />
```

**How it works:**
1. Spidey will wrap this component in a `<spidey-island data-component="InteractiveWidget">` custom element.
2. In your `components/` folder, you must create an `InteractiveWidget.js` file.
3. This `.js` file is bundled by ESBuild and shipped to the client.
4. Spidey's client engine dynamically imports the script and calls its exported `mount` function.

**Example `components/InteractiveWidget.js`**:
```javascript
export function mount(islandElement) {
    islandElement.innerHTML = "<button>I am hydrated!</button>";
    islandElement.querySelector("button").addEventListener("click", () => alert("Works!"));
}
```

*Note: Spidey natively uses vanilla JS for islands. Do not use `.jsx` extensions; standard `.js` is required.*

## Scoped CSS

You can add `<style>` blocks directly inside your `.spidey` components. By default, Spidey's compiler parses the CSS and automatically scopes it to the component by injecting unique `data-spidey-*` attributes to your HTML elements.

```html
<style>
    /* This will only affect this component */
    div { color: red; } 
    
    /* Use :global() to style elements outside the scope (like body or html) */
    :global(body) { background: black; } 
</style>
<div>Scoped text</div>
```

## CSS Modules

For stricter scoping, you can use `<style module>`. This hashes the class names directly, ensuring zero collisions. You access the hashed classes via the `$style` variable.

```html
<style module>
    .title { font-weight: bold; }
</style>
<h1 class="$style.title">Hello</h1>
```
