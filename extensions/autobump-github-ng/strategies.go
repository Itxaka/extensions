package main

import (
	"fmt"
	"os/exec"
)

var Strategies = map[string]func(c *CustomPackage) string{
	// Outputs the errors in the BSD format.
	"git": func(c *CustomPackage) string {
		fmt.Println("Called git")
		return ""
	},
	"custom": func(c *CustomPackage) string {
		versionHook := c.GetLabels()["autobump.version_hook"]
		cmd := exec.Command("bash", "-c", versionHook)
		versionHookOutput, err := cmd.CombinedOutput()
		if err != nil {
			Error("Error getting version for %s: %s", c.GetName(), err)
		}
		return string(versionHookOutput)
	},
	"ref": func(c *CustomPackage) string {
		fmt.Println("Called ref")
		return ""
	},
	"release": func(c *CustomPackage) string {
		fmt.Println("Called release")
		return ""
	},
	"snapshot": func(c *CustomPackage) string {
		fmt.Println("Called snapshot")
		return ""
	},
	"gentoo": func(c *CustomPackage) string {
		fmt.Println("Called gentoo")
		return ""
	},
	"sabayon": func(c *CustomPackage) string {
		fmt.Println("Called sabayon")
		return ""
	},
}