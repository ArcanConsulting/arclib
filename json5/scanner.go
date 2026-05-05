package json5

import (
	"strconv"
	"strings"
)

// normalize converts JSON5 input into valid standard JSON.
// It performs a single pass over the input, handling:
//   - Comment stripping (// and /* */)
//   - Unquoted key quoting
//   - Single-quoted to double-quoted string conversion
//   - Trailing comma removal
//   - Hex number conversion
//   - Infinity/NaN replacement with null
//   - Leading/trailing decimal point normalization
//   - Positive sign removal on numbers
//   - Multi-line string handling (backslash-newline removal)
func normalize(data []byte) []byte {
	s := &scanner{
		src: data,
		out: make([]byte, 0, len(data)),
	}
	s.scan()
	return s.out
}

type scanner struct {
	src []byte
	pos int
	out []byte
}

func (s *scanner) peek() byte {
	if s.pos+1 < len(s.src) {
		return s.src[s.pos+1]
	}
	return 0
}

func (s *scanner) scan() {
	for s.pos < len(s.src) {
		s.scanToken()
	}
}

func (s *scanner) scanToken() {
	ch := s.src[s.pos]

	switch ch {
	case '/':
		s.scanSlash()
	case '"':
		s.scanDoubleQuotedString()
	case '\'':
		s.scanSingleQuotedString()
	case '+':
		s.scanPositive()
	case '.':
		s.scanDot()
	case '-':
		s.scanNegative()
	case ',':
		s.scanComma()
	default:
		s.scanDefault(ch)
	}
}

func (s *scanner) scanSlash() {
	next := s.peek()
	switch next {
	case '/':
		s.skipLineComment()
	case '*':
		s.skipBlockComment()
	default:
		s.out = append(s.out, '/')
		s.pos++
	}
}

func (s *scanner) scanPositive() {
	next := s.peek()
	if isDigit(next) || next == '.' || next == 'I' {
		s.pos++ // skip positive sign
	} else {
		s.out = append(s.out, '+')
		s.pos++
	}
}

func (s *scanner) scanDot() {
	if isDigit(s.peek()) {
		s.scanLeadingDotNumber()
	} else {
		s.out = append(s.out, '.')
		s.pos++
	}
}

func (s *scanner) scanDefault(ch byte) {
	switch {
	case isIdentStart(ch):
		s.scanIdentOrKeyword()
	case ch == '0' && (s.peek() == 'x' || s.peek() == 'X'):
		s.scanHexNumber()
	case isDigit(ch):
		s.scanNumber()
	default:
		s.out = append(s.out, ch)
		s.pos++
	}
}

func (s *scanner) skipLineComment() {
	s.pos += 2 // skip //
	for s.pos < len(s.src) && s.src[s.pos] != '\n' {
		s.pos++
	}
}

func (s *scanner) skipBlockComment() {
	s.pos += 2 // skip /*
	for s.pos+1 < len(s.src) {
		if s.src[s.pos] == '*' && s.src[s.pos+1] == '/' {
			s.pos += 2
			return
		}
		s.pos++
	}
	s.pos = len(s.src)
}

func (s *scanner) scanDoubleQuotedString() {
	s.out = append(s.out, '"')
	s.pos++
	for s.pos < len(s.src) {
		ch := s.src[s.pos]
		if ch == '"' {
			s.out = append(s.out, '"')
			s.pos++
			return
		}
		if ch == '\\' && s.pos+1 < len(s.src) {
			s.handleDoubleQuoteEscape()
			continue
		}
		s.out = append(s.out, ch)
		s.pos++
	}
}

func (s *scanner) handleDoubleQuoteEscape() {
	next := s.src[s.pos+1]
	if next == '\n' {
		s.pos += 2
		return
	}
	if next == '\r' {
		s.pos += 2
		if s.pos < len(s.src) && s.src[s.pos] == '\n' {
			s.pos++
		}
		return
	}
	s.out = append(s.out, '\\', next)
	s.pos += 2
}

func (s *scanner) scanSingleQuotedString() {
	s.out = append(s.out, '"')
	s.pos++ // skip opening '
	for s.pos < len(s.src) {
		ch := s.src[s.pos]
		if ch == '\'' {
			s.out = append(s.out, '"')
			s.pos++
			return
		}
		if ch == '\\' && s.pos+1 < len(s.src) {
			s.handleSingleQuoteEscape()
			continue
		}
		if ch == '"' {
			s.out = append(s.out, '\\', '"')
			s.pos++
			continue
		}
		s.out = append(s.out, ch)
		s.pos++
	}
	s.out = append(s.out, '"')
}

func (s *scanner) handleSingleQuoteEscape() {
	next := s.src[s.pos+1]
	switch next {
	case '\n':
		s.pos += 2
	case '\r':
		s.pos += 2
		if s.pos < len(s.src) && s.src[s.pos] == '\n' {
			s.pos++
		}
	case '\'':
		s.out = append(s.out, '\'')
		s.pos += 2
	case '"':
		s.out = append(s.out, '\\', '"')
		s.pos += 2
	default:
		s.out = append(s.out, '\\', next)
		s.pos += 2
	}
}

func (s *scanner) scanIdentOrKeyword() {
	start := s.pos
	for s.pos < len(s.src) && isIdentContinue(s.src[s.pos]) {
		s.pos++
	}
	word := string(s.src[start:s.pos])

	switch word {
	case "true", "false", "null":
		s.out = append(s.out, word...)
	case "Infinity", "NaN":
		s.out = append(s.out, "null"...)
	default:
		s.emitUnquotedKey(word)
	}
}

func (s *scanner) emitUnquotedKey(word string) {
	if s.isFollowedByColon() {
		s.out = append(s.out, '"')
		s.out = append(s.out, word...)
		s.out = append(s.out, '"')
	} else {
		s.out = append(s.out, word...)
	}
}

func (s *scanner) isFollowedByColon() bool {
	i := s.pos
	for i < len(s.src) && isWhitespace(s.src[i]) {
		i++
	}
	return i < len(s.src) && s.src[i] == ':'
}

func (s *scanner) scanLeadingDotNumber() {
	s.out = append(s.out, '0', '.')
	s.pos++ // skip dot
	for s.pos < len(s.src) && isNumberExpChar(s.src[s.pos]) {
		s.out = append(s.out, s.src[s.pos])
		s.pos++
	}
}

func (s *scanner) scanHexNumber() {
	start := s.pos
	s.pos += 2 // skip 0x
	for s.pos < len(s.src) && isHexDigit(s.src[s.pos]) {
		s.pos++
	}
	hexStr := string(s.src[start:s.pos])
	val, err := strconv.ParseInt(strings.TrimPrefix(strings.TrimPrefix(hexStr, "0x"), "0X"), 16, 64)
	if err != nil {
		s.out = append(s.out, '0')
		return
	}
	s.out = append(s.out, strconv.FormatInt(val, 10)...)
}

func (s *scanner) scanNumber() {
	start := s.pos
	s.consumeDigits()
	if s.pos < len(s.src) && s.src[s.pos] == '.' {
		s.pos++
		s.emitNumberWithDot(start)
	} else {
		s.out = append(s.out, s.src[start:s.pos]...)
	}
	s.consumeExponent()
}

func (s *scanner) emitNumberWithDot(start int) {
	if s.pos >= len(s.src) || !isDigit(s.src[s.pos]) {
		// Trailing dot: 5. -> 5.0
		s.out = append(s.out, s.src[start:s.pos]...)
		s.out = append(s.out, '0')
	} else {
		s.consumeDigits()
		s.out = append(s.out, s.src[start:s.pos]...)
	}
}

func (s *scanner) consumeDigits() {
	for s.pos < len(s.src) && isDigit(s.src[s.pos]) {
		s.pos++
	}
}

func (s *scanner) consumeExponent() {
	if s.pos >= len(s.src) || (s.src[s.pos] != 'e' && s.src[s.pos] != 'E') {
		return
	}
	s.out = append(s.out, s.src[s.pos])
	s.pos++
	if s.pos < len(s.src) && (s.src[s.pos] == '+' || s.src[s.pos] == '-') {
		s.out = append(s.out, s.src[s.pos])
		s.pos++
	}
	for s.pos < len(s.src) && isDigit(s.src[s.pos]) {
		s.out = append(s.out, s.src[s.pos])
		s.pos++
	}
}

func (s *scanner) scanNegative() {
	next := s.peek()
	if next == 'I' {
		// -Infinity -> null
		s.pos++
		s.scanIdentOrKeyword()
		return
	}
	s.out = append(s.out, '-')
	s.pos++
	if next == '.' && s.pos < len(s.src) && isDigit(s.src[s.pos]) {
		s.scanLeadingDotNumber()
	}
}

func (s *scanner) scanComma() {
	saved := s.pos
	s.pos++ // skip comma
	s.skipWhitespaceAndComments()

	if s.pos < len(s.src) && isClosingBracket(s.src[s.pos]) {
		// Trailing comma: omit it, position already past whitespace/comments
		return
	}
	// Not trailing: emit comma and reset to just after it
	s.out = append(s.out, ',')
	s.pos = saved + 1
}

func (s *scanner) skipWhitespaceAndComments() {
	for s.pos < len(s.src) {
		ch := s.src[s.pos]
		if isWhitespace(ch) {
			s.pos++
			continue
		}
		if ch == '/' && s.pos+1 < len(s.src) && s.src[s.pos+1] == '/' {
			s.skipLineComment()
			continue
		}
		if ch == '/' && s.pos+1 < len(s.src) && s.src[s.pos+1] == '*' {
			s.skipBlockComment()
			continue
		}
		return
	}
}

func isClosingBracket(ch byte) bool {
	return ch == '}' || ch == ']'
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '$'
}

func isIdentContinue(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch)
}

func isNumberExpChar(ch byte) bool {
	return isDigit(ch) || ch == 'e' || ch == 'E' || ch == '+' || ch == '-'
}
