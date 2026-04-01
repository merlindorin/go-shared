package cmd

import (
	"fmt"
)

// Licence holds the licence content of an application.
// It is automatically registered as a subcommand when embedded via [Commons].
//
// Usage:
//
//	//go:embed LICENSE
//	var license string
//
//	func main() {
//	    var cli CLI
//	    ctx := kong.Parse(&cli,
//	        kong.DefaultEnvars("MYAPP"),
//	        kong.Bind(cmd.NewLicence(license)),
//	    )
//	    _ = ctx.Run()
//	}
//
// Then run:
//
//	$ myapp licence
//	MIT License ...
type Licence struct {
	content string
}

// NewLicence creates a new Licence instance with the provided licence content.
func NewLicence(s string) Licence {
	return Licence{content: s}
}

// Run prints the licence content to the console when the licence command is invoked.
// This method satisfies the Kong interface contract for command execution.
func (l Licence) Run() error {
	content := l.content
	if content == "" {
		content = "Proprietary – All rights reserved.\n"
	}

	//nolint:forbidigo // using a custom writer is not necessary here
	fmt.Print(content)

	return nil
}
