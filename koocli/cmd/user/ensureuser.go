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
package user

import (
	"github.com/spf13/cobra"
)

// kubectl koo create user titi --provider crdsys --commonName "TITI" --comment "Small bird" --email "titi@cartoon.com" --passwordHash '$2a$10$dO9pDmqhwCVHkqBKdjynTONHRExZm2iDX3yzii/RUgNMt0U/wvNtG' --uid 2001

func init() {
	initUserParams(EnsureUserCmd)
}


var EnsureUserCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{},
	Short:   "Create or update user (Admin)",
	Hidden:  false,
	Run: func(cmd *cobra.Command, args []string) {
		applyUserCommand(cmd, args, "PUT")
	},
}
