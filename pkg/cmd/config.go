package cmd

import (
	"fmt"
	"path"

	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
)

type Option func(c *Config)

func (o Option) Apply(c *Config) {
	o(c)
}

// WithGroup namespaces the config under a parent directory.
// Setting a group changes the lookup paths from the per-binary
// defaults to a shared parent: `~/.<group>/<name>.yaml` and
// `/etc/<group>/<name>.yaml`.
//
// Use a group when several binaries share state on disk and you
// want their config files to live alongside that data — e.g. a
// daemon `myapp` whose state lives at `~/.myapp/` and a sibling
// CLI `myapp-cli` can both read from `~/.myapp/<binary>.yaml` by
// constructing their `Config` with `WithGroup("myapp")`.
//
// Pass an empty group (or skip the option) for single-binary
// projects — the default lookup is `~/.<binary>/config.yaml` and
// `/etc/<binary>/config.yaml`, which is enough for most CLIs.
func WithGroup(group string) Option {
	return func(c *Config) {
		c.group = group
	}
}

// WithName overrides the config file's basename. Without this
// option the basename comes from the parsed kong application name
// (`kong.Model.Name`, normally the binary name). Override when
// the binary name and the desired config filename should differ —
// most often to colocate a binary's config under a state directory
// owned by its broader project, with a generic basename.
//
// Combines with [WithGroup]:
//
//   - `WithGroup("myapp") + WithName("config")` →
//     `~/.myapp/config.yaml`, `/etc/myapp/config.yaml`
//   - `WithGroup("myapp")` alone →
//     `~/.myapp/<binary>.yaml`, `/etc/myapp/<binary>.yaml`
//   - `WithName("config")` alone →
//     `~/.config.yaml`, `/etc/config.yaml`
//     (rarely useful; prefer pairing with `WithGroup`)
//
// Empty string is treated as "no override" — the binary name
// wins.
func WithName(name string) Option {
	return func(c *Config) {
		c.name = name
	}
}

// Config provides automatic YAML configuration file loading for
// Kong CLI applications. It exposes a `--config` / `-c` flag for
// an explicit override, and otherwise looks for files in
// host-scoped `/etc/...` and user-scoped `~/...` locations.
//
// Lookup order at startup, first match wins:
//
//  1. `--config <path>` (the [Config.ConfigFile] flag)
//  2. user-scoped — `~/.<group>/<name>.yaml`, or
//     `~/.<name>/config.yaml` when no group is set
//  3. host-scoped — `/etc/<group>/<name>.yaml`, or
//     `/etc/<name>/config.yaml` when no group is set
//
// `<name>` defaults to the kong application name (the binary
// name); override with [WithName]. `<group>` is unset by default;
// set it with [WithGroup] to colocate the config alongside other
// state on disk.
//
// Usage:
//
//	type CLI struct {
//	    cmd.Commons
//	    Run RunCmd `cmd:"" default:"withargs" help:"Start the application."`
//	}
//
//	func main() {
//	    // Defaults: ~/.<binary>/config.yaml + /etc/<binary>/config.yaml
//	    cfg := cmd.NewConfig()
//
//	    // With group, colocates under a shared state dir:
//	    //   ~/.<group>/<binary>.yaml + /etc/<group>/<binary>.yaml
//	    cfg = cmd.NewConfig(cmd.WithGroup("mygroup"))
//
//	    // With group + name, fully custom:
//	    //   ~/.<group>/<name>.yaml + /etc/<group>/<name>.yaml
//	    cfg = cmd.NewConfig(cmd.WithGroup("mygroup"), cmd.WithName("config"))
//
//	    var cli CLI
//	    ctx := kong.Parse(&cli,
//	        kong.DefaultEnvars("MYAPP"),
//	        kong.Bind(cfg),
//	    )
//	    _ = ctx.Run()
//	}
type Config struct {
	// name application config name
	name string

	// group application group name
	group string

	ConfigFile kong.ConfigFlag `json:"config" name:"config" short:"c" help:"Full path to a user-supplied config file"`
}

func NewConfig(opts ...Option) *Config {
	c := &Config{}

	for _, opt := range opts {
		opt.Apply(c)
	}

	return c
}

func (c *Config) BeforeResolve(k *kong.Kong) error {
	name := c.name
	if name == "" {
		name = k.Model.Name
	}

	etcFileName := path.Join("/etc", c.group, fmt.Sprintf("%s.yaml", name))
	homeFileName := path.Join("~", fmt.Sprintf(".%s", c.group), fmt.Sprintf("%s.yaml", name))

	if c.group == "" {
		etcFileName = path.Join("/etc", name, "config.yaml")
		homeFileName = path.Join("~", fmt.Sprintf(".%s", name), "config.yaml")
	}

	return kong.Configuration(kongyaml.Loader, etcFileName, homeFileName).Apply(k)
}
