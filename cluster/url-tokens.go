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
package cluster

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"

	"k8s.io/client-go/kubernetes"

	uuid "github.com/satori/go.uuid"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type URLToken struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	UserID     uuid.UUID
	Expiration time.Time
	UsedFor    string
	Consumed   bool
}

// NewURLToken generates and stores a new urlToken for the provided user, with the specified validity
func NewURLToken(ctx *Context, userID *uuid.UUID, usedFor string, validity *time.Time) (*uuid.UUID, error) {
	urlToken := uuid.NewV4()
	query := `INSERT INTO
				url_tokens ("id", "user_id", "used_for", "expiration", "sys_created_by")
			  VALUES
				($1, $2, $3, $4, $5)`
	tx, err := ctx.TenantTx()
	if err != nil {
		return nil, err
	}
	// Execute query
	_, err = tx.Exec(query, urlToken, userID, usedFor, validity, ctx.WhoAmI)
	if err != nil {
		return nil, err
	}
	return &urlToken, nil
}

// GetTenantTokenDetails get the details for the provided urlToken
func GetTenantTokenDetails(ctx *Context, urlToken *uuid.UUID) (*URLToken, error) {
	var token URLToken
	// Get an individual token
	queryUser := `
		SELECT 
				id, user_id, expiration, used_for, consumed
			FROM 
				url_tokens
			WHERE id=$1 LIMIT 1`

	tx, err := ctx.TenantTx()
	if err != nil {
		return nil, err
	}

	row := tx.QueryRow(queryUser, urlToken)

	// Save the resulted query on the URLToken struct
	err = row.Scan(&token.ID, &token.UserID, &token.Expiration, &token.UsedFor, &token.Consumed)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// MarkTokenConsumed updates the record for the urlToken as is it has been used
func MarkTokenConsumed(ctx *Context, urlTokenID *uuid.UUID) error {
	query := `UPDATE url_tokens SET consumed=true WHERE id=$1`
	tx, err := ctx.TenantTx()
	if err != nil {
		return err
	}
	// Execute query
	_, err = tx.Exec(query, urlTokenID)
	if err != nil {
		ctx.Rollback()
		return err
	}

	return nil
}

// CompleteSignup takes a urlToken and a password and changes the user password and then marks the token as used
func CompleteSignup(ctx *Context, urlToken *URLToken, password string) error {
	if urlToken.Consumed {
		return errors.New("url token has already been consumed")
	}
	// update the user password
	err := setUserPassword(ctx, &urlToken.UserID, password)
	if err != nil {
		return err
	}

	// mark the url token as consumed
	err = MarkTokenConsumed(ctx, &urlToken.ID)
	if err != nil {
		return err
	}
	// mark the user as accepted invitation
	err = MarkInvitationAccepted(ctx, &urlToken.UserID)
	if err != nil {
		return err
	}
	return nil
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

// buildJwtTokenForURLToken builds a jwt token for a url token and tenant
func buildJwtTokenForURLToken(ctx *Context, urlTokenID *uuid.UUID) (*string, error) {
	// Create a new jwtToken object, specifying signing method and the claims
	// you would like it to contain.
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"t": urlTokenID.String(),
		"e": ctx.Tenant.ID.String(),
	})

	jwtSecret, err := getJWTSecretKey()
	if err != nil {
		return nil, err
	}

	// Sign and get the complete encoded jwtToken as a string using the secret
	tokenString, err := jwtToken.SignedString(jwtSecret)
	if err != nil {
		return nil, err
	}

	return &tokenString, nil
}

type URLJwtToken struct {
	Token    uuid.UUID `json:"t"`
	TenantID uuid.UUID `json:"e"`
	jwt.StandardClaims
}

// ParseAndValidateJwtToken parses and validates the jwt token
func ParseAndValidateJwtToken(tokenString string) (*URLJwtToken, error) {
	jwtSecret, err := getJWTSecretKey()
	if err != nil {
		return nil, err
	}
	token, err := jwt.ParseWithClaims(tokenString, &URLJwtToken{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*URLJwtToken); ok && token.Valid {
		return claims, nil
	}
	return nil, nil
}

// ValidateURLToken ensures Token expiration time and that it hasn't been consumed.
func ValidateURLToken(urlToken *URLToken) (err error) {
	// make sure this jwtToken is not already used
	if urlToken.Consumed {
		err = errors.New("this token has already been consumed")
		fmt.Println(err)
		return err
	}
	// make sure this jwtToken is intended for signup
	if urlToken.UsedFor != TokenSignupEmail {
		err = errors.New("invalid token")
		fmt.Println(err)
		return err
	}
	// make sure this jwtToken is not expired
	if !urlToken.Expiration.After(time.Now()) {
		err = errors.New("expired token")
		fmt.Println(err)
		return err
	}
	return nil
}
