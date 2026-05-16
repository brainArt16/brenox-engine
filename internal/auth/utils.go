package auth

import "golang.org/x/crypto/bcrypt"

/*
	HashPassword converts plain password
	into secure bcrypt hash.
*/

func HashPassword(password string) (string, error) {

	/*
		GenerateFromPassword returns:
		- hashed password bytes
		- error
	*/

	hashed, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		return "", err
	}

	return string(hashed), nil
}

func CheckPassword(password string,hash string) error {

	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}