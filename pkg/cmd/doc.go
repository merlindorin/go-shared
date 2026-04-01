// Package cmd defines structures and implements functionality required for
// creating and managing the command-line interface (CLI) of an application.
// It leverages the Kong library to parse CLI arguments and execute the associated commands.
// The package provides common CLI components such as version information and license display,
// ensuring consistency and reusability throughout the application's command-line utilities.
//
// Struct tags use "name" for CLI flag binding. To also support environment variables,
// use [kong.DefaultEnvars] when creating the Kong parser. This automatically maps
// flags to environment variables (e.g. --postgres-host becomes POSTGRES_HOST).
//
// Usage:
//
//	type CLI struct {
//	    cmd.Commons
//	    cmd.HTTPServer `embed:"" prefix:""`
//	    cmd.Postgres   `embed:"" prefix:""`
//	}
//
//	func main() {
//	    var cli CLI
//	    ctx := kong.Parse(&cli, kong.DefaultEnvars("MYAPP"))
//	    // ...
//	    _ = ctx.Run()
//	}
package cmd
