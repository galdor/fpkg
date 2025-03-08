package main

import (
	"fmt"

	"go.n16f.net/program"
)

var buildId string

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

	p.AddCommand("version", "print the version of the program and exit",
		cmdVersion)

	p.ParseCommandLine()

	p.Run()
}

func cmdVersion(p *program.Program) {
	fmt.Println(buildId)
}
