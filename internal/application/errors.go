package application

import "errors"

var ErrNotFound = errors.New("resource not found")

var ErrProviderAccountAlreadyLinked = errors.New("provider account already linked to another user")
