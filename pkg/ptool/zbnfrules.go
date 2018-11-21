package ptool

//          = { ' ' | \x0d | \x0a | '\t' | comment }
// digit    = '0'-'9'
// hex      = '0'-'9' | 'a'-'f' | 'A'-'F'
// letter   = 'a'-'z' | 'A'-'Z' | '_'
// eof      = '$'
// m1       = "'"
// m2       = '"'
// noSpace  = '#'
// hexCode  = '\x'#hex#hex
// uniCode  = '\u'#hex#hex#hex#hex#[hex]#[hex]#[hex]#[hex]
// range    = (@hexCode|@uniCode|@string) '-' (@hexCode|@uniCode|@string)
// anyRune  = \u0000-\uffffffff
// string   = m1 # @(!m1 # anyRune # {# !m1 # anyRune }) # m1
//          | m2 # @(!m2 # anyRune # {# !m2 # anyRune }) # m2
// comment  = '//' # {# !\x0d !$ anyRune }
//          | '/*' # {# !'*/' !$ anyRune } '*/'
// term     = @string | @range | @hexCode | @uniCode | @eof
// ident    = letter#{#letter|digit}

// stmt     = @lval '=' orSeq
// lval     = ident
// expr     = @orSeq
// orSeq    = @andSeq { '|' @andSeq }
// andSeq   = @single { [@noSpace] @single }
// single   = @negative | positive
// negative = '!' positive
// positive = @ident | @keep | @term | @curlyBr | @squareBr | @roundBr
// keep     = '@' @single
// squareBr = '[' expr ']'
// curlyBr  = '{' [@noSpace] expr '}'
// roundBr  = '(' expr ')'

//        MINIMAL
//
//          = { ' ' | \x0d | \x0a | '\t' }
// digit    = '0'..'9'
// letter   = 'a'..'z' | 'A'..'Z' | '_'
// eof      = '$'
// noSpace  = '#'
// range    = @string '..' @string
// anyRune  = '\x00'-'\xff'
// content   = '\'' # @string # '\''
// string   = !'\'' # anyRune # {# '\'' # anyRune }
// term     = content | @range | @eof
// ident    = letter#{#letter|digit}

// stmt     = @lval '=' orSeq
// lval     = ident
// expr     = @andSeq | @orSeq | single
// orSeq    = @andSeq '|' @andSeq { '|' @andSeq }
// andSeq   = single [@noSpace] single { [@noSpace] single }
// single   = @negative | positive
// negative = '!' positive
// positive = @ident | @keep | @term | @curlyBr | @squareBr | @roundBr
// keep     = '@' @single
// squareBr = '[' expr ']'
// curlyBr  = '{' [@noSpace] expr '}'
// roundBr  = '(' expr ')'

const (
	cErrorNode = "!!!unknown!!!"
	cEntry     = "entry"
	cDigit     = "digit"
	cHexDigit  = "hexDigit"
	cHex8      = "hex8"
	cEscaped   = "escaped"
	cLetter    = "letter"
	cRange     = "range"
	cIdent     = "ident"
	cNoSpace   = "noSpace"
	cKeep      = "keep"
	// cTerm      = "term"
	cString   = "string"
	cEOF      = "EOF"
	cLVal     = "lval"
	cStmt     = "stmt"
	cAnd      = "andSeq"
	cOr       = "orSeq"
	cStar     = "star"
	cMaybe    = "maybe"
	cNegative = "negative"
	cMaxZBNFEntries
)

func makeZBNFRules() *TBuilder {
	zbnf := NewBuilder()
	pin := newParser()
	zbnf.pin = pin
	fn := func(s string) int {
		if len(pin.items) == 0 {
			pin.items = append(pin.items, tProgItem{name: cErrorNode})
		}
		ret := pin.ByName(s)
		if pin.ByName(s) < 0 {
			pin.items = append(pin.items, tProgItem{name: s})
			ret = len(pin.items) - 1
		}
		return ret
	}

	node := NewNode(fn(cEntry), "",
		// = {space}
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), ""),
			NewNode(fn(cStar), "",
				NewNode(fn(cIdent), "space"),
			),
		),
		// space = ' '|'\x0d'|'\x0a'|'\t'
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), "space"),
			NewNode(fn(cOr), "",
				NewNode(fn(cString), " "),
				NewNode(fn(cString), "\x0d"),
				NewNode(fn(cString), "\x0a"),
				NewNode(fn(cString), "\t"),
			),
		),
		// digit = '0'..'9'
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cDigit),
			NewNode(fn(cRange), "",
				NewNode(fn(cString), "0"),
				NewNode(fn(cString), "9"),
			),
		),
		// letter = 'a'..'z'|'A'..'Z'|'_'
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cLetter),
			NewNode(fn(cOr), "",
				NewNode(fn(cRange), "",
					NewNode(fn(cString), "a"),
					NewNode(fn(cString), "z"),
				),
				NewNode(fn(cRange), "",
					NewNode(fn(cString), "A"),
					NewNode(fn(cString), "Z"),
				),
				NewNode(fn(cString), "_"),
			),
		),
		// hexDigit = '0'..'9'|'a'..'f'|'A'..'F'
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cHexDigit),
			NewNode(fn(cOr), "",
				NewNode(fn(cRange), "",
					NewNode(fn(cString), "0"),
					NewNode(fn(cString), "9"),
				),
				NewNode(fn(cRange), "",
					NewNode(fn(cString), "a"),
					NewNode(fn(cString), "f"),
				),
				NewNode(fn(cRange), "",
					NewNode(fn(cString), "A"),
					NewNode(fn(cString), "F"),
				),
			),
		),
		// hex8 = hexDigit#hexDigit
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cHex8),
			NewNode(fn(cAnd), "",
				NewNode(fn(cIdent), cHexDigit),
				NewNode(fn(cNoSpace), "#"),
				NewNode(fn(cIdent), cHexDigit),
			),
		),
		// EOF = '$'
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cEOF),
			NewNode(fn(cString), "$"),
		),
		// noSpace  = '#'
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cNoSpace),
			NewNode(fn(cString), "#"),
		),
		// range = @singleTerm [ '..' @singleTerm ]
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cRange),
			NewNode(fn(cAnd), "",
				NewNode(fn(cIdent), "singleTerm"),
				// NewNode(fn(cMaybe), "",
				NewNode(fn(cString), ".."),
				NewNode(fn(cIdent), "singleTerm"),
				// ),
			),
		),
		// escaped =  '\x'#@hex8
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cEscaped),
			NewNode(fn(cAnd), "",
				NewNode(fn(cString), "\\x"),
				NewNode(fn(cNoSpace), "#"),
				NewNode(fn(cKeep), "",
					NewNode(fn(cIdent), cHex8),
				),
			),
		),
		// anyRune  = '\x00' ..'\xff'
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), "anyRune"),
			NewNode(fn(cRange), "",
				NewNode(fn(cString), "\x00"),
				NewNode(fn(cString), "\xff"),
			),
		),
		// content   = '\'' # @string # '\''
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), "content"),
			NewNode(fn(cAnd), "",
				NewNode(fn(cString), "'"),
				NewNode(fn(cNoSpace), "#"),
				NewNode(fn(cKeep), "",
					NewNode(fn(cIdent), cString),
				),
				NewNode(fn(cNoSpace), "#"),
				NewNode(fn(cString), "'"),
			),
		),
		// string   = !'\'' # anyRune # {# !'\'' # anyRune }
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cString),
			NewNode(fn(cAnd), "",

				NewNode(fn(cStar), "",
					NewNode(fn(cNoSpace), "#"),
					NewNode(fn(cAnd), "",

						NewNode(fn(cNegative), "",
							NewNode(fn(cString), "'"),
						),
						NewNode(fn(cNoSpace), "#"),
						NewNode(fn(cIdent), "anyRune"),
					),
				),
			),
		),
		// singleTerm  =  content | escaped | @EOF
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), "singleTerm"),
			NewNode(fn(cOr), "",
				NewNode(fn(cIdent), "content"),
				// NewNode(fn(cKeep), "",
				NewNode(fn(cIdent), cEscaped),
				// ),
				NewNode(fn(cKeep), "",
					NewNode(fn(cIdent), cEOF),
				),
			),
		),
		// term  =  @range | singleTerm
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), "term"),
			NewNode(fn(cOr), "",
				NewNode(fn(cKeep), "",
					NewNode(fn(cIdent), cRange),
				),
				NewNode(fn(cIdent), "singleTerm"),
				// NewNode(fn(cIdent), cEscaped),
				// NewNode(fn(cKeep), "",
				// 	NewNode(fn(cIdent), cEOF),
				// ),
			),
		),
		// ident = letter#{#letter|digit} | ','
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cIdent),
			NewNode(fn(cOr), "",
				NewNode(fn(cAnd), "",
					NewNode(fn(cIdent), cLetter),
					NewNode(fn(cNoSpace), "#"),
					NewNode(fn(cStar), "",
						NewNode(fn(cNoSpace), "#"),
						NewNode(fn(cOr), "",
							NewNode(fn(cIdent), cLetter),
							NewNode(fn(cIdent), cDigit),
						),
					),
				),
				NewNode(fn(cString), ","),
			),
		),
		// separator = {#!'\x0d'#space}#('\x0d'|';')
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), "separator"),
			NewNode(fn(cAnd), "",
				NewNode(fn(cStar), "",
					NewNode(fn(cNoSpace), "#"),
					NewNode(fn(cAnd), "",
						NewNode(fn(cNegative), "",
							NewNode(fn(cString), "\x0d"),
						),
						NewNode(fn(cNoSpace), "#"),
						NewNode(fn(cIdent), "space"),
					),
				),
				NewNode(fn(cNoSpace), "#"),
				NewNode(fn(cOr), "",
					NewNode(fn(cString), "\x0d"),
					NewNode(fn(cString), ";"),
				),
				NewNode(fn(cEOF), "$"),
			),
		),
		// entry = '' @stmt#{#separator [@stmt] } $
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cEntry),
			NewNode(fn(cAnd), "",
				NewNode(fn(cString), ""),
				NewNode(fn(cMaybe), "",
					NewNode(fn(cAnd), "",
						NewNode(fn(cKeep), "",
							NewNode(fn(cIdent), cStmt),
						),
						NewNode(fn(cStar), "",
							NewNode(fn(cAnd), "",
								NewNode(fn(cString), ";"),
								NewNode(fn(cMaybe), "",
									NewNode(fn(cKeep), "",
										NewNode(fn(cIdent), cStmt),
									),
								),
							),
						),
					),
				),
				NewNode(fn(cEOF), cEOF),
			),
		),
		// lval = [ident]
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cLVal),
			NewNode(fn(cAnd), "",
				NewNode(fn(cMaybe), "",
					NewNode(fn(cIdent), cIdent),
				),
			),
		),
		// stmt = @lval '=' expr
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cStmt),
			NewNode(fn(cAnd), "",
				NewNode(fn(cKeep), "",
					NewNode(fn(cIdent), cLVal),
				),
				NewNode(fn(cString), "="),
				NewNode(fn(cIdent), "expr"),
			),
		),
		// expr = @orSeq | @andSeq | single
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), "expr"),
			NewNode(fn(cOr), "",
				NewNode(fn(cKeep), "",
					NewNode(fn(cIdent), cOr),
				),
				NewNode(fn(cKeep), "",
					NewNode(fn(cIdent), cAnd),
				),
				NewNode(fn(cIdent), "single"),
			),
		),
		// orSeq    = (@andSeq|single) '|' (@andSeq|single) { '|' (@andSeq|single) }
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cOr),
			NewNode(fn(cAnd), "",
				NewNode(fn(cOr), "",
					NewNode(fn(cKeep), "",
						NewNode(fn(cIdent), cAnd),
					),
					NewNode(fn(cIdent), "single"),
				),
				NewNode(fn(cString), "|"),
				NewNode(fn(cOr), "",
					NewNode(fn(cKeep), "",
						NewNode(fn(cIdent), cAnd),
					),
					NewNode(fn(cIdent), "single"),
				),
				NewNode(fn(cStar), "",
					NewNode(fn(cAnd), "",
						NewNode(fn(cString), "|"),
						NewNode(fn(cOr), "",
							NewNode(fn(cKeep), "",
								NewNode(fn(cIdent), cAnd),
							),
							NewNode(fn(cIdent), "single"),
						),
					),
				),
			),
		),
		// andSeq   = single [@noSpace] single { [@noSpace] single }
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cAnd),
			NewNode(fn(cAnd), "",
				NewNode(fn(cIdent), "single"),
				NewNode(fn(cMaybe), "",
					NewNode(fn(cKeep), "",
						NewNode(fn(cIdent), cNoSpace),
					),
				),
				NewNode(fn(cIdent), "single"),
				NewNode(fn(cStar), "",
					NewNode(fn(cAnd), "",
						NewNode(fn(cMaybe), "",
							NewNode(fn(cKeep), "",
								NewNode(fn(cIdent), cNoSpace),
							),
						),
						NewNode(fn(cIdent), "single"),
					),
				),
			),
		),
		// single   = @negative | positive
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), "single"),
			NewNode(fn(cOr), "",
				NewNode(fn(cKeep), "",
					NewNode(fn(cIdent), cNegative),
				),
				NewNode(fn(cIdent), "positive"),
			),
		),
		// negative = '!' positive
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cNegative),
			NewNode(fn(cAnd), "",
				NewNode(fn(cString), "!"),
				NewNode(fn(cIdent), "positive"),
			),
		),
		// positive = @ident | @keep | @term | @curlyBr | @squareBr | roundBr
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), "positive"),
			NewNode(fn(cOr), "",
				NewNode(fn(cKeep), "",
					NewNode(fn(cIdent), cIdent),
				),
				NewNode(fn(cKeep), "",
					NewNode(fn(cIdent), cKeep),
				),
				NewNode(fn(cIdent), "term"),
				NewNode(fn(cKeep), "",
					NewNode(fn(cIdent), cStar),
				),
				NewNode(fn(cKeep), "",
					NewNode(fn(cIdent), cMaybe),
				),
				NewNode(fn(cIdent), "roundBr"),
			),
		),
		// keep = '@' single
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cKeep),
			NewNode(fn(cAnd), "",
				NewNode(fn(cString), "@"),
				NewNode(fn(cIdent), "single"),
			),
		),
		// squareBr = '[' expr ']'
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cMaybe),
			NewNode(fn(cAnd), "",
				NewNode(fn(cString), "["),
				NewNode(fn(cIdent), "expr"),
				NewNode(fn(cString), "]"),
			),
		),
		// curlyBr  = '{' [@noSpace] expr '}'
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), cStar),
			NewNode(fn(cAnd), "",
				NewNode(fn(cString), "{"),
				NewNode(fn(cMaybe), "",
					NewNode(fn(cKeep), "",
						NewNode(fn(cIdent), cNoSpace),
					),
				),
				NewNode(fn(cIdent), "expr"),
				NewNode(fn(cString), "}"),
			),
		),
		// roundBr  = '(' expr ')'
		NewNode(fn(cStmt), "",
			NewNode(fn(cIdent), "roundBr"),
			NewNode(fn(cAnd), "",
				NewNode(fn(cString), "("),
				NewNode(fn(cIdent), "expr"),
				NewNode(fn(cString), ")"),
			),
		),
	)
	zbnf.tree = node
	return zbnf
}
