package main

import (
	"errors"
	"fmt"
	"regexp"

	"go.n16f.net/program"
)

func cmdVersion(p *program.Program) {
	version, err := Version(buildId)
	if err != nil {
		p.Fatal("%v", err)
	}

	fmt.Println(version)
}

var buildIdRE = regexp.MustCompile("^v(.+?)(-(\\d+)-(.*))?$")

func Version(buildId string) (string, error) {
	matches := buildIdRE.FindStringSubmatch(buildId)
	if matches == nil {
		return "", errors.New("invalid build id format")
	}

	if len(matches) == 2 {
		return matches[1], nil
	}

	return matches[1] + "-" + matches[3], nil
}
