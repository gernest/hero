package hero

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/now"
	"github.com/pborman/uuid"
)

// SimpleTokenGen implements TokenGenerator interface
type SimpleTokenGen struct{}

// Generate returns a UUID v4 string
func (s *SimpleTokenGen) Generate() string {
	return uuid.NewRandom().String()
}

type JWTTokenGen struct {
	publicKey  []byte
	privateKey []byte
}

func NewJWTGen(public, private []byte) *JWTTokenGen {
	return &JWTTokenGen{publicKey: public, privateKey: private}
}

func (j *JWTTokenGen) Generate() string {
	exp := now.EndOfYear()
	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims["exp"] = exp.Unix()
	tok, err := token.SignedString(j.privateKey)
	if err != nil {
		panic(err)
	}
	return tok
}
