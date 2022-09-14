# yamlnetlink [![Test Status](https://github.com/mdlayher/yamlnetlink/workflows/Test/badge.svg)](https://github.com/mdlayher/yamlnetlink/actions) [![Go Reference](https://pkg.go.dev/badge/github.com/mdlayher/yamlnetlink.svg)](https://pkg.go.dev/github.com/mdlayher/yamlnetlink) [![Go Report Card](https://goreportcard.com/badge/github.com/mdlayher/yamlnetlink)](https://goreportcard.com/report/github.com/mdlayher/yamlnetlink)

Package `yamlnetlink` provides support for parsing YAML netlink specifications.
MIT Licensed.

For more information, see:
  - https://lore.kernel.org/all/20220811022304.583300-1-kuba@kernel.org/
  - https://github.com/kuba-moo/ynl/blob/main/Documentation/netlink/netlink-bindings.rst

A goal of this project is to provide a `yamlnetlink-go` tool which automatically
generate Go code from YAML netlink specifications using
[`github.com/mdlayher/netlink`](https://github.com/mdlayher/netlink) and
[`github.com/mdlayher/genetlink`](https://github.com/mdlayher/genetlink).
