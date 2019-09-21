// This file is part of MinIO Cloud Storage
// Copyright (c) 2019 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
package cmd

import (
	"github.com/minio/cli"
	"github.com/minio/minio/pkg/trie"
)

// Collection of mcs commands currently supported
var commands = []cli.Command{}

// Collection of mcs commands currently supported in a trie tree
var commandsTree = trie.NewTrie()

// registerCmd registers a cli command
func registerCmd(cmd cli.Command) {
	commands = append(commands, cmd)
	commandsTree.Insert(cmd.Name)
}
