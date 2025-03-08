package main

import (
	"go.n16f.net/program"
)

func main() {
	var c *program.Command

	p := program.NewProgram("fpkg",
		"tools to manipulate freebsd packages")

	c = p.AddCommand("build", "build a package", cmdBuild)
	c.AddOptionalArgument("directory",
		"the directory containing files to package")
	c.AddOption("c", "config", "path", "fpkg.yaml",
		"the path of the configuration file")
	c.AddOption("v", "version", "string", "",
		"set the version of the package")

	p.ParseCommandLine()
	p.Run()
}
