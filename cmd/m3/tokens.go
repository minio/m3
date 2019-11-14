// This file is part of MinIO Kubernetes Cloud
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

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"time"

	"github.com/pelletier/go-toml"
)

type OperatorTokens struct {
	Token               string
	RefreshToken        string
	Expires             time.Time
	RefreshTokenExpires time.Time
}

func SaveOpToken(opToken *OperatorTokens) error {
	// serialize to toml then save
	outToml, err := toml.Marshal(opToken)
	if err != nil {
		return err
	}
	// save to file
	// get home folder
	usr, err := user.Current()
	if err != nil {
		return err
	}
	homeFolder := fmt.Sprintf("%s/.op8r", usr.HomeDir)
	tokenFile := fmt.Sprintf("%s/token", homeFolder)
	// make folder if it doesn't exist
	_ = os.Mkdir(homeFolder, 0777)
	err = ioutil.WriteFile(tokenFile, outToml, 0644)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func GetOpTokens() (*OperatorTokens, error) {
	// get home folder
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	homeFolder := fmt.Sprintf("%s/.op8r", usr.HomeDir)
	tokenFile := fmt.Sprintf("%s/token", homeFolder)
	// read file
	dat, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, err
	}
	var opToken OperatorTokens
	err = toml.Unmarshal(dat, &opToken)

	if err != nil {
		return nil, err
	}
	return &opToken, nil

}
