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

func WithGroup(group string) Option {
	return func(c *Config) {
		c.group = group
	}
}

// Config provides automatic YAML configuration file loading for Kong CLI applications.
// It looks for config files in /etc and the user's home directory, and supports
// an explicit --config flag override. Use [WithGroup] to namespace config files
// under a group directory.
//
// Usage:
//
//	type CLI struct {
//	    cmd.Commons
//	    Run RunCmd `cmd:"" default:"withargs" help:"Start the application."`
//	}
//
//	func main() {
//	    // name defaults to kong.Model.Name (the binary name)
//	    cfg := cmd.NewConfig(cmd.WithGroup("mygroup"))
//	    // loads from /etc/mygroup/<name>.yaml and ~/.<name>/mygroup.yaml
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
