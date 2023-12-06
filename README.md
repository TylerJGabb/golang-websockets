# Websocket Server from Scratch

This is a simple websocket server written in golang. See [RFC6455](https://tools.ietf.org/html/rfc6455) for the websocket protocol specification.

It handles the handshake and the frame parsing, and only echos back a slice of the received message.

If the client sends a close frame, it will close the connection gracefully, following the rules specified in [RFC6455](https://tools.ietf.org/html/rfc6455#section-5.5.1)

## Whats Next

I want to visit concurrency, and find out how multiple connections at the same time function.
According to documentation, the http library starts a new goroutine for every incoming request, so each connection
that is hijacked from said request should be able to perform read/writes on the connection without blocking the other connections.

In order to gain visibility I want to start using structured logging. Json format is fine, idealy I want key value pairs.

