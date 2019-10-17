package portal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// PortalExpTime is the Portal expiration time for the authentication token.
// Used in minutes.
const PortalExpTime = 1

// temporary user db for testing
// {"userid":"hashedpassword",...}
var users = map[string]map[string]string{
	"acme": {
		"cesnietor@acme.com": "cesnietor_hashed",
		"daniel@acme.com":    "daniel_hashed",
	},
}
var jwtKey = []byte("secret_key")

// Credentials requested on the portal to log in
type Credentials struct {
	Tenant   string `json:"tenant"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims is a struct that will be encoded to a JWT, contains jwt.StandardClaims
// as an embedded type to provide fields like expiry time.
// Claims should not have secret information
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Login Handles the Login request by receiving the user credentials
// and returning a hased token.
func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	// Get Json Body and return into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Password validation
	expectedPwd, ok := users[creds.Tenant][creds.Username]

	// If a password exists for the given user
	// AND, if it is the same as the password we received, then we can move ahead
	// if NOT, then we return an "Unauthorized" status
	if !ok || expectedPwd != creds.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Declare the expiration time of the token
	expTime := time.Now().Add(PortalExpTime * time.Minute)

	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expTime.Unix(),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the new token as the users `token` cookin
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expTime,
	})
}

func ValidateTokenFromCookie(w http.ResponseWriter, r *http.Request) bool {
	// We can obtain the session token from the requests cookies, which come with every request
	c, err := r.Cookie("token")
	if err != nil {
		fmt.Println(err)
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return false
	}

	// Get the JWT string from the cookie
	tknStr := c.Value

	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passsing the key in this method as well.
	// This method will return an error if the token is invalid
	// (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		fmt.Println(err)
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		w.WriteHeader(http.StatusBadRequest)
		return false
	}
	if !tkn.Valid {
		fmt.Println("tkn not valid")
		w.WriteHeader(http.StatusUnauthorized)
		return false
	}
	return true
}
