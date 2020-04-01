// This file is part of MinIO Console Server
// Copyright (c) 2020 MinIO, Inc.
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

package sessions

import (
	"testing"

	mcCmd "github.com/minio/mc/cmd"

	"github.com/stretchr/testify/assert"
)

func TestNewSession(t *testing.T) {
	assert := assert.New(t)
	// We will write a test against play
	// Probe the credentials
	cfg := mcCmd.Config{}
	sessionID, err := GetInstance().NewSession(&cfg)
	assert.NotEmpty(sessionID, "Session ID was returned empty")
	assert.Nil(err, "error creating a session")
}

func TestValidateSession(t *testing.T) {
	assert := assert.New(t)
	// We will write a test against play
	// Probe the credentials
	cfg := mcCmd.Config{}
	sessionID, _ := GetInstance().NewSession(&cfg)
	isValid := GetInstance().ValidSession(sessionID)
	assert.Equal(isValid, true, "Session was not found valid")
	isInvalid := GetInstance().ValidSession("random")
	assert.Equal(isInvalid, false, "Session was found valid")
}
