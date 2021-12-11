// This is a simple command-line utility for printing the output of a serial
// device.  I'm only writing it because I want something cross-platform.
package main

import (
	"flag"
	"fmt"
	"go.bug.st/serial"
	"io"
	"os"
	"os/signal"
	"time"
)

// A quick type for copying output to a file in addition to stdout.
type logWriter struct {
	file *os.File
}

// Returns a logWriter with output going to the given filename. Returns e if
// any error occurs. The returned logWriter must be Close()'d when no longer
// needed.
func newLogWriter(filename string) (*logWriter, error) {
	f, e := os.Create(filename)
	if e != nil {
		return nil, fmt.Errorf("Failed creating log file %s: %w", filename, e)
	}
	return &logWriter{
		file: f,
	}, nil
}

func (w *logWriter) Close() error {
	if w.file == nil {
		return nil
	}
	e := w.file.Close()
	w.file = nil
	return e
}

func (w *logWriter) Write(data []byte) (int, error) {
	fmt.Printf("%s", data)
	return w.file.Write(data)
}

func run() int {
	var portNumber uint
	var baud int
	var logFile string
	flag.UintVar(&portNumber, "port", 0, "The selected serial port.")
	flag.IntVar(&baud, "baud", 9600, "The port's baud rate.")
	flag.StringVar(&logFile, "log_file", "", "An optional name for a file to"+
		" which serial output will be copied. (It will still be printed to "+
		"stdout.)")
	flag.Parse()
	portsList, e := serial.GetPortsList()
	if e != nil {
		fmt.Printf("Failed getting list of serial ports: %s\n", e)
		return 1
	}
	if len(portsList) == 0 {
		fmt.Printf("No serial devices were found.\n")
		return 0
	}
	if (portNumber == 0) || (portNumber > uint(len(portsList))) {
		fmt.Printf("You must select a valid serial port. Options:\n")
		for i := 0; i < len(portsList); i++ {
			fmt.Printf(" %d: %s\n", i+1, portsList[i])
		}
		fmt.Printf("Run with \"-help\" for more information.\n")
		return 1
	}

	// Connect to the requested port.
	portName := portsList[portNumber-1]
	mode := &serial.Mode{
		BaudRate: baud,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}
	port, e := serial.Open(portName, mode)
	if e != nil {
		fmt.Printf("Failed opening serial port %s: %s\n", portName, e)
		return 1
	}
	defer port.Close()
	port.SetReadTimeout(500 * time.Millisecond)

	// Create the log file if it was requested. We can't use the "log" package
	// since it inserts a newline after every write.
	var outputDst io.Writer
	if logFile != "" {
		logOutput, e := newLogWriter(logFile)
		if e != nil {
			fmt.Printf("Error establishing log file: %s\n", e)
			return 1
		}
		defer logOutput.Close()
		outputDst = logOutput
	} else {
		outputDst = os.Stdout
	}

	shouldExit := false
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		s, ok := <-c
		if !ok {
			// The channel closed early.
			return
		}
		fmt.Printf("Signal %s pressed. Exiting after next read...\n", s)
		shouldExit = true
	}()
	fmt.Printf("Reading from device %s. Press ctrl+C to quit.\n", portName)
	data := make([]byte, 512)
	bytesRead := 0
	// Read from the port until we get the exit signal or an error.
	for !shouldExit {
		bytesRead, e = port.Read(data)
		if e != nil {
			shouldExit = true
			close(c)
			fmt.Printf("Failed reading from %s: %s\n", portName, e)
			return 1
		}
		if bytesRead == 0 {
			continue
		}
		fmt.Fprintf(outputDst, "%s", data[0:bytesRead])
	}
	return 0
}

func main() {
	os.Exit(run())
}
