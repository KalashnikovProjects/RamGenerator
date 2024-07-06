package auth

import "golang.org/x/crypto/bcrypt"

func ComparePasswordWithHashed(password, hashedPassword string) error {
	incoming := []byte(password)
	existing := []byte(hashedPassword)
	return bcrypt.CompareHashAndPassword(existing, incoming)
}

func GenerateHashedPassword(password string) (string, error) {
	saltedBytes := []byte(password)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hash := string(hashedBytes)
	return hash, nil
}
