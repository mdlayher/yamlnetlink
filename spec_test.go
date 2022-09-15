package yamlnetlink_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mdlayher/yamlnetlink"
)

func TestParse(t *testing.T) {
	s, err := yamlnetlink.Parse(strings.NewReader(nlctrlYAML))
	if err != nil {
		t.Fatalf("failed to parse nlctrl YAML: %v", err)
	}

	if diff := cmp.Diff(nlctrl(), s); diff != "" {
		t.Fatalf("unexpected Spec (-want +got):\n%s", diff)
	}
}

// nlctrl returns a well-formed YAML netlink Spec for the generic netlink nlctrl
// family, for use in tests.
func nlctrl() *yamlnetlink.Spec {
	return &yamlnetlink.Spec{
		Name:        "nlctrl",
		Protocol:    "genetlink-legacy",
		Description: "Generic netlink control protocol. Interface to query information about generic netlink families registered in the kernel - their names, ids, accepted messages and attributes.",
		UAPIHeader:  "linux/genetlink.h",
		AttributeSets: []yamlnetlink.AttributeSet{
			{
				Name:       "main",
				NamePrefix: "ctrl-attr-",
				Attributes: []yamlnetlink.Attribute{
					{
						Name:        "family-id",
						Type:        "u16",
						Description: "Numerical identifier of the family.",
					},
					{
						Name:        "family-name",
						Type:        "nul-string",
						Len:         "GENL_NAMSIZ - 1",
						Description: "String identifier of the family. Guaranteed to be unique.",
					},
					{
						Name: "version",
						Type: "u32",
					},
					{
						Name: "hdrsize",
						Type: "u32",
					},
					{
						Name: "maxattr",
						Type: "u32",
					},
					{
						Name:             "ops",
						Type:             "array-nest",
						NestedAttributes: "operation",
					},
					{
						Name:             "mcast-groups",
						Type:             "array-nest",
						NestedAttributes: "mcast-group",
					},
					{
						Name: "op",
						Type: "u32",
					},
					{
						Name:             "op-policy",
						Type:             "nest-type-value",
						TypeValue:        []string{"cmd"},
						NestedAttributes: "policy",
					},
					{
						Name:             "policy",
						Type:             "nest-type-value",
						TypeValue:        []string{"current-policy-idx", "attr-idx"},
						NestedAttributes: "nl-policy",
					},
				},
			},
			{
				Name:       "operation",
				NamePrefix: "ctrl-attr-op-",
				Attributes: []yamlnetlink.Attribute{
					{
						Name: "id",
						Type: "u32",
					},
					{
						Name: "flags",
						Type: "u32",
					},
				},
			},
			{
				Name:       "mcast-group",
				NamePrefix: "ctrl-attr-mcast-grp-",
				Attributes: []yamlnetlink.Attribute{
					{
						Name: "id",
						Type: "u32",
					},
					{
						Name: "name",
						Type: "nul-string", Len: "GENL_NAMSIZ - 1",
					},
				},
			},
			{
				Name:       "policy",
				NamePrefix: "ctrl-attr-policy-",
				Attributes: []yamlnetlink.Attribute{
					{
						Name: "do",
						Type: "u32",
					},
					{
						Name: "dump",
						Type: "u32",
					},
				},
			},
			{
				Name:       "nl-policy",
				NamePrefix: "nl-policy-type-attr-",
				Attributes: []yamlnetlink.Attribute{
					{
						Name: "type",
						Type: "u32",
					},
					{
						Name: "min-value-u",
						Type: "u64",
					},
					{
						Name: "max-value-u",
						Type: "u64",
					},
					{
						Name: "min-value-s",
						Type: "s64",
					},
					{
						Name: "max-value-s",
						Type: "s64",
					},
					{
						Name: "mask",
						Type: "u64",
					},
					{
						Name: "min-length",
						Type: "u32",
					},
					{
						Name: "max-length",
						Type: "u32",
					},
					{
						Name: "policy-idx",
						Type: "u32",
					},
					{
						Name: "policy-maxtype",
						Type: "u32",
					},
					{
						Name: "bitfield32-mask",
						Type: "u32",
					},
				},
			},
		},
		Operations: yamlnetlink.Operations{
			NamePrefix: "ctrl-cmd-",
			List: []yamlnetlink.Operation{
				{
					Name:         "getfamily",
					Description:  "Get information about genetlink family.",
					AttributeSet: "main",
					DontValidate: []string{"strict", "dump"},

					Do: yamlnetlink.OperationAttributes{
						Request: yamlnetlink.OperationAttributesList{
							Attributes: []string{"family-id", "family-name"},
						},
						Reply: yamlnetlink.OperationAttributesList{
							Attributes: []string{
								"family-id", "family-name", "version", "hdrsize", "maxattr", "ops", "mcast-groups",
							},
						},
					},
					Dump: yamlnetlink.OperationAttributes{
						Reply: yamlnetlink.OperationAttributesList{
							Attributes: []string{
								"family-id", "family-name", "version", "hdrsize", "maxattr", "ops", "mcast-groups",
							},
						},
					},
				},
				{
					Name:        "newfamily",
					Description: "Notification for new families being registered.",
					Notify:      "getfamily",
				},
				{
					Name:        "delfamily",
					Description: "Notification for families being unregistered.",
					Notify:      "getfamily",
				},
				{
					Name:        "newmcast-grp",
					Description: "Notification for new multicast groups.",
					Notify:      "getfamily",
				},
				{
					Name:        "delmcast-grp",
					Description: "Notification for deleted multicast groups.",
					Notify:      "getfamily",
				},
				{
					Name:         "getpolicy",
					Description:  "Get attribute policy for a genetlink family.",
					AttributeSet: "main",
					Dump: yamlnetlink.OperationAttributes{
						Request: yamlnetlink.OperationAttributesList{
							Attributes: []string{"family-id", "family-name", "op"},
						},
						Reply: yamlnetlink.OperationAttributesList{
							Attributes: []string{"family-id", "op-policy", "policy"},
						},
					},
				},
			},
		},
	}
}

// Copied from:
// https://github.com/kuba-moo/ynl/blob/main/Documentation/netlink/specs/genetlink.yaml
const nlctrlYAML = `
name: nlctrl

protocol: genetlink-legacy

description: |
  Generic netlink control protocol. Interface to query information about
  generic netlink families registered in the kernel - their names, ids,
  accepted messages and attributes.

uapi-header: linux/genetlink.h

attribute-sets:
  -
    name: main
    name-prefix: ctrl-attr-
    attributes:
      -
        name: family-id
        type: u16
        description: |
            Numerical identifier of the family.
      -
        name: family-name
        type: nul-string
        len: GENL_NAMSIZ - 1
        description: |
            String identifier of the family. Guaranteed to be unique.
      -
        name: version
        type: u32
      -
        name: hdrsize
        type: u32
      -
        name: maxattr
        type: u32
      -
        name: ops
        type: array-nest
        nested-attributes: operation
      -
        name: mcast-groups
        type: array-nest
        nested-attributes: mcast-group
      -
        name: op
        type: u32
      -
        name: op-policy
        type: nest-type-value
        type-value: [ cmd ]
        nested-attributes: policy
      -
        name: policy
        type: nest-type-value
        type-value: [ current-policy-idx, attr-idx ]
        nested-attributes: nl-policy
  -
    name: operation
    name-prefix: ctrl-attr-op-
    attributes:
      -
        name: id
        type: u32
      -
        name: flags
        type: u32
  -
    name: mcast-group
    name-prefix: ctrl-attr-mcast-grp-
    attributes:
      -
        name: id
        type: u32
      -
        name: name
        type: nul-string
        len: GENL_NAMSIZ - 1
  -
    name: policy
    name-prefix: ctrl-attr-policy-
    attributes:
      -
        name: do
        type: u32
      -
        name: dump
        type: u32
  -
    name: nl-policy
    name-prefix: nl-policy-type-attr-
    attributes:
      -
        name: type
        type: u32
      -
        name: min-value-u
        type: u64
      -
        name: max-value-u
        type: u64
      -
        name: min-value-s
        type: s64
      -
        name: max-value-s
        type: s64
      -
        name: mask
        type: u64
      -
        name: min-length
        type: u32
      -
        name: max-length
        type: u32
      -
        name: policy-idx
        type: u32
      -
        name: policy-maxtype
        type: u32
      -
        name: bitfield32-mask
        type: u32

operations:
  name-prefix: ctrl-cmd-
  list:
    -
      name: getfamily
      description: Get information about genetlink family.
      attribute-set: main
      dont-validate: [ strict, dump ]

      do:
        request:
          attributes:
            - family-id
            - family-name
        reply: &getfamily-do-reply
          attributes:
            - family-id
            - family-name
            - version
            - hdrsize
            - maxattr
            - ops
            - mcast-groups
      dump:
        reply: *getfamily-do-reply
    -
      name: newfamily
      description: Notification for new families being registered.
      notify: getfamily
    -
      name: delfamily
      description: Notification for families being unregistered.
      notify: getfamily
    -
      name: newmcast-grp
      description: Notification for new multicast groups.
      notify: getfamily
    -
      name: delmcast-grp
      description: Notification for deleted multicast groups.
      notify: getfamily
    -
      name: getpolicy
      description: Get attribute policy for a genetlink family.
      attribute-set: main

      dump:
        request:
          attributes:
            - family-id
            - family-name
            - op
        reply:
          attributes:
            - family-id
            - op-policy
            - policy
`
