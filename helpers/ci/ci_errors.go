package ci

import "fmt"

import log "github.com/sirupsen/logrus"

func logAndReturnError(err error) error {
	log.Error(err)
	return err
}

func logAndReturnErrorf(format string, args ...any) error {
	err := fmt.Errorf(format, args...)
	log.Error(err)
	return err
}
