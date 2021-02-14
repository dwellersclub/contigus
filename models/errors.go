package models

//SimpleError Text based error
type SimpleError string

func (t SimpleError) Error() string {
	return string(t)
}