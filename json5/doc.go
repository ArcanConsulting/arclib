// Package json5 provides a JSON5 parser that extends standard JSON with
// developer-friendly syntax.
//
// Supported JSON5 features beyond standard JSON:
//
//   - Single-line comments (//)
//   - Multi-line comments (/* */)
//   - Trailing commas in objects and arrays
//   - Unquoted object keys (identifiers matching [a-zA-Z_$][a-zA-Z0-9_$]*)
//   - Single-quoted strings
//   - Multi-line strings (escaped newlines with backslash)
//   - Hexadecimal numbers (0x...)
//   - Infinity and NaN (converted to null)
//   - Leading/trailing decimal point (.5, 5.)
//   - Positive sign on numbers (+1)
//
// The implementation preprocesses JSON5 input into valid JSON, then delegates
// to encoding/json for actual parsing. This keeps the implementation simple
// and leverages the battle-tested standard library for struct mapping.
package json5
