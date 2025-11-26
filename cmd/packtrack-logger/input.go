package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	packtrack "github.com/commandant-labs/pack-track-sdk"
)

func readEventsFromFile(path string, ndjson bool) ([]packtrack.Event, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if ndjson {
		return readNDJSON(f)
	}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	// Try array first
	var arr []packtrack.Event
	if err := json.Unmarshal(b, &arr); err == nil {
		return arr, nil
	}
	// Try single object
	var e packtrack.Event
	if err := json.Unmarshal(b, &e); err == nil {
		return []packtrack.Event{e}, nil
	}
	return nil, fmt.Errorf("file is not valid JSON event or array")
}

func readEventsFromStdin(ndjson bool) ([]packtrack.Event, error) {
	if ndjson {
		return readNDJSON(os.Stdin)
	}
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	var e packtrack.Event
	if err := json.Unmarshal(b, &e); err == nil {
		return []packtrack.Event{e}, nil
	}
	var arr []packtrack.Event
	if err := json.Unmarshal(b, &arr); err == nil {
		return arr, nil
	}
	return nil, fmt.Errorf("STDIN is not valid JSON event or array; use --ndjson for newline-delimited input")
}

func readNDJSON(r io.Reader) ([]packtrack.Event, error) {
	s := bufio.NewScanner(r)
	res := make([]packtrack.Event, 0, 128)
	for s.Scan() {
		line := s.Bytes()
		if len(bytesTrimSpace(line)) == 0 {
			continue
		}
		var e packtrack.Event
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("invalid NDJSON line: %w", err)
		}
		res = append(res, e)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return res, nil
}

func bytesTrimSpace(b []byte) []byte {
	i, j := 0, len(b)
	for i < j && (b[i] == ' ' || b[i] == '\n' || b[i] == '\r' || b[i] == '\t') {
		i++
	}
	for i < j && (b[j-1] == ' ' || b[j-1] == '\n' || b[j-1] == '\r' || b[j-1] == '\t') {
		j--
	}
	return b[i:j]
}
