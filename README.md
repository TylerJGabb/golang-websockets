# Websocket Server from Scratch

This is a simple websocket server written in golang. See [RFC6455](https://tools.ietf.org/html/rfc6455) for the websocket protocol specification.

It handles the handshake and the frame parsing, and only echos back a slice of the received message.

If the client sends a close frame, it will close the connection gracefully, following the rules specified in [RFC6455](https://tools.ietf.org/html/rfc6455#section-5.5.1)
