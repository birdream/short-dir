package main_test

import (
	main "main"

	"github.com/stretchr/testify/mock"
)

const (
	expTime       = 60
	longURL       = "http://www.norman.com"
	shortLink     = "IFHzaO"
	shortlinkInfo = "{\"url\":\"http://www.norman.com\",\"created_at\":\"2019-11-12 15:35:53.542478 +0800 CST m=+218.985180477\",\"expiration_in_minutes\":5}"
)

type storageMock struct {
	mock.Mock
}

var mockR *storageMock

func (s *storageMock) Shorten(url string, exp int64) (string, error) {
	args := s.Called(url, exp)
	return args.String(0), args.Error(1)
}

func (s *storageMock) ShortlinkInfo(eid string) (interface{}, error) {
	args := s.Called(eid)
	return args.String(0), args.Error(1)
}

func (s *storageMock) Unshorten(eid string) (string, error) {
	args := s.Called(eid)
	return args.String(0), args.Error(1)
}

func init() {
	app := main.App{}
	mockR = new(storageMock)
	app.Initialize()
}
