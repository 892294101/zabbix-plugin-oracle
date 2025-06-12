/*
** Copyright (C) 2001-2025 Zabbix SIA
**
** Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated
** documentation files (the "Software"), to deal in the Software without restriction, including without limitation the
** rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to
** permit persons to whom the Software is furnished to do so, subject to the following conditions:
**
** The above copyright notice and this permission notice shall be included in all copies or substantial portions
** of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE
** WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
** COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
** TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
**/

package flag

import (
	"fmt"
	"os"
	"runtime"

	"golang.zabbix.com/sdk/plugin/comms"
)

// PrintVersion prints plugin version information to stdout.
func PrintVersion(pluginName, copyrightMessage string, majorVersion, minorVersion, patchVersion int, alphatag string) {
	fmt.Fprintf(os.Stdout, "Zabbix %s plugin\n", pluginName)
	fmt.Fprintf(
		os.Stdout,
		"Version %d.%d.%d%s, built with %s\n",
		majorVersion, minorVersion, patchVersion, alphatag, runtime.Version(),
	)
	fmt.Fprintf(os.Stdout, "Protocol version %s\n", comms.ProtocolVersion)
	fmt.Fprintln(os.Stdout)
	fmt.Fprintln(os.Stdout, copyrightMessage)
}
