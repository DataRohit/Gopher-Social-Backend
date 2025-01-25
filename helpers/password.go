package helpers

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes the password using bcrypt.
//
// Parameters:
//   - password (string): The password to be hashed.
//
// Returns:
//   - string: The hashed password.
//   - error: An error if hashing fails.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// ComparePassword compares the hashed password with the plain password.
//
// Parameters:
//   - hashedPassword (string): The hashed password to compare.
//   - plainPassword (string): The plain password to compare.
//
// Returns:
//   - error: An error if passwords do not match.
//     Returns nil if passwords match.
func ComparePassword(hashedPassword string, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
