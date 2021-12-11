About
=====

A simple command-line wrapper around `go.bug.st/serial`, for reading the
output of a serial device. I'm using it to read from an Arduino; to use it
with other serial devices you may need to make small changes to the defaults
I put in the `serial.Mode` struct.

Usage
-----

Compile using:

```
git clone github.com/yalue/serial_reader
cd serial_reader
go build .

# To list ports
./serial_reader

# To read from whatever was the first port in the list printed above at 115200
# baud, copying output to output_log.txt.
./serial_reader -port 1 -baud 115200 -log_file output_log.txt

# Press ctrl+C to gracefully exit.
```

Run with the `-help` flag for command-line options. With no arguments, the
program will print a list of available ports.

