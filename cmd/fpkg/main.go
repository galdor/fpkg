package main

import (
	"go.n16f.net/program"
)

var buildId string
var version string

func main() {
	version = buildId
	if version[0] == 'v' {
		version = version[1:]
	}

	var c *program.Command

	p := program.NewProgram("fpkg",
		"tools to manipulate freebsd packages")

	c = p.AddCommand("build", "build a package", cmdBuild)
	c.AddOptionalArgument("directory",
		"the directory containing files to package")
	c.AddOption("c", "config", "path", "fpkg.yaml",
		"the path of the configuration file")
	c.AddOption("o", "output", "path", ".",
		"the path of the output directory")
	c.AddOption("v", "version", "string", "",
		"set the version of the package")

	p.AddCommand("version", "print the version of the program and exit",
		cmdVersion)

	p.ParseCommandLine()

	p.Run()
}
