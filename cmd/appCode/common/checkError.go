package common

import "log"

func CheckError(err error) {
	if err != nil {
		log.Fatal("DEBUG: ", err)
	}
}