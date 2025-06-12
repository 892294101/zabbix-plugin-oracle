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
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"golang.zabbix.com/sdk/zbxerr"
	"golang.zabbix.com/sdk/zbxflag"
)

const usageMessageFormat = //
`%[1]s plugin for Zabbix agent 2

Usage of "%[2]s"
    %[2]s -h
    %[2]s -V

Options:
%[3]s
`

// HandleFlags registers and parses version and help command line flags.
// Help and version should be the only command line flags a plugin implement as
// when a plugin is loaded by Agent 2 no command line flags are passed, however
// arguments are used to establish a socket connection.
func HandleFlags(
	pluginName, pluginBinName, copyrightMessage, alphatag string,
	majorVersion, minorVersion, patchVersion int,
) error {
	var (
		versionFlag bool
		helpFlag    bool
	)

	// define a new flag set to avoid collisions with possible user defined
	// flags.
	cl := flag.NewFlagSet("", flag.ContinueOnError)

	f := zbxflag.Flags{
		&zbxflag.BoolFlag{
			Flag: zbxflag.Flag{
				Name:        "help",
				Shorthand:   "h",
				Description: "Display this help message",
			},
			Default: false,
			Dest:    &helpFlag,
		},
		&zbxflag.BoolFlag{
			Flag: zbxflag.Flag{
				Name:        "version",
				Shorthand:   "V",
				Description: "Print program version and exit",
			},
			Default: false,
			Dest:    &versionFlag,
		},
	}

	f.Register(cl)

	cl.Usage = func() {
		fmt.Fprintf(
			os.Stdout,
			usageMessageFormat,
			pluginName, filepath.Base(pluginBinName), f.Usage(),
		)
	}

	err := cl.Parse(os.Args[1:])
	if err != nil {
		return err
	}

	// a plugin when called by agent will always have at least one argument -
	// socket, hence if there are no arguments the plugin is called manually
	// and a help message should be printed.
	if len(os.Args) == 1 || helpFlag {
		cl.Usage()

		return zbxerr.ErrorOSExitZero
	}

	if versionFlag {
		PrintVersion(
			pluginName,
			copyrightMessage,
			majorVersion,
			minorVersion,
			patchVersion,
			alphatag,
		)

		return zbxerr.ErrorOSExitZero
	}

	return nil
}
