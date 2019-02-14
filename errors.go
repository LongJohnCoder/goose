package goose

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"runtime"
	"strings"
)

// errorReporter groups methods for reporting errors, documenting what kind of
// issue was encountered in a uniform way.
type errorReporter struct {
	fset *token.FileSet
}

func newErrorReporter(fset *token.FileSet) errorReporter {
	return errorReporter{fset}
}

func (r errorReporter) printGo(n ast.Node) string {
	var what bytes.Buffer
	printer.Fprint(&what, r.fset, n)
	return string(what.Bytes())
}

func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(1 + skip)
	if !ok {
		return "<no caller>"
	}

	return fmt.Sprintf("%s:%d", file, line)
}

type gooseError struct{ err *ConversionError }

// A ConversionError reports detailed information on an error producing Coq code.
//
// Errors include a category describing the severity of the error.
//
// The category "unsupported" is the only error that should result from normal usage,
// when attempting to use a feature goose intentionally does not support.
//
// "todo" and "future" are markers for code that could be supported but is not currently
// handled.
//
// The categories "impossible(go)" and "impossible(no-examples)" indicate a bug in goose
// (at the very least these cases should be checked and result in an unsupported error)
type ConversionError struct {
	Category string
	// the main description of what went wrong
	Message string
	// the snippet in the source program responsible for the error
	GoCode string
	// (for internal debugging) file:lineno for the goose code that threw the error
	GooseCaller string
	// file:lineno for the source program where GoCode appears
	GoSrcFile string
}

func (e *ConversionError) Error() string {
	lines := []string{
		fmt.Sprintf("[%s]: %s", e.Category, e.Message),
		fmt.Sprintf("%s", e.GoCode),
		fmt.Sprintf("  %s", e.GooseCaller),
		fmt.Sprintf("  src: %s", e.GoSrcFile),
	}
	return strings.Join(lines, "\n")
}

func (r errorReporter) prefixed(prefix string, n ast.Node, msg string, args ...interface{}) {
	where := r.fset.Position(n.Pos())
	what := r.printGo(n)
	formatted := fmt.Sprintf(msg, args...)

	err := &ConversionError{
		Category:    prefix,
		Message:     formatted,
		GoCode:      what,
		GooseCaller: getCaller(2),
		GoSrcFile:   where.String()}

	panic(gooseError{err: err})
}

// nope reports a situation that I thought was impossible from reading the documentation.
func (r errorReporter) nope(n ast.Node, msg string, args ...interface{}) {
	r.prefixed("impossible(go)", n, msg, args...)
}

// noExample reports a situation I thought was impossible because I couldn't
// think of how to do it in Go.
func (r errorReporter) noExample(n ast.Node, msg string, args ...interface{}) {
	r.prefixed("impossible(no-examples)", n, msg, args...)
}

// futureWork reports something we could theoretically handle but probably
// won't.
func (r errorReporter) futureWork(n ast.Node, msg string, args ...interface{}) {
	r.prefixed("future", n, msg, args...)
}

// todo reports a situation that is intended to be handled but we haven't gotten
// around to.
func (r errorReporter) todo(n ast.Node, msg string, args ...interface{}) {
	r.prefixed("todo", n, msg, args...)
}

// unsupported reports something intentionally unhandled (the code should not use
// this feature).
func (r errorReporter) unsupported(n ast.Node, msg string, args ...interface{}) {
	r.prefixed("unsupported", n, msg, args...)
}
