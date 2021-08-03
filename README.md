# Log Server

### NOTE: This is intended as a development tool only! It is not secure!

Have you tried aggregating and interpreting really verbose logs for multiple services in one place all at once? It can be a pain.

This provides a way to collect all relevant information in one place, even from multiple running processes.

## Protocol
Currently, the protocol is simple JSON over TCP.

The server expects the client to specify which protocol version it's using, to ensure that it's compatible with the server's version.

Then arbitrary strings may be sent to the server to be printed to its stdout in the order received. Each client connection
will be serviced in its own goroutine, and serial ordering is maintained by a single log writer pulling from a common channel.

## Current Client List
* [Java Client](https://github.com/drognisep/logclient-java)
