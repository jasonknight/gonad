# Gonad

A simple TCP logging server. 

Have you ever tried to setup something like syslog-ng? In a Docker container? Yeah.

This is a simple, small, go application that just takes whatever you send it and logs
it. It doesn't parse it, futz with it, or wash your bum. It's designed to do exactly
one thing, and do it quickly - and be easy to set up.

This app follows the 12 factor environment based configuration.

# TODO
- Add support for multiple ports, or generally multiple listeners.
- Discover and add basic support for any standard logging server methodologies.
- Add support for deadlines
