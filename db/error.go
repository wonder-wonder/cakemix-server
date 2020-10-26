package db

import "errors"

// Errors
var (
	// Auth
	ErrIDPassInvalid = errors.New("ID or password is incorrect")
	ErrInvalidToken  = errors.New("Token is expired or invalid")
	ErrExistUser     = errors.New("UserName or email is already exist")

	// User/Team
	ErrUserNotFound     = errors.New("The user is not found")
	ErrUserTeamNotFound = errors.New("User/team is not found")

	// Document/Folder
	ErrDocumentNotFound = errors.New("Document is not found")
	ErrFolderNotFound   = errors.New("Folder is not found")
)
