package msgpack_test

import (
	"bytes"
	"math"
	"reflect"
	"testing"

	"arcan-it.de/arclib/msgpack"
)

type testPayload struct {
	Name   string   `msgpack:"name"`
	Code   int      `msgpack:"code"`
	Active bool     `msgpack:"active"`
	Score  float64  `msgpack:"score"`
	Tags   []string `msgpack:"tags"`
}

type nested struct {
	Inner testPayload `msgpack:"inner"`
	Extra string      `msgpack:"extra"`
}

func TestMarshalUnmarshalStruct(t *testing.T) {
	orig := testPayload{
		Name:   "session_init",
		Code:   42,
		Active: true,
		Score:  3.14,
		Tags:   []string{"alpha", "beta"},
	}

	data, err := msgpack.Marshal(&orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Marshal returned empty data")
	}

	var got testPayload
	if err := msgpack.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Fatalf("roundtrip mismatch:\n got %+v\nwant %+v", got, orig)
	}
}

func TestMarshalUnmarshalMap(t *testing.T) {
	orig := map[string]interface{}{
		"error":   "not_found",
		"code":    int64(404),
		"details": nil,
	}

	data, err := msgpack.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got map[string]interface{}
	if err := msgpack.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got["error"] != "not_found" {
		t.Fatalf("expected error=not_found, got %v", got["error"])
	}
	if got["details"] != nil {
		t.Fatalf("expected details=nil, got %v", got["details"])
	}
}

func TestMarshalUnmarshalTypes(t *testing.T) {
	tests := []struct {
		name string
		val  interface{}
	}{
		{"string", "hello"},
		{"int", 12345},
		{"float64", 2.718},
		{"float64_max", math.MaxFloat64},
		{"bool_true", true},
		{"bool_false", false},
		{"nil", (*string)(nil)},
		{"slice_int", []int{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := msgpack.Marshal(tt.val)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if len(data) == 0 {
				t.Fatal("Marshal returned empty bytes")
			}
			// Just verify it decodes without error into interface{}
			var got interface{}
			if err := msgpack.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
		})
	}
}

func TestMarshalUnmarshalNested(t *testing.T) {
	orig := nested{
		Inner: testPayload{
			Name:   "inner_payload",
			Code:   7,
			Active: false,
			Score:  0.5,
			Tags:   []string{"nested"},
		},
		Extra: "metadata",
	}

	data, err := msgpack.Marshal(&orig)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got nested
	if err := msgpack.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !reflect.DeepEqual(orig, got) {
		t.Fatalf("roundtrip mismatch:\n got %+v\nwant %+v", got, orig)
	}
}

func TestEncoderDecoder(t *testing.T) {
	var buf bytes.Buffer

	enc := msgpack.NewEncoder(&buf)
	payloads := []testPayload{
		{Name: "first", Code: 1, Active: true, Score: 1.0, Tags: nil},
		{Name: "second", Code: 2, Active: false, Score: 2.0, Tags: []string{"x"}},
		{Name: "third", Code: 3, Active: true, Score: 3.0, Tags: []string{"a", "b"}},
	}

	for i := range payloads {
		if err := enc.Encode(&payloads[i]); err != nil {
			t.Fatalf("Encode[%d]: %v", i, err)
		}
	}

	dec := msgpack.NewDecoder(&buf)
	for i := range payloads {
		var got testPayload
		if err := dec.Decode(&got); err != nil {
			t.Fatalf("Decode[%d]: %v", i, err)
		}
		if !reflect.DeepEqual(payloads[i], got) {
			t.Fatalf("stream mismatch at %d:\n got %+v\nwant %+v", i, got, payloads[i])
		}
	}
}

func TestUnmarshalInvalidData(t *testing.T) {
	var got testPayload
	err := msgpack.Unmarshal([]byte{0xff, 0xfe, 0xfd}, &got)
	if err == nil {
		t.Fatal("expected error for invalid msgpack data")
	}
}
