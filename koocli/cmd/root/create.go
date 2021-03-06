/*
  Copyright (C) 2020 Serge ALEXANDRE

  This file is part of koobind project

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
package root

import (
	"github.com/koobind/koobind/koocli/cmd/group"
	"github.com/koobind/koobind/koocli/cmd/groupbinding"
	"github.com/koobind/koobind/koocli/cmd/user"
	"github.com/spf13/cobra"
)


func init() {
	CreateCmd.AddCommand(user.CreateUserCmd)
	CreateCmd.AddCommand(group.CreateGroupCmd)
	CreateCmd.AddCommand(groupbinding.CreateGroupBindingCmd)

}

var CreateCmd = &cobra.Command{
	Use:	"create",
	Short:  "Create ressources",
}


