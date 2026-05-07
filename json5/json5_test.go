package json5

import (
	"testing"
)

func TestStandardJSON(t *testing.T) {
	input := `{"name": "test", "value": 42, "active": true, "items": [1, 2, 3]}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("standard JSON failed: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("name = %v, want test", result["name"])
	}
	if result["value"].(float64) != 42 {
		t.Errorf("value = %v, want 42", result["value"])
	}
	if result["active"] != true {
		t.Errorf("active = %v, want true", result["active"])
	}
}

func TestSingleLineComments(t *testing.T) {
	input := `{
		// This is a comment
		"name": "test", // inline comment
		"value": 42
		// trailing comment
	}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("single-line comments failed: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("name = %v, want test", result["name"])
	}
	if result["value"].(float64) != 42 {
		t.Errorf("value = %v, want 42", result["value"])
	}
}

func TestMultiLineComments(t *testing.T) {
	input := `{
		/* This is a
		   multi-line comment */
		"name": "test",
		"value": /* inline block comment */ 42
	}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("multi-line comments failed: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("name = %v, want test", result["name"])
	}
	if result["value"].(float64) != 42 {
		t.Errorf("value = %v, want 42", result["value"])
	}
}

func TestTrailingCommaObject(t *testing.T) {
	input := `{"a": 1, "b": 2,}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("trailing comma in object failed: %v", err)
	}
	if result["a"].(float64) != 1 {
		t.Errorf("a = %v, want 1", result["a"])
	}
	if result["b"].(float64) != 2 {
		t.Errorf("b = %v, want 2", result["b"])
	}
}

func TestTrailingCommaArray(t *testing.T) {
	input := `{"items": [1, 2, 3,]}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("trailing comma in array failed: %v", err)
	}
	items := result["items"].([]interface{})
	if len(items) != 3 {
		t.Errorf("len(items) = %d, want 3", len(items))
	}
}

func TestTrailingCommaWithComment(t *testing.T) {
	input := `{
		"a": 1,
		"b": 2, // trailing comma with comment
	}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("trailing comma with comment failed: %v", err)
	}
	if result["b"].(float64) != 2 {
		t.Errorf("b = %v, want 2", result["b"])
	}
}

func TestNestedTrailingCommas(t *testing.T) {
	input := `{
		"obj": {"x": 1, "y": 2,},
		"arr": [1, 2,],
	}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("nested trailing commas failed: %v", err)
	}
	obj := result["obj"].(map[string]interface{})
	if obj["x"].(float64) != 1 {
		t.Errorf("obj.x = %v, want 1", obj["x"])
	}
}

func TestUnquotedKeys(t *testing.T) {
	input := `{name: "test", value: 42, _private: true, $special: "yes"}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("unquoted keys failed: %v", err)
	}
	if result["name"] != "test" {
		t.Errorf("name = %v, want test", result["name"])
	}
	if result["value"].(float64) != 42 {
		t.Errorf("value = %v, want 42", result["value"])
	}
	if result["_private"] != true {
		t.Errorf("_private = %v, want true", result["_private"])
	}
	if result["$special"] != "yes" {
		t.Errorf("$special = %v, want yes", result["$special"])
	}
}

func TestSingleQuotedStrings(t *testing.T) {
	input := `{"name": 'hello world', "msg": 'it\'s fine'}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("single-quoted strings failed: %v", err)
	}
	if result["name"] != "hello world" {
		t.Errorf("name = %v, want 'hello world'", result["name"])
	}
	if result["msg"] != "it's fine" {
		t.Errorf("msg = %v, want \"it's fine\"", result["msg"])
	}
}

func TestSingleQuotedWithDoubleQuoteInside(t *testing.T) {
	input := `{"msg": 'say "hello"'}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("single-quoted with double quote failed: %v", err)
	}
	if result["msg"] != `say "hello"` {
		t.Errorf("msg = %v, want say \"hello\"", result["msg"])
	}
}

func TestHexNumbers(t *testing.T) {
	input := `{"val": 0xFF, "big": 0x1A2B}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("hex numbers failed: %v", err)
	}
	if result["val"].(float64) != 255 {
		t.Errorf("val = %v, want 255", result["val"])
	}
	if result["big"].(float64) != 6699 {
		t.Errorf("big = %v, want 6699", result["big"])
	}
}

func TestInfinity(t *testing.T) {
	input := `{"pos": Infinity, "neg": -Infinity}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("Infinity failed: %v", err)
	}
	if result["pos"] != nil {
		t.Errorf("pos = %v, want nil", result["pos"])
	}
	if result["neg"] != nil {
		t.Errorf("neg = %v, want nil", result["neg"])
	}
}

func TestNaN(t *testing.T) {
	input := `{"val": NaN}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("NaN failed: %v", err)
	}
	if result["val"] != nil {
		t.Errorf("val = %v, want nil", result["val"])
	}
}

func TestLeadingDecimalPoint(t *testing.T) {
	input := `{"val": .5, "neg": -.75}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("leading decimal point failed: %v", err)
	}
	if result["val"].(float64) != 0.5 {
		t.Errorf("val = %v, want 0.5", result["val"])
	}
	if result["neg"].(float64) != -0.75 {
		t.Errorf("neg = %v, want -0.75", result["neg"])
	}
}

func TestTrailingDecimalPoint(t *testing.T) {
	input := `{"val": 5., "big": 100.}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("trailing decimal point failed: %v", err)
	}
	if result["val"].(float64) != 5.0 {
		t.Errorf("val = %v, want 5.0", result["val"])
	}
	if result["big"].(float64) != 100.0 {
		t.Errorf("big = %v, want 100.0", result["big"])
	}
}

func TestPositiveSign(t *testing.T) {
	input := `{"val": +42, "dec": +.5}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("positive sign failed: %v", err)
	}
	if result["val"].(float64) != 42 {
		t.Errorf("val = %v, want 42", result["val"])
	}
	if result["dec"].(float64) != 0.5 {
		t.Errorf("dec = %v, want 0.5", result["dec"])
	}
}

func TestPositiveInfinity(t *testing.T) {
	input := `{"val": +Infinity}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("+Infinity failed: %v", err)
	}
	if result["val"] != nil {
		t.Errorf("val = %v, want nil", result["val"])
	}
}

func TestMultilineString(t *testing.T) {
	input := "{\"msg\": \"hello \\\nworld\"}"
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("multiline string failed: %v", err)
	}
	if result["msg"] != "hello world" {
		t.Errorf("msg = %q, want \"hello world\"", result["msg"])
	}
}

func TestUnmarshalToStruct(t *testing.T) {
	type Config struct {
		Host string `json:"host"`
		Port int    `json:"port"`
		SSL  bool   `json:"ssl"`
	}

	input := `{
		// Server configuration
		host: 'localhost',
		port: 8080,
		ssl: false,
	}`

	var cfg Config
	err := Unmarshal([]byte(input), &cfg)
	if err != nil {
		t.Fatalf("unmarshal to struct failed: %v", err)
	}
	if cfg.Host != "localhost" {
		t.Errorf("Host = %v, want localhost", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %v, want 8080", cfg.Port)
	}
	if cfg.SSL != false {
		t.Errorf("SSL = %v, want false", cfg.SSL)
	}
}

func TestRealisticConfigFile(t *testing.T) {
	input := `{
		// ArcHub Configuration
		/* Server settings */
		host: 'hub.arcan.de',
		port: 443,
		ssl: true,

		// Database
		database: {
			driver: 'postgres',
			host: 'localhost',
			port: 5432,
			name: 'archub',
			maxConns: 0x14, // 20 connections
		},

		// Features
		features: [
			'schemas',
			'wizards',
			'profiles',
		],

		// Rate limits
		rateLimit: {
			requestsPerMinute: 60,
			burstSize: +10,
			windowSize: .5, // half second
		},
	}`

	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("realistic config failed: %v", err)
	}

	if result["host"] != "hub.arcan.de" {
		t.Errorf("host = %v, want hub.arcan.de", result["host"])
	}
	if result["port"].(float64) != 443 {
		t.Errorf("port = %v, want 443", result["port"])
	}

	db := result["database"].(map[string]interface{})
	if db["driver"] != "postgres" {
		t.Errorf("db.driver = %v, want postgres", db["driver"])
	}
	if db["maxConns"].(float64) != 20 {
		t.Errorf("db.maxConns = %v, want 20", db["maxConns"])
	}

	features := result["features"].([]interface{})
	if len(features) != 3 {
		t.Errorf("len(features) = %d, want 3", len(features))
	}
	if features[0] != "schemas" {
		t.Errorf("features[0] = %v, want schemas", features[0])
	}

	rl := result["rateLimit"].(map[string]interface{})
	if rl["burstSize"].(float64) != 10 {
		t.Errorf("rateLimit.burstSize = %v, want 10", rl["burstSize"])
	}
	if rl["windowSize"].(float64) != 0.5 {
		t.Errorf("rateLimit.windowSize = %v, want 0.5", rl["windowSize"])
	}
}

func TestEmptyObject(t *testing.T) {
	input := `{}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("empty object failed: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("len = %d, want 0", len(result))
	}
}

func TestNullValue(t *testing.T) {
	input := `{"val": null}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("null value failed: %v", err)
	}
	if result["val"] != nil {
		t.Errorf("val = %v, want nil", result["val"])
	}
}

func TestCommentBeforeClosingBrace(t *testing.T) {
	input := `{
		"a": 1,
		"b": 2, /* trailing comment after comma */
	}`
	result, err := UnmarshalToMap([]byte(input))
	if err != nil {
		t.Fatalf("comment before closing brace failed: %v", err)
	}
	if result["a"].(float64) != 1 {
		t.Errorf("a = %v, want 1", result["a"])
	}
}
