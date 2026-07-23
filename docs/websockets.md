# WebSockets in Spidey

Spidey comes with a powerful, **native WebSocket engine** that allows you to build real-time, highly scalable applications (like chat apps, live dashboards, and multiplayer games) using the same declarative `s-tags` you already know!

## The Backend (Go)

Spidey's `core.Context` provides a built-in `c.Upgrade()` method to instantly upgrade an HTTP connection to a WebSocket.

Spidey also provides a powerful `c.App.WS` Hub that automatically handles Pub/Sub, Channels/Rooms, and even horizontal scaling via Redis!

### 1. Simple WebSocket Endpoint
```go
//spidey:route GET /api/chat
func ChatSocket(c *core.Context) {
	// Upgrade the connection
	conn, err := c.Upgrade()
	if err != nil {
		return
	}
	
	// Subscribe this connection to a specific room
	conn.Subscribe("room:global")
}
```

### 2. Broadcasting Messages
You can broadcast HTML directly to connected clients from anywhere in your application!
```go
// Broadcast an OOB HTML snippet to everyone in 'room:global'
html := `<div id="messages" s-swap-oob="beforeend"><p>Hello World!</p></div>`
app.WS.BroadcastTo("room:global", []byte(html))
```

### 3. Scaling with Redis Pub/Sub
If you run multiple servers, you can easily scale your WebSockets using Redis!
```go
// In your main.go
app.WS.UseRedis("localhost:6379", "", 0)
```
Once connected, `app.WS.BroadcastTo()` will automatically use Redis Pub/Sub to reach clients across all your servers!

---

## The Frontend (`s-ws` Tags)

Spidey allows you to use WebSockets natively from HTML without writing a single line of JavaScript.

### 1. Connecting (`s-ws-connect`)
Attach this to a container element to open a persistent WebSocket connection.
```html
<div s-ws-connect="/api/chat" s-ws-reconnect="true" s-ws-queue="true">
    <!-- WebSocket connection active for children -->
</div>
```
- **`s-ws-reconnect="true"`**: Automatically reconnects using exponential backoff if the server restarts or the user loses Wi-Fi!
- **`s-ws-queue="true"`**: If the connection drops, it will queue outgoing messages in memory and send them instantly when the connection is restored!

### 2. Sending Data (`s-ws-send`)
Attach this to a form. When the user submits, Spidey intercepts the form, converts it to JSON, and sends it instantly over the WebSocket connection (bypassing normal HTTP requests).
```html
<div s-ws-connect="/api/chat">
    <form s-ws-send>
        <input type="text" name="message" placeholder="Type a message..." />
        <button type="submit">Send</button>
    </form>
</div>
```

### 3. Receiving DOM Updates (Real-Time HTML)
When the server broadcasts HTML (e.g., `app.WS.BroadcastTo(...)`), Spidey's frontend engine instantly intercepts it. If the HTML contains `s-swap-oob`, it dynamically updates the page!

If the server sends:
```html
<div id="messages" s-swap-oob="beforeend">
    <p>New message arrived!</p>
</div>
```
Spidey will automatically append that `<p>` tag inside the existing `<div id="messages">` on your page!

---

## Security & Cross-Origin Policies

By default, Spidey is **Secure-by-Default** against Cross-Site WebSocket Hijacking (CSWSH). 

When a client attempts to connect to your WebSocket endpoints, the framework enforces the **Same-Origin Policy**. If the `Origin` header of the incoming request does not perfectly match your server's host, the connection is instantly rejected.

If you are building a public API or a mobile application and explicitly need to allow cross-origin WebSocket connections, you can define an allow-list in your `spidey.config.json`:

```json
{
    "wsAllowedOrigins": [
        "https://app.example.com",
        "https://admin.example.com"
    ]
}
```

To explicitly allow *all* origins (use with extreme caution!), you can specify the wildcard `*`:

```json
{
    "wsAllowedOrigins": ["*"]
}
```
