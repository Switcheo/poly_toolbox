/*
* Copyright (C) 2020 The poly network Authors
* This file is part of The poly network library.
*
* The poly network is free software: you can redistribute it and/or modify
* it under the terms of the GNU Lesser General Public License as published by
* the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* The poly network is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
* GNU Lesser General Public License for more details.
* You should have received a copy of the GNU Lesser General Public License
* along with The poly network . If not, see <http://www.gnu.org/licenses/>.
 */
package main

import (
	"fmt"
	"os"

	"github.com/ontio/poly_toolbox/lib"
	"github.com/spf13/cobra"
)

func main() {
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:   "toolbox",
		Short: "Command line interface for interacting with poly network",
	}

	rootCmd.DisableFlagsInUseLine = true
	rootCmd.AddCommand(
		lib.PolyCmd(),
		lib.CarbonCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("failed to run: %v", err)
		os.Exit(1)
	}
}
