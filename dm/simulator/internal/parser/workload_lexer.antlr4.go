// Code generated from ../../grammar/Workload.g4 by ANTLR 4.9.3. DO NOT EDIT.

package parser

import (
	"fmt"
	"unicode"

	"github.com/antlr/antlr4/runtime/Go/antlr"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = unicode.IsLetter

var serializedLexerAtn = []uint16{
	3, 24715, 42794, 33075, 47597, 16764, 15335, 30598, 22884, 2, 18, 121,
	8, 1, 4, 2, 9, 2, 4, 3, 9, 3, 4, 4, 9, 4, 4, 5, 9, 5, 4, 6, 9, 6, 4, 7,
	9, 7, 4, 8, 9, 8, 4, 9, 9, 9, 4, 10, 9, 10, 4, 11, 9, 11, 4, 12, 9, 12,
	4, 13, 9, 13, 4, 14, 9, 14, 4, 15, 9, 15, 4, 16, 9, 16, 4, 17, 9, 17, 3,
	2, 3, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4, 3, 4, 3, 5, 3,
	5, 3, 6, 3, 6, 3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 3, 7, 3, 8, 3, 8, 3,
	8, 3, 8, 3, 8, 3, 8, 3, 8, 3, 9, 3, 9, 3, 9, 3, 9, 3, 9, 3, 9, 3, 9, 3,
	10, 3, 10, 3, 10, 3, 10, 3, 10, 3, 10, 3, 10, 3, 10, 3, 10, 3, 10, 3, 10,
	3, 11, 3, 11, 3, 11, 5, 11, 86, 10, 11, 3, 12, 3, 12, 3, 12, 3, 13, 3,
	13, 5, 13, 93, 10, 13, 3, 14, 3, 14, 6, 14, 97, 10, 14, 13, 14, 14, 14,
	98, 3, 14, 3, 14, 3, 15, 6, 15, 104, 10, 15, 13, 15, 14, 15, 105, 3, 16,
	3, 16, 7, 16, 110, 10, 16, 12, 16, 14, 16, 113, 11, 16, 3, 17, 6, 17, 116,
	10, 17, 13, 17, 14, 17, 117, 3, 17, 3, 17, 2, 2, 18, 3, 3, 5, 4, 7, 5,
	9, 6, 11, 7, 13, 8, 15, 9, 17, 10, 19, 11, 21, 12, 23, 13, 25, 14, 27,
	15, 29, 16, 31, 17, 33, 18, 3, 2, 7, 3, 2, 98, 98, 3, 2, 50, 59, 5, 2,
	67, 92, 97, 97, 99, 124, 6, 2, 50, 59, 67, 92, 97, 97, 99, 124, 5, 2, 11,
	12, 15, 15, 34, 34, 2, 126, 2, 3, 3, 2, 2, 2, 2, 5, 3, 2, 2, 2, 2, 7, 3,
	2, 2, 2, 2, 9, 3, 2, 2, 2, 2, 11, 3, 2, 2, 2, 2, 13, 3, 2, 2, 2, 2, 15,
	3, 2, 2, 2, 2, 17, 3, 2, 2, 2, 2, 19, 3, 2, 2, 2, 2, 21, 3, 2, 2, 2, 2,
	23, 3, 2, 2, 2, 2, 25, 3, 2, 2, 2, 2, 27, 3, 2, 2, 2, 2, 29, 3, 2, 2, 2,
	2, 31, 3, 2, 2, 2, 2, 33, 3, 2, 2, 2, 3, 35, 3, 2, 2, 2, 5, 37, 3, 2, 2,
	2, 7, 44, 3, 2, 2, 2, 9, 46, 3, 2, 2, 2, 11, 48, 3, 2, 2, 2, 13, 50, 3,
	2, 2, 2, 15, 57, 3, 2, 2, 2, 17, 64, 3, 2, 2, 2, 19, 71, 3, 2, 2, 2, 21,
	82, 3, 2, 2, 2, 23, 87, 3, 2, 2, 2, 25, 92, 3, 2, 2, 2, 27, 94, 3, 2, 2,
	2, 29, 103, 3, 2, 2, 2, 31, 107, 3, 2, 2, 2, 33, 115, 3, 2, 2, 2, 35, 36,
	7, 61, 2, 2, 36, 4, 3, 2, 2, 2, 37, 38, 7, 84, 2, 2, 38, 39, 7, 71, 2,
	2, 39, 40, 7, 82, 2, 2, 40, 41, 7, 71, 2, 2, 41, 42, 7, 67, 2, 2, 42, 43,
	7, 86, 2, 2, 43, 6, 3, 2, 2, 2, 44, 45, 7, 42, 2, 2, 45, 8, 3, 2, 2, 2,
	46, 47, 7, 43, 2, 2, 47, 10, 3, 2, 2, 2, 48, 49, 7, 63, 2, 2, 49, 12, 3,
	2, 2, 2, 50, 51, 7, 75, 2, 2, 51, 52, 7, 80, 2, 2, 52, 53, 7, 85, 2, 2,
	53, 54, 7, 71, 2, 2, 54, 55, 7, 84, 2, 2, 55, 56, 7, 86, 2, 2, 56, 14,
	3, 2, 2, 2, 57, 58, 7, 87, 2, 2, 58, 59, 7, 82, 2, 2, 59, 60, 7, 70, 2,
	2, 60, 61, 7, 67, 2, 2, 61, 62, 7, 86, 2, 2, 62, 63, 7, 71, 2, 2, 63, 16,
	3, 2, 2, 2, 64, 65, 7, 70, 2, 2, 65, 66, 7, 71, 2, 2, 66, 67, 7, 78, 2,
	2, 67, 68, 7, 71, 2, 2, 68, 69, 7, 86, 2, 2, 69, 70, 7, 71, 2, 2, 70, 18,
	3, 2, 2, 2, 71, 72, 7, 84, 2, 2, 72, 73, 7, 67, 2, 2, 73, 74, 7, 80, 2,
	2, 74, 75, 7, 70, 2, 2, 75, 76, 7, 81, 2, 2, 76, 77, 7, 79, 2, 2, 77, 78,
	7, 47, 2, 2, 78, 79, 7, 70, 2, 2, 79, 80, 7, 79, 2, 2, 80, 81, 7, 78, 2,
	2, 81, 20, 3, 2, 2, 2, 82, 85, 5, 25, 13, 2, 83, 84, 7, 48, 2, 2, 84, 86,
	5, 25, 13, 2, 85, 83, 3, 2, 2, 2, 85, 86, 3, 2, 2, 2, 86, 22, 3, 2, 2,
	2, 87, 88, 7, 66, 2, 2, 88, 89, 5, 31, 16, 2, 89, 24, 3, 2, 2, 2, 90, 93,
	5, 27, 14, 2, 91, 93, 5, 31, 16, 2, 92, 90, 3, 2, 2, 2, 92, 91, 3, 2, 2,
	2, 93, 26, 3, 2, 2, 2, 94, 96, 7, 98, 2, 2, 95, 97, 10, 2, 2, 2, 96, 95,
	3, 2, 2, 2, 97, 98, 3, 2, 2, 2, 98, 96, 3, 2, 2, 2, 98, 99, 3, 2, 2, 2,
	99, 100, 3, 2, 2, 2, 100, 101, 7, 98, 2, 2, 101, 28, 3, 2, 2, 2, 102, 104,
	9, 3, 2, 2, 103, 102, 3, 2, 2, 2, 104, 105, 3, 2, 2, 2, 105, 103, 3, 2,
	2, 2, 105, 106, 3, 2, 2, 2, 106, 30, 3, 2, 2, 2, 107, 111, 9, 4, 2, 2,
	108, 110, 9, 5, 2, 2, 109, 108, 3, 2, 2, 2, 110, 113, 3, 2, 2, 2, 111,
	109, 3, 2, 2, 2, 111, 112, 3, 2, 2, 2, 112, 32, 3, 2, 2, 2, 113, 111, 3,
	2, 2, 2, 114, 116, 9, 6, 2, 2, 115, 114, 3, 2, 2, 2, 116, 117, 3, 2, 2,
	2, 117, 115, 3, 2, 2, 2, 117, 118, 3, 2, 2, 2, 118, 119, 3, 2, 2, 2, 119,
	120, 8, 17, 2, 2, 120, 34, 3, 2, 2, 2, 9, 2, 85, 92, 98, 105, 111, 117,
	3, 8, 2, 2,
}

var lexerChannelNames = []string{
	"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
}

var lexerModeNames = []string{
	"DEFAULT_MODE",
}

var lexerLiteralNames = []string{
	"", "';'", "'REPEAT'", "'('", "')'", "'='", "'INSERT'", "'UPDATE'", "'DELETE'",
	"'RANDOM-DML'",
}

var lexerSymbolicNames = []string{
	"", "", "", "", "", "", "", "", "", "", "TableName", "RowID", "UID", "ReverseQuoteID",
	"Int", "ID", "WS",
}

var lexerRuleNames = []string{
	"T__0", "T__1", "T__2", "T__3", "T__4", "T__5", "T__6", "T__7", "T__8",
	"TableName", "RowID", "UID", "ReverseQuoteID", "Int", "ID", "WS",
}

type WorkloadLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

// NewWorkloadLexer produces a new lexer instance for the optional input antlr.CharStream.
//
// The *WorkloadLexer instance produced may be reused by calling the SetInputStream method.
// The initial lexer configuration is expensive to construct, and the object is not thread-safe;
// however, if used within a Golang sync.Pool, the construction cost amortizes well and the
// objects can be used in a thread-safe manner.
func NewWorkloadLexer(input antlr.CharStream) *WorkloadLexer {
	l := new(WorkloadLexer)
	lexerDeserializer := antlr.NewATNDeserializer(nil)
	lexerAtn := lexerDeserializer.DeserializeFromUInt16(serializedLexerAtn)
	lexerDecisionToDFA := make([]*antlr.DFA, len(lexerAtn.DecisionToState))
	for index, ds := range lexerAtn.DecisionToState {
		lexerDecisionToDFA[index] = antlr.NewDFA(ds, index)
	}
	l.BaseLexer = antlr.NewBaseLexer(input)
	l.Interpreter = antlr.NewLexerATNSimulator(l, lexerAtn, lexerDecisionToDFA, antlr.NewPredictionContextCache())

	l.channelNames = lexerChannelNames
	l.modeNames = lexerModeNames
	l.RuleNames = lexerRuleNames
	l.LiteralNames = lexerLiteralNames
	l.SymbolicNames = lexerSymbolicNames
	l.GrammarFileName = "Workload.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// WorkloadLexer tokens.
const (
	WorkloadLexerT__0           = 1
	WorkloadLexerT__1           = 2
	WorkloadLexerT__2           = 3
	WorkloadLexerT__3           = 4
	WorkloadLexerT__4           = 5
	WorkloadLexerT__5           = 6
	WorkloadLexerT__6           = 7
	WorkloadLexerT__7           = 8
	WorkloadLexerT__8           = 9
	WorkloadLexerTableName      = 10
	WorkloadLexerRowID          = 11
	WorkloadLexerUID            = 12
	WorkloadLexerReverseQuoteID = 13
	WorkloadLexerInt            = 14
	WorkloadLexerID             = 15
	WorkloadLexerWS             = 16
)