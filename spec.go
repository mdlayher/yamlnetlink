// Package yamlnetlink provides support for parsing YAML netlink specifications.
//
// For more information, see:
// - https://lore.kernel.org/all/20220811022304.583300-1-kuba@kernel.org/
// - https://github.com/kuba-moo/ynl/blob/main/Documentation/netlink/netlink-bindings.rst
package yamlnetlink

import (
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// A Spec is a YAML netlink specification.
type Spec struct {
	Name          string         `yaml:"name"`
	Protocol      string         `yaml:"protocol"`
	Description   string         `yaml:"description"`
	UAPIHeader    string         `yaml:"uapi-header"`
	AttributeSets []AttributeSet `yaml:"attribute-sets"`
	Operations    Operations     `yaml:"operations"`
}

// Parse parses a YAML netlink specification into a Spec.
func Parse(r io.Reader) (*Spec, error) {
	var s Spec
	if err := yaml.NewDecoder(r).Decode(&s); err != nil {
		return nil, err
	}

	// After decoding, clean up strings so they're more machine friendly.
	s.sanitize()

	return &s, nil
}

// sanitize tidies a Spec's description strings in-place.
func (s *Spec) sanitize() {
	sanitize(&s.Description)

	for i := range s.AttributeSets {
		for j := range s.AttributeSets[i].Attributes {
			sanitize(&s.AttributeSets[i].Attributes[j].Description)
		}
	}
}

// An AttributeSet describes the netlink attributes for a given family.
type AttributeSet struct {
	Name       string      `yaml:"name"`
	NamePrefix string      `yaml:"name-prefix"`
	Attributes []Attribute `yaml:"attributes"`
}

// An Attribute describes a single netlink attribute.
type Attribute struct {
	Name             string   `yaml:"name"`
	Type             string   `yaml:"type"`
	TypeValue        []string `yaml:"type-value"`
	Len              string   `yaml:"len"`
	Description      string   `yaml:"description"`
	NestedAttributes string   `yaml:"nested-attributes"`
}

// Operations describes the request and reply operations available for a netlink
// family.
type Operations struct {
	NamePrefix string      `yaml:"name-prefix"`
	List       []Operation `yaml:"list"`
}

// An Operation describes a single netlink request/reply operation.
type Operation struct {
	Name         string              `yaml:"name"`
	Description  string              `yaml:"description"`
	AttributeSet string              `yaml:"attribute-set"`
	DontValidate []string            `yaml:"dont-validate"`
	Notify       string              `yaml:"notify"`
	Do           OperationAttributes `yaml:"do"`
	Dump         OperationAttributes `yaml:"dump"`
}

// OperationAttributes describes the list of attributes used in netlink request
// and replies for a given Operation.
type OperationAttributes struct {
	Request OperationAttributesList `yaml:"request"`
	Reply   OperationAttributesList `yaml:"reply"`
}

// An OperationAttributesList contains the actual attributes used in a netlink
// request or reply operation.
type OperationAttributesList struct {
	Attributes []string `yaml:"attributes"`
}

// sanitize cleans up a string in-place.
func sanitize(s *string) {
	if s == nil {
		panic("yamlnetlink: cannot sanitize nil string pointer")
	}

	// Newlines become spaces, no leading or trailing whitespace.
	*s = strings.TrimSpace(strings.ReplaceAll(*s, "\n", " "))
}
