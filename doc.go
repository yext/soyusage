// Package soyusage analyzes soy templates compiled using github.com/yext/soy to
// determine the usage of parameters from within the template and other templates
// it calls.
//
// The AST for a template is walked, and a tree of parameters is constructed
// defining the root parameters and sub-fields of these parameters, along with
// where and how they are used.
package soyusage
