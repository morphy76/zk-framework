/*
Package lockerr provides error handling for the lock package.
*/
package lockerr

import "errors"

/*
ErrSubjectEmpty is returned when the subject is empty.
*/
var ErrSubjectEmpty = errors.New("subject cannot be empty")

/*
IsSubjectEmpty checks if the error is ErrSubjectEmpty.
*/
func IsSubjectEmpty(err error) bool {
	return err == ErrSubjectEmpty
}
