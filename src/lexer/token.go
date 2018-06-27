package lexer

import (
	"os"
	"io/ioutil"
	"common"
	"strings"
)

type TokenType int

const (
	Identifier       TokenType = iota + 0xcc
	StringLiteral
	CharacterLiteral
	Integer
	Punctuator
)

var punctuators = "[](){}~?:;,%\\"

type Token struct {
	Type     TokenType
	Str      string
	FileName string
	Line     int
}

func (token *Token) String() string {
	return token.Str
}

func newToken(t TokenType, str string, filename string, line int) *Token {
	return &Token{
		Type:     t,
		Str:      str,
		FileName: filename,
		Line:     line,
	}
}

type TokenList []*Token

func isAlpha(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

type Tokenizer struct {
	tokens   TokenList
	input    []byte
	filename string
	i        int
	line     int
}

func NewTokenizer(filename string) (tokenizer *Tokenizer, err error) {
	if _, err = os.Stat(filename); err != nil {
		return
	}

	var data []byte
	if data, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	tokenizer = &Tokenizer{
		input:    data,
		filename: filename,
		i:        0,
		line:     1,
		tokens:   make(TokenList, 0, 32),
	}
	return
}

func (t *Tokenizer) newToken(typ TokenType, str string) *Token {
	return newToken(typ, str, t.filename, t.line)
}

func (t *Tokenizer) Tokenize() TokenList {
	str := t.input

	for t.i < len(str) {
		if str[t.i] == ' ' || str[t.i] == '\t' {
			t.i++
		} else if str[t.i] == '\n' {
			t.i++
			t.line++
		} else {
			token := t.commonTokenizer()
			if token != nil {
				t.tokens = append(t.tokens, token)
			}
		}
	}

	return t.tokens
}

func (t *Tokenizer) preprocess() {
	str := t.input
	i := t.i

	numOfToken := len(t.tokens)

	for i < len(str) {
		if str[i] == ' ' || str[i] == '\t' {
			i++
		} else if str[i] == '\n' {
			i++
			break
		} else if str[i] == '\\' {
			// if "\\\n" , continue parsing beyond the lines.
			// otherwise, raise Error
			i++
			if i >= len(str) || str[i] != '\n' {
				common.Error("Unexpected char '%c'\n", str[i])
			}
			i++
			if i >= len(str) {
				common.Error("Unexpected EOF")
			}
			t.i = i
			token := t.commonTokenizer()
			if token != nil {
				t.tokens = append(t.tokens, token)
			}
		}
	}

	var directive *Token = nil
	if len(t.tokens) >= numOfToken {
		directive = t.tokens[numOfToken]
	}
	if directive != nil && "include" == t.tokens[numOfToken].Str {
		common.Error("#include not implemented")
	} else {
		arg := "(null)"
		if directive != nil {
			arg = directive.Str
		}
		common.Error("unknow preprocessor directive '%s'", arg)
	}
}

func (t *Tokenizer) commonTokenizer() *Token {
	str := t.input
	i := t.i

	if isAlpha(str[i]) {
		begin := i
		i++
		for i < len(str) && isAlpha(str[i]) || isDigit(str[i]) {
			i++
		}
		t.i = i
		return t.newToken(Identifier, string(str[begin:i]))
	} else if isDigit(str[i]) {
		begin := i
		i++
		for i < len(str) && isDigit(str[i]) {
			i++
		}
		t.i = i
		return t.newToken(Integer, string(str[begin:i]))
	} else if str[i] == '"' || str[i] == '\'' {
		begin := i
		i++
		for i < len(str) && str[i] != str[begin] {
			if str[i] == '\\' {
				i++
			}
			i++
		}

		if i == len(str) {
			common.Error("Unmatched %c at the end of file %s", str[begin], t.filename)
		} else if str[i] != str[begin] {
			common.Error("Expected %c but got char 0x%02X", str[begin], str[i])
		}

		t.i = i + 1

		tokenType := StringLiteral
		if str[begin] == '\'' {
			tokenType = CharacterLiteral
		}
		return t.newToken(tokenType, string(str[begin+1:i]))

	} else if strings.IndexByte(punctuators, str[i]) != -1 {
		t.i = i + 1
		return t.newToken(Punctuator, string(str[i]))
	} else if str[i] == '#' {
		t.i = i + 1
		t.preprocess()
	} else if str[i] == '|' || str[i] == '&' || str[i] == '+' || str[i] == '/' {
		// | || |=
		// & && &=
		// + ++ +=
		// / // /=
		begin := i
		i++
		if i < len(str) && (str[i] == '=' || str[i] == str[begin]) {
			i++
		}
		t.i = i
		return t.newToken(Punctuator, string(str[begin:i]))
	} else if str[i] == '-' {
		// - -- -= ->
		begin := i
		i++
		if i < len(str) && (str[i] == '=' || str[i] == '>') {
			i++
		}
		t.i = i
		return t.newToken(Punctuator, string(str[begin:i]))
	} else if str[i] == '=' || str[i] == '!' || str[i] == '*' {
		// = ==
		// ! !=
		// * *=
		begin := i
		i++
		if i < len(str) && str[i] == '=' {
			i++
		}
		t.i = i
		return t.newToken(Punctuator, string(str[begin:i]))
	} else if str[i] == '<' || str[i] == '>' {
		// < << <= <<=
		// > >> >= >>=
		begin := i
		i++
		if i < len(str) {
			if str[i] == '=' {
				i++
			} else if str[i] == str[begin] {
				i++
				if i < len(str) && str[i] == '=' {
					i++
				}
			}
		}
		t.i = i
		return t.newToken(Punctuator, string(str[begin:i]))
	} else if str[i] == '.' {
		// .
		// ...
		begin := i
		i++
		if i+1 < len(str) && str[i] == '.' && str[i+1] == '.' {
			i += 2
		}
		t.i = i
		return t.newToken(Punctuator, string(str[begin:i]))
	} else {
		common.Error("Unexpected char '%c'\n", str[i])
	}
	return nil
}
