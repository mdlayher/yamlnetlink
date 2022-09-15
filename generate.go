package yamlnetlink

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Config specifies configuration for Generate.
type Config struct {
	// Package specifies an optional package name for the generated code. If
	// unset, the default is to use the Spec.Name field.
	Package string
}

// Generate generates formatted Go code from a YAML netlink Spec. If cfg is nil,
// a default Config is used.
func Generate(s *Spec, cfg *Config) ([]byte, error) {
	if cfg == nil {
		cfg = &Config{
			Package: s.Name,
		}
	}

	var b bytes.Buffer
	g := newGenerator(s, &b)

	g.header(cfg.Package)
	g.conn()

	for _, op := range s.Operations.List {
		g.op(op)
	}

	return format.Source(b.Bytes())
}

// A generator generates code from a Spec and writes it to w.
type generator struct {
	s *Spec
	w io.Writer

	// An index of attribute set names to AttributeSets.
	asIndex map[string]AttributeSet
}

// newGenerator creates a generator which outputs to w.
func newGenerator(s *Spec, w io.Writer) *generator {
	asIndex := make(map[string]AttributeSet)
	for _, as := range s.AttributeSets {
		asIndex[as.Name] = as
	}

	return &generator{
		s:       s,
		w:       w,
		asIndex: asIndex,
	}
}

// header writes the Go package and import headers for a given package name.
func (g *generator) header(pkg string) {
	g.pf("// Package %s is generated from a YAML netlink specification for family %q.", pkg, g.s.Name)
	g.pf("//")
	g.pf("// Description: %s", g.s.Description)
	g.pf("//")
	g.pf("// Code generated by yamlnetlink-go. DO NOT EDIT.")
	g.pf("package %s", pkg)
	g.pf("")

	g.pf("import (")
	g.pf(`	"errors"`)
	g.pf("")
	g.pf(`	"github.com/mdlayher/genetlink"`)
	g.pf(`	"github.com/mdlayher/netlink"`)
	g.pf(`	"golang.org/x/sys/unix"`)
	g.pf(")")
	g.pf("")
}

// conn generates a Conn type for a netlink family.
func (g *generator) conn() {
	g.pf("// A Conn is a connection to netlink family %q.", g.s.Name)
	g.pf("type Conn struct {")
	g.pf("	c *genetlink.Conn")
	g.pf("}")
	g.pf("")

	g.pf("// Dial opens a Conn for netlink family %q. Any options are passed directly", g.s.Name)
	g.pf("// to the underlying netlink package.")
	g.pf("func Dial(cfg *netlink.Config) (*Conn, error) {")
	g.pf("	c, err := genetlink.Dial(cfg)")
	g.pf("	if err != nil {")
	g.pf("		return nil, err")
	g.pf("	}")
	g.pf("")
	g.pf("	return &Conn{c: c}, nil")
	g.pf("}")
	g.pf("")

	g.pf("// Close closes the Conn's underlying netlink connection.")
	g.pf("func (c *Conn) Close() error { return c.c.Close() }")
	g.pf("")
}

// op begins generating code for the input Operation.
func (g *generator) op(op Operation) {
	// Only generate operations where either the request or response has at
	// least one attribute.
	if len(op.Do.Request.Attributes) > 0 || len(op.Do.Reply.Attributes) > 0 {
		g.structs(op, doOp, doRequest)
		g.structs(op, doOp, doReply)
		g.method(op, doOp)
	}

	if len(op.Dump.Request.Attributes) > 0 || len(op.Dump.Reply.Attributes) > 0 {
		g.structs(op, dumpOp, doRequest)
		g.structs(op, dumpOp, doReply)
		g.method(op, dumpOp)
	}
}

// method generates a Do or Dump method for an Operation.
func (g *generator) method(op Operation, dod doOrDump) {
	var (
		slice, flags string
		oas          OperationAttributes
	)

	// Do and Dump operations have different flags, attributes, and return
	// types.
	switch dod {
	case doOp:
		flags = "netlink.Request"
		oas = op.Do
	case dumpOp:
		slice = "[]"
		flags = "netlink.Request|netlink.Dump"
		oas = op.Dump
	}

	{
		s := dod.String() + title(op.Name)

		g.pf("// %s wraps the %q operation:", s, op.Name)
		g.pf("// %s", op.Description)

		// If there are no attributes, generate no parameter names.
		var params string
		if len(oas.Request.Attributes) > 0 {
			params = fmt.Sprintf("req %sRequest", s)
		}

		g.pf("func (c *Conn) %s(%s) (%s*%sReply, error) {", s, params, slice, s)
	}

	// Generate the attribute encoder for arguments.
	g.encoder(op, oas.Request)

	// Use packed arguments in a genetlink message body to execute a command.
	//
	// TODO(mdlayher): do all families let you omit a version number here?
	g.pf("msg := genetlink.Message{")
	g.pf("	Header: genetlink.Header{")
	g.pf("		Command: %s,", unixConst(g.s.Operations.NamePrefix+op.Name))
	g.pf("	},")
	g.pf("	Data: b,")
	g.pf("}")
	g.pf("")

	// TODO: where does ID come from?
	g.pf("msgs, err := c.c.Execute(msg, unix.GENL_ID_CTRL, %s)", flags)
	g.pf("if err != nil {")
	g.pf("	return nil, err")
	g.pf("}")
	g.pf("")

	// Generate an attribute decoder for outputs.
	g.decoder(op, dod)

	g.pf("}")
	g.pf("")
}

// structs generates a struct for an Operation. Different attribute sets are
// used depending on the values for Do/Dump and Request/Reply.
func (g *generator) structs(
	op Operation,
	dod doOrDump,
	ror requestOrReply,
) {
	// Narrow down which list of attributes we'll be generating from.
	var list OperationAttributesList
	{
		var oas OperationAttributes
		switch dod {
		case doOp:
			oas = op.Do
		case dumpOp:
			oas = op.Dump
		}

		switch ror {
		case doRequest:
			list = oas.Request
		case doReply:
			list = oas.Reply
		}
	}

	var (
		opName   = dod.String() + title(op.Name)
		fullName = opName + ror.String()
	)

	if len(list.Attributes) == 0 {
		// The chosen list has no attributes, don't generate a struct.
		return
	}

	g.pf("// %s is used with the %s method.", fullName, opName)
	g.pf("type %s struct {", fullName)
	for _, a := range g.attrs(op.AttributeSet, list.Attributes) {
		if a.Description != "" {
			g.pf("// %s", a.Description)
		}

		// Generate the actual fields.
		switch f := camelCase(a.Name); a.Type {
		case "u8":
			g.pf("%s uint8", f)
		case "u16":
			g.pf("%s uint16", f)
		case "u32":
			g.pf("%s uint32", f)
		case "u64":
			g.pf("%s uint64", f)
		case "nul-string":
			g.pf("%s string", f)
		default:
			g.pf("// TODO: field %q, type %q", f, a.Type)
		}
	}

	g.pf("}")
	g.pf("")
}

// encoder generates a netlink attribute encoder for a set of attribute
// arguments for a command.
func (g *generator) encoder(op Operation, list OperationAttributesList) {
	if len(list.Attributes) == 0 {
		// Shortcut.
		g.pf("// No attribute arguments.")
		g.pf("var b []byte")
		g.pf("")
		return
	}

	g.pf("ae := netlink.NewAttributeEncoder()")
	for _, a := range g.attrs(op.AttributeSet, list.Attributes) {
		// Use the unix package const for each type, and field to fill in the
		// arguments that are non-zero.
		var (
			typ = unixConst(g.asIndex[op.AttributeSet].NamePrefix + a.Name)
			f   = "req." + camelCase(a.Name)
		)

		// mkUint generates a uint* case.
		mkUint := func(bits int) {
			g.pf("if %s != 0 {", f)
			g.pf("	ae.Uint%d(%s, %s)", bits, typ, f)
			g.pf("}")
		}

		switch a.Type {
		case "u8":
			mkUint(8)
		case "u16":
			mkUint(16)
		case "u32":
			mkUint(32)
		case "u64":
			mkUint(64)
		case "nul-string":
			g.pf(`if %s != "" {`, f)
			g.pf("	ae.String(%s, %s)", typ, f)
			g.pf("}")
		default:
			g.pf("	// TODO: field %q, type %q", f, a.Type)
		}
	}

	// Finally pack the attributes.
	g.pf("")
	g.pf("b, err := ae.Encode()")
	g.pf("if err != nil {")
	g.pf("	return nil, err")
	g.pf("}")
	g.pf("")
}

// decoder generates a netlink attribute decoder loop to to iterate over reply
// messages from a Do or Dump.
func (g *generator) decoder(op Operation, dod doOrDump) {
	name := dod.String() + title(op.Name) + doReply.String()

	// Preallocate replies and range over all inputs, decoding each.
	g.pf("replies := make([]*%s, 0, len(msgs))", name)
	g.pf("for _, m := range msgs {")

	g.pf("ad, err := netlink.NewAttributeDecoder(m.Data)")
	g.pf("if err != nil {")
	g.pf("	return nil, err")
	g.pf("}")
	g.pf("")
	g.pf("var reply %s", name)
	g.pf("for ad.Next() {")
	g.pf("	switch ad.Type() {")

	var oas OperationAttributesList
	switch dod {
	case doOp:
		oas = op.Do.Reply
	case dumpOp:
		oas = op.Dump.Reply
	}

	// Begin generating switch cases.
	for _, a := range g.attrs(op.AttributeSet, oas.Attributes) {
		// Use the unix package const for each type, and field to fill in the
		// arguments that are non-zero.
		var (
			typ = unixConst(g.asIndex[op.AttributeSet].NamePrefix + a.Name)
			f   = "reply." + camelCase(a.Name)
		)

		// mkUint generates a uint* case.
		mkUint := func(bits int) { g.pf("%s = ad.Uint%d()", f, bits) }

		g.pf("case %s:", typ)
		switch a.Type {
		case "u8":
			mkUint(8)
		case "u16":
			mkUint(16)
		case "u32":
			mkUint(32)
		case "u64":
			mkUint(64)
		case "nul-string":
			g.pf("%s = ad.String()", f)
		default:
			g.pf("	// TODO: field %q, type %q", f, a.Type)
		}
	}

	g.pf("	}")
	g.pf("}")
	g.pf("")

	// Make sure to check for decoder errors for reach reply.
	g.pf("if err := ad.Err(); err != nil {")
	g.pf("	return nil, err")
	g.pf("}")
	g.pf("")
	g.pf("replies = append(replies, &reply)")
	g.pf("}")
	g.pf("")

	// Do returns a single reply, Dump returns all.
	switch dod {
	case doOp:
		g.pf("if len(replies) != 1 {")
		g.pf(`	return nil, errors.New("%s: expected exactly one %s")`, g.s.Name, name)
		g.pf("}")
		g.pf("")

		g.pf("return replies[0], nil")
	case dumpOp:
		g.pf("return replies, nil")
	}
}

// attrs generates a list of wanted attributes given an attribute set and the
// names of the attributes that are expected.
func (g *generator) attrs(aset string, list []string) []Attribute {
	if aset == "" {
		panic("empty attribute set")
	}

	if len(list) == 0 {
		// Shortcut, no attributes wanted.
		return nil
	}

	idx := make(map[string]Attribute)
	for _, a := range g.asIndex[aset].Attributes {
		idx[a.Name] = a
	}

	var as []Attribute
	for _, l := range list {
		if attr, ok := idx[l]; ok {
			as = append(as, attr)
		}
	}

	if len(as) == 0 {
		panicf("found no attributes for set %q in list %v", aset, list)
	}

	return as
}

// pf is short for "printf" and writes formatted data to g.w. All format strings
// receive a trailing newline. If format is empty, a newline is written.
func (g *generator) pf(format string, v ...any) {
	if format == "" {
		fmt.Fprintln(g.w)
		return
	}

	fmt.Fprintf(g.w, format, v...)
	fmt.Fprintln(g.w)
}

// doOrDump signifies a Do operation or Dump operation.
type doOrDump bool

const (
	doOp   doOrDump = false
	dumpOp doOrDump = true
)

func (dod doOrDump) String() string {
	switch dod {
	case doOp:
		return "Do"
	case dumpOp:
		return "Dump"
	}

	panic("unreachable")
}

// requestOrReply signifies a Request type or Reply type.
type requestOrReply bool

const (
	doRequest requestOrReply = false
	doReply   requestOrReply = true
)

func (ror requestOrReply) String() string {
	switch ror {
	case doRequest:
		return "Request"
	case doReply:
		return "Reply"
	}

	panic("unreachable")
}

// camelCase transforms a string like "family-id" to "FamilyId".
func camelCase(s string) string {
	return strings.ReplaceAll(
		title(strings.ReplaceAll(s, "-", " ")),
		" ", "",
	)
}

// unixConst transforms a string like "ctrl-cmd-getpolicy" to
// "unix.CTRL_CMD_GETPOLICY".
func unixConst(s string) string {
	return "unix." + cases.Upper(language.AmericanEnglish).
		String(strings.ReplaceAll(s, "-", "_"))
}

// title transforms a string like "family id" to "Family Id".
func title(s string) string {
	return cases.Title(language.AmericanEnglish).String(s)
}

func panicf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}
