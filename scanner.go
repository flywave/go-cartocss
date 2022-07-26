package cartocss

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

type tokenType int

func (t tokenType) String() string {
	return tokenNames[t]
}

type token struct {
	t      tokenType
	value  string
	line   int
	column int
}

func (t *token) String() string {
	if len(t.value) > 10 {
		return fmt.Sprintf("%s (line: %d, column: %d): %.10q...",
			t.t, t.line, t.column, t.value)
	}
	return fmt.Sprintf("%s (line: %d, column: %d): %q",
		t.t, t.line, t.column, t.value)
}

const (
	tokenError tokenType = iota
	tokenEOF
	// regular tokens
	tokenIdent
	tokenAtKeyword
	tokenString
	tokenHash
	tokenAttachment
	tokenClass
	tokenInstance
	tokenLBrace
	tokenRBrace
	tokenLBracket
	tokenRBracket
	tokenLParen
	tokenRParen
	tokenColon
	tokenSemicolon
	tokenComma
	tokenPlus
	tokenMinus
	tokenMultiply
	tokenDivide
	tokenComp
	tokenNumber
	tokenPercentage
	tokenDimension
	tokenURI
	tokenUnicodeRange
	tokenS
	tokenComment
	tokenFunction
	tokenIncludes
	tokenDashMatch
	tokenPrefixMatch
	tokenSuffixMatch
	tokenSubstringMatch
	tokenChar
	tokenBOM
)

var tokenNames = map[tokenType]string{
	tokenError:          "error",
	tokenEOF:            "EOF",
	tokenIdent:          "IDENT",
	tokenAtKeyword:      "ATKEYWORD",
	tokenString:         "STRING",
	tokenHash:           "HASH",
	tokenAttachment:     "ATTACHMENT",
	tokenClass:          "CLASS",
	tokenInstance:       "INSTANCE",
	tokenLBrace:         "LBRACE",
	tokenRBrace:         "RBRACE",
	tokenLBracket:       "LBRACKET",
	tokenRBracket:       "RBRACKET",
	tokenLParen:         "LPAREN",
	tokenRParen:         "RPAREN",
	tokenColon:          "COLON",
	tokenSemicolon:      "SEMICOLON",
	tokenComma:          "COMMA",
	tokenPlus:           "PLUS",
	tokenMinus:          "MINUS",
	tokenMultiply:       "MULTIPLY",
	tokenDivide:         "DIVIDE",
	tokenComp:           "COMP",
	tokenNumber:         "NUMBER",
	tokenPercentage:     "PERCENTAGE",
	tokenDimension:      "DIMENSION",
	tokenURI:            "URI",
	tokenUnicodeRange:   "UNICODE-RANGE",
	tokenS:              "S",
	tokenComment:        "COMMENT",
	tokenFunction:       "FUNCTION",
	tokenIncludes:       "INCLUDES",
	tokenDashMatch:      "DASHMATCH",
	tokenPrefixMatch:    "PREFIXMATCH",
	tokenSuffixMatch:    "SUFFIXMATCH",
	tokenSubstringMatch: "SUBSTRINGMATCH",
	tokenChar:           "CHAR",
	tokenBOM:            "BOM",
}

var macroRegexp = regexp.MustCompile(`\{[a-z]+\}`)

var macros = map[string]string{
	"ident":      `-?{nmstart}{nmchar}*`,
	"name":       `{nmchar}+`,
	"nmstart":    `[a-zA-Z_]|{nonascii}|{escape}`,
	"nonascii":   "[\u0080-\uD7FF\uE000-\uFFFD\U00010000-\U0010FFFF]",
	"unicode":    `\\[0-9a-fA-F]{1,6}{wc}?`,
	"escape":     "{unicode}|\\\\[\u0020-\u007E\u0080-\uD7FF\uE000-\uFFFD\U00010000-\U0010FFFF]",
	"nmchar":     `[a-zA-Z0-9_-]|{nonascii}|{escape}`,
	"num":        `-?[0-9]*\.?[0-9]+`,
	"string":     `"(?:{stringchar}|')*"|'(?:{stringchar}|")*'`,
	"stringchar": `{urlchar}|[ ]|\\{nl}`,
	"urlchar":    "[\u0009\u0021\u0023-\u0026\u0028-\u007E]|{nonascii}|{escape}",
	"nl":         `[\n\r\f]|\r\n`,
	"w":          `{wc}*`,
	"wc":         `[\t\n\f\r ]`,
}

var productions = map[tokenType]string{
	tokenIdent:        `{ident}`,
	tokenAtKeyword:    `@{ident}`,
	tokenString:       `{string}`,
	tokenHash:         `#{name}`,
	tokenAttachment:   `::{name}`,
	tokenClass:        `\.{name}`,
	tokenInstance:     `{ident}/`,
	tokenNumber:       `{num}`,
	tokenPercentage:   `{num}%`,
	tokenDimension:    `{num}{ident}`,
	tokenURI:          `url\({w}(?:{string}|{urlchar}*){w}\)`,
	tokenUnicodeRange: `U\+[0-9A-F\?]{1,6}(?:-[0-9A-F]{1,6})?`,
	tokenS:            `{wc}+`,
	tokenComment:      `/\*[^\*]*[\*]+(?:[^/][^\*]*[\*]+)*/`,
	tokenFunction:     `{ident}\(`,
	tokenComp:         `>=|<=|>|<|!=|=~|=`,
	//tokenIncludes:       `~=`,
	//tokenDashMatch:      `\|=`,
	//tokenPrefixMatch:    `\^=`,
	//tokenSuffixMatch:    `\$=`,
	//tokenSubstringMatch: `\*=`,
	//tokenChar:           `[^"']`,
	//tokenBOM:            "\uFEFF",
}

var matchers = map[tokenType]*regexp.Regexp{}

var matchOrder = []tokenType{
	tokenURI,
	tokenFunction,
	tokenUnicodeRange,
	tokenInstance,
	tokenIdent,
	tokenDimension,
	tokenPercentage,
	tokenNumber,
	tokenComp,
}

func init() {
	replaceMacro := func(s string) string {
		return "(?:" + macros[s[1:len(s)-1]] + ")"
	}
	for t, s := range productions {
		for macroRegexp.MatchString(s) {
			s = macroRegexp.ReplaceAllStringFunc(s, replaceMacro)
		}
		matchers[t] = regexp.MustCompile("^(?:" + s + ")")
	}
}

func ScannerIds(css string) []string {
	scan := newScanner(css)
	var ids []string
	for {
		t := scan.Next()
		if t.t == tokenEOF {
			break
		}
		if t.t == tokenHash {
			ids = append(ids, string([]byte(t.value)[1:]))
		}
	}
	return ids
}

func newScanner(input string) *scanner {
	input = strings.Replace(input, "\r\n", "\n", -1)
	return &scanner{
		input: input,
		row:   1,
		col:   1,
	}
}

type scanner struct {
	input string
	pos   int
	row   int
	col   int
	err   *token
}

func (s *scanner) Next() *token {
	if s.err != nil {
		return s.err
	}
	if s.pos >= len(s.input) {
		s.err = &token{tokenEOF, "", s.row, s.col}
		return s.err
	}
	if s.pos == 0 {
		if strings.HasPrefix(s.input, "\uFEFF") {
			return s.emitSimple(tokenBOM, "\uFEFF")
		}
	}
	input := s.input[s.pos:]
	switch input[0] {
	case '\t', '\n', '\f', '\r', ' ':
		return s.emitToken(tokenS, matchers[tokenS].FindString(input))
	case '.':
		if len(input) > 1 && !unicode.IsDigit(rune(input[1])) {
			if match := matchers[tokenClass].FindString(input); match != "" {
				return s.emitSimple(tokenClass, match)
			}
			return s.emitSimple(tokenChar, ".")
		}
	case '#':
		if match := matchers[tokenHash].FindString(input); match != "" {
			return s.emitSimple(tokenHash, match)
		}
		return s.emitSimple(tokenChar, "#")
	case '@':
		if match := matchers[tokenAtKeyword].FindString(input); match != "" {
			return s.emitSimple(tokenAtKeyword, match)
		}
		return s.emitSimple(tokenChar, "@")
	case ':':
		if match := matchers[tokenAttachment].FindString(input); match != "" {
			return s.emitSimple(tokenAttachment, match)
		}
		return s.emitSimple(tokenColon, ":")
	case '%', '&':
		return s.emitSimple(tokenChar, string(input[0]))
	case ',':
		return s.emitSimple(tokenComma, string(input[0]))
	case ';':
		return s.emitSimple(tokenSemicolon, string(input[0]))
	case '(':
		return s.emitSimple(tokenLParen, string(input[0]))
	case ')':
		return s.emitSimple(tokenRParen, string(input[0]))
	case '[':
		return s.emitSimple(tokenLBracket, string(input[0]))
	case ']':
		return s.emitSimple(tokenRBracket, string(input[0]))
	case '{':
		return s.emitSimple(tokenLBrace, string(input[0]))
	case '}':
		return s.emitSimple(tokenRBrace, string(input[0]))
	case '+':
		return s.emitSimple(tokenPlus, string(input[0]))
	case '-':
		if match := matchers[tokenNumber].FindString(input); match != "" {
			return s.emitSimple(tokenNumber, match)
		}
		if match := matchers[tokenFunction].FindString(input); match != "" {
			return s.emitSimple(tokenFunction, match)
		}

		return s.emitSimple(tokenMinus, string(input[0]))
	case '*':
		return s.emitSimple(tokenMultiply, string(input[0]))
	case '"', '\'':
		match := matchers[tokenString].FindString(input)
		if match != "" {
			return s.emitToken(tokenString, match)
		} else {
			s.err = &token{tokenError, "unclosed quotation mark", s.row, s.col}
			return s.err
		}
	case '/':
		if len(input) > 1 && input[1] == '*' {
			match := matchers[tokenComment].FindString(input)
			if match != "" {
				return s.emitToken(tokenComment, match)
			} else {
				s.err = &token{tokenError, "unclosed comment", s.row, s.col}
				return s.err
			}
		} else if len(input) > 1 && input[1] == '/' {
			idx := strings.Index(input, "\n")
			if idx < 0 {
				idx = len(input)
			}
			return s.emitToken(tokenComment, input[:idx])
		}
		return s.emitSimple(tokenDivide, "/")
	}
	for _, token := range matchOrder {
		if match := matchers[token].FindString(input); match != "" {
			return s.emitToken(token, match)
		}
	}
	r, width := utf8.DecodeRuneInString(input)
	token := &token{tokenChar, string(r), s.row, s.col}
	s.col += width
	s.pos += width
	return token
}

func (s *scanner) updatePosition(text string) {
	width := utf8.RuneCountInString(text)
	lines := strings.Count(text, "\n")
	s.row += lines
	if lines == 0 {
		s.col += width
	} else {
		s.col = utf8.RuneCountInString(text[strings.LastIndex(text, "\n"):])
	}
	s.pos += len(text)
}

func (s *scanner) emitToken(t tokenType, v string) *token {
	token := &token{t, v, s.row, s.col}
	s.updatePosition(v)
	return token
}

func (s *scanner) emitSimple(t tokenType, v string) *token {
	token := &token{t, v, s.row, s.col}
	s.col += len(v)
	s.pos += len(v)
	return token
}
