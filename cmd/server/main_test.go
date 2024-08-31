package main

import (
	"flag"
	"os"
	"testing"
)

func TestMainFlagParsing(t *testing.T) {
	// Save original os.Args and flag.CommandLine
	oldArgs := os.Args
	oldFlagCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlagCommandLine
	}()

	tests := []struct {
		name           string
		args           []string
		expectedServer bool
		expectedInput  string
		expectedOutput string
		expectedPort   int
	}{
		{
			name:           "Default Local Mode",
			args:           []string{"cmd", "--input", "test.json"},
			expectedServer: false,
			expectedInput:  "test.json",
			expectedOutput: "output/calendar.ics",
			expectedPort:   8080,
		},
		{
			name:           "Server Mode",
			args:           []string{"cmd", "--server", "--port", "9000"},
			expectedServer: true,
			expectedInput:  "",
			expectedOutput: "output/calendar.ics",
			expectedPort:   9000,
		},
		{
			name:           "Local Mode with Custom Output",
			args:           []string{"cmd", "--input", "test.json", "--output", "custom.ics"},
			expectedServer: false,
			expectedInput:  "test.json",
			expectedOutput: "custom.ics",
			expectedPort:   8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags and args for each test
			os.Args = tt.args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Reinitialize flags
			serverMode = flag.Bool("server", false, "Run in server mode")
			inputFile = flag.String("input", "", "Input JSON file (for local mode)")
			outputFile = flag.String("output", "output/calendar.ics", "Output ICS file (for local mode)")
			port = flag.Int("port", 8080, "Port to run the server on (for server mode)")

			// Parse flags
			flag.Parse()

			// Check results
			if *serverMode != tt.expectedServer {
				t.Errorf("Expected server mode to be %v, got %v", tt.expectedServer, *serverMode)
			}
			if *inputFile != tt.expectedInput {
				t.Errorf("Expected input file to be %s, got %s", tt.expectedInput, *inputFile)
			}
			if *outputFile != tt.expectedOutput {
				t.Errorf("Expected output file to be %s, got %s", tt.expectedOutput, *outputFile)
			}
			if *port != tt.expectedPort {
				t.Errorf("Expected port to be %d, got %d", tt.expectedPort, *port)
			}
		})
	}
}
