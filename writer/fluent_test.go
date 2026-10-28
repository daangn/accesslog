package writer

import (
	"encoding/json"
	"io"
	"net"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/fluent/fluent-logger-golang/fluent"
)

type streamCapture struct {
	data []byte
	err  error
}

type jsonCapture struct {
	tag     string
	record  map[string]interface{}
	options map[string]string
	err     error
}

type messagePackCapture struct {
	message fluent.MessageExt
	err     error
}

func startFluentListener(t *testing.T) (net.Listener, string, int) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	host, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	return listener, host, port
}

func captureMessagePack(listener net.Listener) <-chan streamCapture {
	result := make(chan streamCapture, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			result <- streamCapture{err: err}
			return
		}
		defer conn.Close()
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		data, err := io.ReadAll(conn)
		result <- streamCapture{data: data, err: err}
	}()
	return result
}

func captureMessagePackAck(listener net.Listener) <-chan messagePackCapture {
	result := make(chan messagePackCapture, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			result <- messagePackCapture{err: err}
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

		data := make([]byte, 0, 1024)
		buffer := make([]byte, 1024)
		var message fluent.MessageExt
		for {
			n, readErr := conn.Read(buffer)
			if n > 0 {
				data = append(data, buffer[:n]...)
				var candidate fluent.MessageExt
				rest, decodeErr := candidate.UnmarshalMsg(data)
				if decodeErr == nil {
					if len(rest) != 0 {
						result <- messagePackCapture{err: io.ErrUnexpectedEOF}
						return
					}
					message = candidate
					break
				}
			}
			if readErr != nil {
				result <- messagePackCapture{err: readErr}
				return
			}
		}

		ack, err := (fluent.AckResp{Ack: message.Option["chunk"]}).MarshalMsg(nil)
		if err != nil {
			result <- messagePackCapture{err: err}
			return
		}
		for len(ack) > 0 {
			n, writeErr := conn.Write(ack)
			if writeErr != nil {
				result <- messagePackCapture{err: writeErr}
				return
			}
			if n == 0 {
				result <- messagePackCapture{err: io.ErrShortWrite}
				return
			}
			ack = ack[n:]
		}
		result <- messagePackCapture{message: message}
		_, _ = io.Copy(io.Discard, conn)
	}()
	return result
}

func captureJSON(listener net.Listener, acknowledge bool) <-chan jsonCapture {
	result := make(chan jsonCapture, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			result <- jsonCapture{err: err}
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))

		var fields []json.RawMessage
		if err := json.NewDecoder(conn).Decode(&fields); err != nil {
			if acknowledge {
				_ = json.NewEncoder(conn).Encode(map[string]string{"ack": ""})
			}
			result <- jsonCapture{err: err}
			return
		}
		if len(fields) != 4 {
			result <- jsonCapture{err: io.ErrUnexpectedEOF}
			return
		}

		var event jsonCapture
		if err := json.Unmarshal(fields[0], &event.tag); err != nil {
			result <- jsonCapture{err: err}
			return
		}
		if err := json.Unmarshal(fields[2], &event.record); err != nil {
			result <- jsonCapture{err: err}
			return
		}
		if err := json.Unmarshal(fields[3], &event.options); err != nil {
			result <- jsonCapture{err: err}
			return
		}
		if acknowledge {
			_ = json.NewEncoder(conn).Encode(map[string]string{"ack": event.options["chunk"]})
		}
		result <- event
		_, _ = io.Copy(io.Discard, conn)
	}()
	return result
}

func receiveStream(t *testing.T, result <-chan streamCapture) streamCapture {
	t.Helper()
	select {
	case capture := <-result:
		if capture.err != nil {
			t.Fatal(capture.err)
		}
		return capture
	case <-time.After(6 * time.Second):
		t.Fatal("timed out waiting for Fluent stream")
		return streamCapture{}
	}
}

func receiveMessagePackAck(t *testing.T, result <-chan messagePackCapture) fluent.MessageExt {
	t.Helper()
	select {
	case capture := <-result:
		if capture.err != nil {
			t.Fatal(capture.err)
		}
		return capture.message
	case <-time.After(6 * time.Second):
		t.Fatal("timed out waiting for Fluent MessagePack acknowledgment")
		return fluent.MessageExt{}
	}
}

func receiveJSON(t *testing.T, result <-chan jsonCapture) jsonCapture {
	t.Helper()
	select {
	case capture := <-result:
		if capture.err != nil {
			t.Fatal(capture.err)
		}
		return capture
	case <-time.After(6 * time.Second):
		t.Fatal("timed out waiting for Fluent JSON message")
		return jsonCapture{}
	}
}

func decodeMessages(t *testing.T, data []byte) []fluent.Message {
	t.Helper()
	var messages []fluent.Message
	for len(data) != 0 {
		var message fluent.Message
		rest, err := message.UnmarshalMsg(data)
		if err != nil {
			t.Fatal(err)
		}
		if len(rest) == len(data) {
			t.Fatal("MessagePack decoder made no progress")
		}
		messages = append(messages, message)
		data = rest
	}
	return messages
}

func newTestForwarder(t *testing.T, host string, port int, config fluent.Config) *FluentLogWriter {
	t.Helper()
	config.FluentHost = host
	config.FluentPort = port
	forwarder, err := fluent.New(config)
	if err != nil {
		t.Fatal(err)
	}
	return NewFluentLogWriterFromForwarder("access", forwarder)
}

func TestFluentLogWriterWritePreservesRecords(t *testing.T) {
	listener, host, port := startFluentListener(t)
	captured := captureMessagePack(listener)
	logWriter, err := NewFluentLogWriter("access", host, port)
	if err != nil {
		t.Fatal(err)
	}
	defer logWriter.Close()

	input := []byte(`{"name":"request","count":2.5,"enabled":true,"nested":{"region":"west","nil":null},"items":[1,false,"x"]}`)
	n, err := logWriter.Write(input)
	if err != nil || n != len(input) {
		t.Fatalf("Write() = (%d, %v), want (%d, nil)", n, err, len(input))
	}
	if n, err := logWriter.Write([]byte("null")); err != nil || n != len("null") {
		t.Fatalf("Write(null) = (%d, %v), want (%d, nil)", n, err, len("null"))
	}
	if err := logWriter.Close(); err != nil {
		t.Fatal(err)
	}

	messages := decodeMessages(t, receiveStream(t, captured).data)
	if len(messages) != 2 {
		t.Fatalf("received %d messages, want 2", len(messages))
	}
	if messages[0].Tag != "access" {
		t.Errorf("tag = %q, want %q", messages[0].Tag, "access")
	}
	want := map[string]interface{}{
		"name":    "request",
		"count":   float64(2.5),
		"enabled": true,
		"nested": map[string]interface{}{
			"region": "west",
			"nil":    nil,
		},
		"items": []interface{}{float64(1), false, "x"},
	}
	if !reflect.DeepEqual(messages[0].Record, want) {
		t.Errorf("record = %#v, want %#v", messages[0].Record, want)
	}
	if got, ok := messages[1].Record.(map[string]interface{}); !ok || len(got) != 0 {
		t.Errorf("top-level null record = %#v, want an empty object", messages[1].Record)
	}
}

func TestFluentLogWriterWritePreservesJSONForwarderOptions(t *testing.T) {
	listener, host, port := startFluentListener(t)
	captured := captureJSON(listener, true)
	logWriter := newTestForwarder(t, host, port, fluent.Config{
		TagPrefix:     "prefix",
		MarshalAsJSON: true,
		RequestAck:    true,
	})
	defer logWriter.Close()

	input := []byte(`{"name":"request","count":3.5,"nested":{"ok":true}}`)
	n, err := logWriter.Write(input)
	if err != nil || n != len(input) {
		t.Fatalf("Write() = (%d, %v), want (%d, nil)", n, err, len(input))
	}
	if err := logWriter.Close(); err != nil {
		t.Fatal(err)
	}

	message := receiveJSON(t, captured)
	if message.tag != "prefix.access" {
		t.Errorf("tag = %q, want %q", message.tag, "prefix.access")
	}
	want := map[string]interface{}{
		"name":  "request",
		"count": float64(3.5),
		"nested": map[string]interface{}{
			"ok": true,
		},
	}
	if !reflect.DeepEqual(message.record, want) {
		t.Errorf("record = %#v, want %#v", message.record, want)
	}
	if message.options["chunk"] == "" {
		t.Error("acknowledgment chunk option is missing")
	}
}

func TestFluentLogWriterWritePreservesJSONNullRecord(t *testing.T) {
	listener, host, port := startFluentListener(t)
	captured := captureJSON(listener, false)
	logWriter := newTestForwarder(t, host, port, fluent.Config{MarshalAsJSON: true})
	defer logWriter.Close()

	if n, err := logWriter.Write([]byte("null")); err != nil || n != len("null") {
		t.Fatalf("Write(null) = (%d, %v), want (%d, nil)", n, err, len("null"))
	}
	if err := logWriter.Close(); err != nil {
		t.Fatal(err)
	}

	message := receiveJSON(t, captured)
	if got := message.record; !reflect.DeepEqual(got, map[string]interface{}{}) {
		t.Errorf("top-level null record = %#v, want an empty object", got)
	}
}

func TestFluentLogWriterWriteUsesEventTimeFormatAndAck(t *testing.T) {
	listener, host, port := startFluentListener(t)
	captured := captureMessagePackAck(listener)
	logWriter := newTestForwarder(t, host, port, fluent.Config{
		TagPrefix:          "prefix",
		RequestAck:         true,
		SubSecondPrecision: true,
	})
	defer logWriter.Close()

	input := []byte(`{"name":"request","value":4.5}`)
	n, err := logWriter.Write(input)
	if err != nil || n != len(input) {
		t.Fatalf("Write() = (%d, %v), want (%d, nil)", n, err, len(input))
	}
	if err := logWriter.Close(); err != nil {
		t.Fatal(err)
	}

	// MessageExt requires the EventTime extension.
	message := receiveMessagePackAck(t, captured)
	if message.Tag != "prefix.access" {
		t.Errorf("tag = %q, want %q", message.Tag, "prefix.access")
	}
	if message.Option["chunk"] == "" {
		t.Error("acknowledgment chunk option is missing")
	}
	want := map[string]interface{}{"name": "request", "value": float64(4.5)}
	if !reflect.DeepEqual(message.Record, want) {
		t.Errorf("record = %#v, want %#v", message.Record, want)
	}
}

func TestFluentLogWriterWriteRejectsInvalidJSONAndWrapsForwarderErrors(t *testing.T) {
	forwarder := &fluent.Fluent{Config: fluent.Config{FluentNetwork: "invalid", MaxRetry: 1}}
	logWriter := NewFluentLogWriterFromForwarder("access", forwarder)
	defer logWriter.Close()

	for _, input := range []string{`{"broken":`, `{"ok":true} {"trailing":true}`} {
		n, err := logWriter.Write([]byte(input))
		if n != 0 || err == nil {
			t.Errorf("Write(%q) = (%d, %v), want (0, error)", input, n, err)
			continue
		}
		if !strings.HasPrefix(err.Error(), "fluent logger write: ") {
			t.Errorf("Write(%q) error = %q, want fluent logger prefix", input, err)
		}
	}

	n, err := logWriter.Write([]byte(`{"ok":true}`))
	if n != 0 || err == nil {
		t.Fatalf("Write(valid JSON) with a failing forwarder = (%d, %v), want (0, error)", n, err)
	}
	if !strings.HasPrefix(err.Error(), "fluent logger write: ") || !strings.Contains(err.Error(), "unknown network invalid") {
		t.Errorf("forwarder error = %q, want wrapped unknown-network error", err)
	}
}
