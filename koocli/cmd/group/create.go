/*
Copyright (C) 2020 Serge ALEXANDRE

# This file is part of koobind project

koobind is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

koobind is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with koobind.  If not, see <http://www.gnu.org/licenses/>.
*/
package group

import (
	"github.com/spf13/cobra"
)

// kubectl koo create group grp1 --description "Group 1"

func init() {
	initGroupParams(groupCreateCmd)
}

var groupCreateCmd = &cobra.Command{
	Use:    "create",
	Short:  "Create new group (Admin)",
	Hidden: false,
	Run: func(cmd *cobra.Command, args []string) {
		applyGroupCommand(cmd, args, "POST")
	},
}
