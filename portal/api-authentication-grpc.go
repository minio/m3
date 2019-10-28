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
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	jwtgo "github.com/dgrijalva/jwt-go"
	pq "github.com/lib/pq"
	cluster "github.com/minio/m3/cluster"
	common "github.com/minio/m3/common"
	pb "github.com/minio/m3/portal/stubs"
	uuid "github.com/satori/go.uuid"
	metadata "google.golang.org/grpc/metadata"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

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
		return []byte{}, err
	}
	res, err := clientset.CoreV1().Secrets("default").Get("jwtkey", metav1.GetOptions{})
	return []byte(string(res.Data["M3_JWT_KEY"])), err
}

// Login handles the Login request by receiving the user credentials
// and returning a hashed token.
func (s *server) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	log.Printf("Calling Login")
	// Create Credentials
	// TODO: validate credentials: username->email, tenant->shortname?
	var res pb.LoginResponse
	tenant := in.GetCompany()
	email := in.GetEmail()
	pwd := in.GetPassword()

	// Password validation
	user, ok := getUser(tenant, email)
	// If a password exists for the given user
	// AND, if it is the same as the password we received, then we can move ahead
	expectedPwd := user.Password
	// TODO: password will come not hashed and stored hashed so we need to hash it and compare it against db
	if !ok || expectedPwd != pwd {
		err := errors.New("Wrong tenant, email or password")
		res.Error = err.Error()
		return &res, err
	}

	// Add the session within a transaction in case anything goes wrong during the adding process
	db := cluster.GetInstance().Db
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		res.Error = err.Error()
		return &res, err
	}

	// Set query parameters
	loginCtx := cluster.NewContext(ctx, tx)
	// Insert a new session with random string as id
	sessionID := common.GetRandString(32, "sha256")
	userID := uuid.NewV4()

	query :=
		`INSERT INTO
				m3.provisioning.sessions ("id","user_id", "occurred_at")
			  VALUES
				($1,$2,$3)`

	// Execute Query
	_, err = loginCtx.Tx.Exec(query, sessionID, userID, time.Now())
	if err != nil {
		tx.Rollback()
		res.Error = err.Error()
		return &res, err
	}
	fmt.Println("sessionID: ", sessionID)

	// Return session in Token Response
	res.JwtToken = sessionID

	// if no error happened to this point commit transaction
	err = tx.Commit()
	if err != nil {
		res.Error = err.Error()
		return &res, err
	}
	return &res, nil
}

// ValidateGRPCToken extracts the token from the header of the request and validates it
func ValidateGRPCToken(ctx context.Context) (bool, error) {
	// Get the JWT string from context
	md, ok := metadata.FromIncomingContext(ctx)
	tknStr := md["token"][0]
	fmt.Printf("token: %s", tknStr)
	if !ok {
		fmt.Println("Error getting Token")
		return false, errors.New("Error getting Token")
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
			return false, err
		}
		return false, err
	}
	if !tkn.Valid {
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

	jwtKey, err := getJWTSecretKey()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if _, ok := jwtToken.Claims.(*jwtgo.StandardClaims); ok {
		return jwtKey, nil
	}

	return nil, errAuthentication
}

// getUser searches for the user in the defined tenant's database
// and returns the User if it was found
func getUser(tenant string, email string) (user User, ok bool) {

	bgCtx := context.Background()
	db := cluster.GetInstance().GetTenantDB(tenant)

	tx, err := db.BeginTx(bgCtx, nil)
	if err != nil {
		tx.Rollback()
		return user, false
	}
	loginCtx := cluster.NewContext(bgCtx, tx)

	// Get user from tenants database
	var userEmail string
	quoted := pq.QuoteIdentifier(tenant)
	queryUser := fmt.Sprintf(`
		SELECT 
				t1.id, t1.email, t1.password
			FROM 
				tenants.%s.users t1
			WHERE email=$1`, quoted)
	row := loginCtx.Tx.QueryRow(queryUser, email)

	// Save the resulted query on the User struct
	err = row.Scan(&user.UUID, &userEmail, &user.Password)
	if err != nil {
		tx.Rollback()
		return user, false
	}

	// if no error happened to this point commit transaction
	err = loginCtx.Tx.Commit()
	if err != nil {
		return user, false
	}
	return user, true
}
