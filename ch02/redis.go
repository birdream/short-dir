package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/mattheath/base62"
)

const (
	URLIDKEY = "next.url.id"

	ShortlinkKey = "shortlink:%s:url"

	URLHashKey = "urlhash:%s:url"

	ShortlinkDetailKey = "shortlink:%s:detail"
)

type RedisCli struct {
	Cli *redis.Client
}

type URLDetail struct {
	URL                 string        `json:"url"`
	CreatedAt           string        `json:"created_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

func NewRedisClient(addr, pwd string, db int) *RedisCli {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})

	if _, err := c.Ping().Result(); err != nil {
		panic(err)
	}

	return &RedisCli{Cli: c}
}

// Shorten convert url to shortlink
func (r *RedisCli) Shorten(url string, exp int64) (string, error) {
	var (
		d      string
		err    error
		id     int64
		eid    string
		detail []byte
		h      string
	)

	// convert url to sha1 hash
	h = toSha1(url)

	if d, err = r.Cli.Get(fmt.Sprintf(URLHashKey, h)).Result(); err == redis.Nil {
		// not existed, nothing to do
	} else if err != nil {
		return "", err
	} else {
		if d == "{}" {
			// expired, nothing to do
		} else {
			return d, nil
		}
	}

	// increase the global counter
	if err = r.Cli.Incr(URLIDKEY).Err(); err != nil {
		return "", err
	}

	// encode global counter to base64
	if id, err = r.Cli.Get(URLIDKEY).Int64(); err != nil {
		return "", err
	}

	eid = base62.EncodeInt64(id)

	// store the url against this encoded id
	if err = r.Cli.Set(fmt.Sprintf(ShortlinkKey, eid), url, time.Minute*time.Duration(exp)).Err(); err != nil {
		return "", err
	}

	// store the url against the hash of it
	if err = r.Cli.Set(fmt.Sprintf(URLHashKey, h), eid, time.Minute*time.Duration(exp)).Err(); err != nil {
		return "", err
	}

	// store the detail info
	if detail, err = json.Marshal(
		&URLDetail{
			URL:                 url,
			CreatedAt:           time.Now().String(),
			ExpirationInMinutes: time.Duration(exp),
		},
	); err != nil {
		return "", err
	}

	if err = r.Cli.Set(fmt.Sprintf(ShortlinkDetailKey, eid), detail, time.Minute*time.Duration(exp)).Err(); err != nil {
		return "", err
	}

	return eid, nil
}

// ShortlinkInfo returns the detail of the shortlink
func (r *RedisCli) ShortlinkInfo(eid string) (interface{}, error) {
	var (
		d   string
		err error
	)

	if d, err = r.Cli.Get(fmt.Sprintf(ShortlinkDetailKey, eid)).Result(); err == redis.Nil {
		return "", StatusError{404, errors.New("Unknown short URL")}
	} else if err != nil {
		return "", err
	} else {
		return d, nil
	}
}

// Unshorten convert shortlink to url
func (r *RedisCli) Unshorten(eid string) (string, error) {
	var (
		url string
		err error
	)

	if url, err = r.Cli.Get(fmt.Sprintf(ShortlinkKey, eid)).Result(); err == redis.Nil {
		return "", StatusError{404, err}
	} else if err != nil {
		return "", err
	} else {
		return url, nil
	}
}

func toSha1(str string) string {
	var (
		sha = sha1.New()
	)
	return string(sha.Sum([]byte(str)))
}
