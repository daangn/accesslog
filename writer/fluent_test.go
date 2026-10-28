package writer_test

import (
	"context"
	"math"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/daangn/accesslog"
	"github.com/daangn/accesslog/middleware"
	"github.com/daangn/accesslog/writer"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/rs/zerolog"
	"github.com/tinylib/msgp/msgp"
	"google.golang.org/grpc"
)

const (
	minInt64  = int64(-1 << 63)
	maxInt64  = int64(1<<63 - 1)
	maxUint64 = uint64(1<<64 - 1)
)

type receivedMessage struct {
	message fluent.Message
	err     error
}

type fluentFixture struct {
	writer   *writer.FluentLogWriter
	messages chan receivedMessage
}

type signedLogObject struct {
	value int64
}

func (o signedLogObject) MarshalZerologObject(e *zerolog.Event) {
	e.Int64("signed", o.value)
}

func TestFluentLogWriterPreservesIntegerValues(t *testing.T) {
	signedFields := []struct {
		name  string
		value int64
	}{
		{name: "below_2_53", value: (1 << 53) - 1},
		{name: "at_2_53", value: 1 << 53},
		{name: "above_2_53", value: (1 << 53) + 1},
		{name: "negative_2_53_minus_one", value: -((1 << 53) - 1)},
		{name: "negative_2_53", value: -(1 << 53)},
		{name: "negative_2_53_plus_one", value: -((1 << 53) + 1)},
		{name: "int64_min", value: minInt64},
		{name: "int64_min_plus_one", value: minInt64 + 1},
		{name: "int64_max_minus_one", value: maxInt64 - 1},
		{name: "int64_max", value: maxInt64},
	}
	unsignedFields := []struct {
		name  string
		value uint64
	}{
		{name: "uint64_above_2_53", value: uint64((1 << 53) + 1)},
		{name: "uint64_max_minus_one", value: maxUint64 - 1},
		{name: "uint64_max", value: maxUint64},
	}

	fixture := newFluentFixture(t)
	result, callErr, callbacks := writeGRPCLog(t, fixture, func(e *zerolog.Event) {
		for _, field := range signedFields {
			e.Int64(field.name, field.value)
		}
		for _, field := range unsignedFields {
			e.Uint64(field.name, field.value)
		}
		e.Dict("nested", zerolog.Dict().
			Int64("signed", (1<<53)+1).
			Uint64("unsigned", maxUint64).
			Array("values", zerolog.Arr().Int64(minInt64).Float64(0.5)),
		)
		e.Array("array", zerolog.Arr().
			Int64((1<<53)+1).
			Uint64(maxUint64).
			Float64(0.25).
			Object(signedLogObject{value: minInt64}),
		)
	})
	if callErr != nil || result != "handled" || callbacks != 1 {
		t.Fatalf("gRPC setup = (%v, %v), callback count %d; want (handled, nil), one callback", result, callErr, callbacks)
	}

	message := fixture.receive(t)
	record := recordFields(t, message)
	for _, field := range signedFields {
		assertExactInteger(t, field.name, record[field.name], big.NewInt(field.value))
	}
	for _, field := range unsignedFields {
		assertExactInteger(t, field.name, record[field.name], new(big.Int).SetUint64(field.value))
	}

	nested, ok := record["nested"].(map[string]interface{})
	if !ok {
		t.Errorf("integer value mismatch for %q: got %T, want nested object", "nested", record["nested"])
	} else {
		assertExactInteger(t, "nested.signed", nested["signed"], big.NewInt((1<<53)+1))
		assertExactInteger(t, "nested.unsigned", nested["unsigned"], new(big.Int).SetUint64(maxUint64))
		nestedValues, ok := nested["values"].([]interface{})
		if !ok || len(nestedValues) != 2 {
			t.Errorf("integer value mismatch for %q: got %T, want two-element array", "nested.values", nested["values"])
		} else {
			assertExactInteger(t, "nested.values[0]", nestedValues[0], big.NewInt(minInt64))
		}
	}

	array, ok := record["array"].([]interface{})
	if !ok || len(array) != 4 {
		t.Errorf("integer value mismatch for %q: got %T, want four-element array", "array", record["array"])
	} else {
		assertExactInteger(t, "array[0]", array[0], big.NewInt((1<<53)+1))
		assertExactInteger(t, "array[1]", array[1], new(big.Int).SetUint64(maxUint64))
		arrayObject, ok := array[3].(map[string]interface{})
		if !ok {
			t.Errorf("integer value mismatch for %q: got %T, want nested object", "array[3]", array[3])
		} else {
			assertExactInteger(t, "array[3].signed", arrayObject["signed"], big.NewInt(minInt64))
		}
	}
}

func TestFluentLogWriterPreservesOtherBehavior(t *testing.T) {
	fixture := newFluentFixture(t)
	for _, input := range []string{
		`{"field":`,
		`{"field":1} {"other":2}`,
		`{"field":1e400,"field":0}`,
	} {
		if n, err := fixture.writer.Write([]byte(input)); err == nil || n != 0 {
			t.Fatalf("Write(%q) = (%d, %v), want an error and zero bytes", input, n, err)
		}
	}

	result, callErr, callbacks := writeGRPCLog(t, fixture, func(e *zerolog.Event) {
		e.Int64("ordinary_signed", -42)
		e.Uint64("ordinary_unsigned", 42)
		e.Float64("fractional", 1.25)
		e.Float64("negative_zero", math.Copysign(0, -1))
	})
	if callErr != nil || result != "handled" || callbacks != 1 {
		t.Fatalf("gRPC result = (%v, %v), callback count %d; want (handled, nil), one callback", result, callErr, callbacks)
	}

	message := fixture.receive(t)
	if message.Tag != "test.accesslog" {
		t.Errorf("Fluent tag = %q, want test.accesslog", message.Tag)
	}
	record := recordFields(t, message)
	if got, want := record["protocol"], "grpc"; got != want {
		t.Errorf("protocol = %v, want %q", got, want)
	}
	if got, want := record["method"], "/test.Service/Method"; got != want {
		t.Errorf("method = %v, want %q", got, want)
	}
	if got, want := record["status"], "OK"; got != want {
		t.Errorf("status = %v, want %q", got, want)
	}
	for _, key := range []string{"req", "res"} {
		if _, ok := record[key]; ok {
			t.Errorf("unexpected %q field with default gRPC formatter", key)
		}
	}
	assertExactInteger(t, "ordinary_signed", record["ordinary_signed"], big.NewInt(-42))
	assertExactInteger(t, "ordinary_unsigned", record["ordinary_unsigned"], big.NewInt(42))
	assertFloat(t, "fractional", record["fractional"], 1.25)
	negativeZero, ok := record["negative_zero"].(float64)
	if !ok || !math.Signbit(negativeZero) {
		t.Errorf("negative_zero = %T(%v), want a negative float zero", record["negative_zero"], record["negative_zero"])
	}
}

func newFluentFixture(t *testing.T) *fluentFixture {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	fixture := &fluentFixture{
		messages: make(chan receivedMessage, 1),
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			fixture.messages <- receivedMessage{err: err}
			return
		}
		defer conn.Close()
		if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			fixture.messages <- receivedMessage{err: err}
			return
		}
		var message fluent.Message
		if err := message.DecodeMsg(msgp.NewReader(conn)); err != nil {
			fixture.messages <- receivedMessage{err: err}
			return
		}
		fixture.messages <- receivedMessage{message: message}
	}()

	port := listener.Addr().(*net.TCPAddr).Port
	fixture.writer, err = writer.NewFluentLogWriter("test.accesslog", "127.0.0.1", port)
	if err != nil {
		listener.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := fixture.writer.Close(); err != nil {
			t.Errorf("close Fluent log writer: %v", err)
		}
		if err := listener.Close(); err != nil {
			t.Errorf("close Fluent receiver: %v", err)
		}
	})
	return fixture
}

func (fixture *fluentFixture) receive(t *testing.T) fluent.Message {
	t.Helper()
	if err := fixture.writer.Close(); err != nil {
		t.Fatalf("close Fluent log writer: %v", err)
	}
	select {
	case message := <-fixture.messages:
		if message.err != nil {
			t.Fatalf("receive Fluent record: %v", message.err)
		}
		return message.message
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the Fluent record")
		return fluent.Message{}
	}
}

func writeGRPCLog(t *testing.T, fixture *fluentFixture, addFields func(*zerolog.Event)) (interface{}, error, int) {
	t.Helper()
	var callbacks int
	interceptor := middleware.UnaryServerInterceptor(accesslog.NewGRPCLogger(
		fixture.writer,
		accesslog.NewDefaultGRPCLogFormatter(),
	))
	result, callErr := interceptor(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"},
		func(ctx context.Context, _ interface{}) (interface{}, error) {
			entry := accesslog.GetLogEntry(ctx)
			if entry == nil {
				t.Fatal("gRPC handler did not receive a log entry")
			}
			entry.Add(func(e *zerolog.Event) {
				callbacks++
				addFields(e)
			})
			return "handled", nil
		},
	)
	return result, callErr, callbacks
}

func recordFields(t *testing.T, message fluent.Message) map[string]interface{} {
	t.Helper()
	record, ok := message.Record.(map[string]interface{})
	if !ok {
		t.Fatalf("Fluent record has type %T, want map[string]interface{}", message.Record)
	}
	return record
}

func assertExactInteger(t *testing.T, name string, got interface{}, want *big.Int) {
	t.Helper()
	if exactInteger(got, want) {
		return
	}
	t.Errorf("integer value mismatch for %q: got %T(%v), want %s", name, got, got, want.String())
}

func exactInteger(got interface{}, want *big.Int) bool {
	switch value := got.(type) {
	case int64:
		return big.NewInt(value).Cmp(want) == 0
	case uint64:
		return new(big.Int).SetUint64(value).Cmp(want) == 0
	case float32:
		return exactFloatInteger(float64(value), want)
	case float64:
		return exactFloatInteger(value, want)
	default:
		return false
	}
}

func exactFloatInteger(value float64, want *big.Int) bool {
	rational := new(big.Rat).SetFloat64(value)
	return rational != nil && rational.IsInt() && rational.Num().Cmp(want) == 0
}

func assertFloat(t *testing.T, name string, got interface{}, want float64) {
	t.Helper()
	value, ok := got.(float64)
	if !ok || value != want {
		t.Errorf("%s = %T(%v), want float64(%v)", name, got, got, want)
	}
}
