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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// defaultJWTExpTime is the Portal expiration time for the authentication token.
// Default is 12 hours
const defaultJWTExpTime = 12 * time.Hour

var (
	errAuthentication = errors.New("Authentication failed, check your access credentials")
)

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

func getConfig() *rest.Config {
	//when doing local development, mount k8s api via `kubectl proxy`
	config := &rest.Config{
		// TODO: switch to using cluster DNS.
		Host:            "http://localhost:8001",
		TLSClientConfig: rest.TLSClientConfig{},
		BearerToken:     "eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImRhc2hib2FyZC10b2tlbi1mZ2J4NSIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJkYXNoYm9hcmQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiIyNGE3Mjg1OC00YjE4LTRhZDEtYjM4YS03ZTA2NGM2ODI1ZmEiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDpkYXNoYm9hcmQifQ.OTj-gB3OnDA5yDmtRZVF9wxMx-6fT1o3vSmd_lZrCpddTBgSkUb2vnaB8eVDQ_DKN2fHsnWw6JvZoPftJ27gKVZ_dAM_21XwgUJy72_lhI_XLinGcx5TAqObxhLp5-YlCTQPDbVEW56DUs59mvx2KKaYeeS7KE-ORYN4wpH6ecZnhUR7_jhSdJAb9MBp3reUU6Iou2YDfEHtHgrSoF7EpZrQME8zjtTQE0Fkl6YavKA1zjHMg-yKuiFRjLkKcrcXyYa_j4lFXL_ZGEICy94FsjGAPv4iwCqZW9ruTU9EX0B0BbG4xGYEZfgG6B5iqIUdleYzHl86eSpWQMS5H5xguQ",
		BearerTokenFile: "some/file",
	}

	return config
}

// getJWTSecretKey gets jwt secret key from kubernetes secrets
func getJWTSecretKey() ([]byte, error) {
	config := getConfig()
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println(err)
		panic(err.Error())
	}
	res, err := clientset.CoreV1().Secrets("default").Get("jwtkey", metav1.GetOptions{})
	return []byte(string(res.Data["M3_JWT_KEY"])), err
}

// Login Handles the Login request by receiving the user credentials
// and returning a hashed token.
func Login(w http.ResponseWriter, r *http.Request) {
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
	// TODO: password will come not hashed and stored hashed so we need to hash it and compare it against db
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
	jwtKey, err := getJWTSecretKey()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
