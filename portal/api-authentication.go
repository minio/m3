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

package portal

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	jwtreq "github.com/dgrijalva/jwt-go/request"
)

// defaultJWTExpTime is the Portal expiration time for the authentication token.
// Default is five minutes
const defaultJWTExpTime = 5 * time.Minute

var (
	errAuthentication = errors.New("Authentication failed, check your access credentials")
)

// temporary user db for testing
// {"userid":"hashedpassword",...}
var users = map[string]map[string]map[string]string{
	"acme": {
		"cesnietor@acme.com": {
			"password": "cesnietor_hashed",
			"uuid":     "123e4567-e89b-12d3-a456-426655440000",
		},
		"daniel@acme.com": {
			"password": "daniel_hashed",
			"uuid":     "00112233-4455-6677-8899-aabbccddeeff",
		},
	},
}

// TODO: jwtKey should be gotten from a m3 global config
var jwtKey = []byte("secret_key")

// Credentials requested on the portal to log in
type Credentials struct {
	Tenant   string `json:"tenant"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims is a struct that will be encoded to a JWT, contains jwtgo.StandardClaims
// as an embedded type to provide fields like expiry time.
// Claims should not have secret information
type Claims struct {
	jwtgo.StandardClaims
}

type User struct {
	Tenant   string
	IsAdmin  bool
	Password string
	UUID     string
}

type LoginResp struct {
	Token string `json:"token"`
}

// Login Handles the Login request by receiving the user credentials
// and returning a hashed token.
func Login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("LogIn")
	// Create Credentials
	// TODO: validate credentials: username->email, tenant->shortname?
	var creds Credentials
	// Get Json Body and return into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Password validation
	user, ok := getUser(creds.Tenant, creds.Username)

	// If a password exists for the given user
	// AND, if it is the same as the password we received, then we can move ahead
	// if NOT, then we return an "Unauthorized" status
	expectedPwd := user.Password
	if !ok || expectedPwd != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Declare the token with signing method and the claims
	token := jwtgo.NewWithClaims(jwtgo.SigningMethodHS512, jwtgo.StandardClaims{
		// Declare the expiration time of the token
		ExpiresAt: UTCNow().Add(defaultJWTExpTime).Unix(),
		Subject:   user.UUID,
	})

	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return Token in Response
	resp := &LoginResp{}
	resp.Token = tokenString

	output, err := json.Marshal(resp)
	if err != nil {
		log.Fatal("Cannot Marshal error")
	}

	w.Write(output)
}

// getUser returns the user struct
func getUser(tenant string, username string) (user User, ok bool) {
	// TODO: get user from db and validate password (GetUser user.key, user.uuid)
	dbUser, ok := users[tenant][username]
	if !ok {
		return user, ok
	}
	user.Password = dbUser["password"]
	user.UUID = dbUser["uuid"]
	return user, ok
}

// ValidateWebToken extracts the token from the header of the request and validates it
func ValidateWebToken(w http.ResponseWriter, r *http.Request) (bool, error) {
	// Get the JWT string from Header
	tknStr, err := jwtreq.AuthorizationHeaderExtractor.ExtractToken(r)
	if err != nil {
		fmt.Println(err)
		if err == jwtreq.ErrNoTokenInRequest {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return false, err
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return false, err
	}

	// Initialize a new instance of `Claims`
	claims := &jwtgo.StandardClaims{}

	// Parse the jwtgo string and store the result in `claims`.
	// Note that we are passsing the key in this method as well.
	// This method will return an error if the token is invalid
	// (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwtgo.ParseWithClaims(tknStr, claims, webTokenCallback)
	if err != nil {
		if err == jwtgo.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return false, err
		}
		w.WriteHeader(http.StatusBadRequest)
		return false, err
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return false, err
	}
	return true, nil
}

func webTokenCallback(jwtToken *jwtgo.Token) (interface{}, error) {
	if _, ok := jwtToken.Method.(*jwtgo.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", jwtToken.Header["alg"])
	}

	if err := jwtToken.Claims.Valid(); err != nil {
		return nil, errAuthentication
	}

	if _, ok := jwtToken.Claims.(*jwtgo.StandardClaims); ok {
		return jwtKey, nil
	}

	return nil, errAuthentication
}
