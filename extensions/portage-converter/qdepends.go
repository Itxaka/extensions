/*
Copyright (C) 2020  Daniele Rondina <geaaru@sabayonlinux.org>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	helpers "github.com/MottainaiCI/lxd-compose/pkg/helpers"

	gentoo "github.com/Sabayon/pkgs-checker/pkg/gentoo"
)

type QDependsResolver struct{}

func NewQDependsResolver() *QDependsResolver {
	return &QDependsResolver{}
}

func runQdepends(solution *PortageSolution, runtime bool) error {
	var outBuffer, errBuffer bytes.Buffer

	cmd := []string{"qdepends", "-qC", "-F", "deps"}

	if runtime {
		cmd = append(cmd, "-r")
	} else {
		cmd = append(cmd, "-bd")
	}

	pkg := solution.Package.GetPackageName()
	if solution.Package.Slot != "0" {
		pkg = fmt.Sprintf("%s:%s", pkg, solution.Package.Slot)
	}
	cmd = append(cmd, pkg)

	qdepends := exec.Command(cmd[0], cmd[1:]...)
	qdepends.Stdout = helpers.NewNopCloseWriter(&outBuffer)
	qdepends.Stderr = helpers.NewNopCloseWriter(&errBuffer)

	err := qdepends.Start()
	if err != nil {
		return err
	}

	err = qdepends.Wait()
	if err != nil {
		return err
	}

	ans := qdepends.ProcessState.ExitCode()
	if ans != 0 {
		return errors.New("Error on running rdepends for package " + pkg + ": " + errBuffer.String())
	}

	out := outBuffer.String()
	if len(out) > 0 {
		// Drop prefix
		out = out[6:]

		deps := strings.Split(out, " ")

		for _, dep := range deps {

			gp, err := gentoo.ParsePackageStr(strings.TrimSuffix(dep, "\n"))
			if err != nil {
				return errors.New("On convert dep " + dep + ": " + err.Error())
			}
			if runtime {
				solution.RuntimeDeps = append(solution.RuntimeDeps, *gp)
			} else {
				solution.BuildDeps = append(solution.BuildDeps, *gp)
			}
		}

	} else {
		typeDeps := "build-time"
		if runtime {
			typeDeps = "runtime"
		}
		fmt.Println(fmt.Sprintf("No %s dependencies found for package %s.",
			typeDeps, solution.Package.GetPackageName()))
	}

	return nil
}

func (r *QDependsResolver) Resolve(pkg string) (*PortageSolution, error) {
	ans := &PortageSolution{
		BuildDeps:   make([]gentoo.GentooPackage, 0),
		RuntimeDeps: make([]gentoo.GentooPackage, 0),
	}

	gp, err := gentoo.ParsePackageStr(pkg)
	if err != nil {
		return nil, err
	}

	ans.Package = *gp

	// Retrieve runtime deps
	err = runQdepends(ans, true)
	if err != nil {
		return nil, err
	}

	// Retrieve build-time deps
	err = runQdepends(ans, false)
	if err != nil {
		return nil, err
	}

	return ans, nil
}
