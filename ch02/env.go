package main

import (
	"log"
	"os"
	"strconv"
)

type Env struct {
	S Storage
}

func getEnv() *Env {
	var (
		addr string
		pwd  string
		dbS  string
		db   int
		err  error
	)

	if addr = os.Getenv("APP_REDIS_ADDR"); addr == "" {
		addr = "localhost:6379"
	}

	if pwd = os.Getenv("APP_REDIS_PASSWD"); pwd == "" {
		pwd = ""
	}

	if dbS = os.Getenv("APP_REDIS_DB"); dbS == "" {
		dbS = "0"
	}

	if db, err = strconv.Atoi(dbS); err != nil {
		log.Fatal(err)
	}

	log.Printf("connect to redis (addr %s pwd %s db %d)", addr, pwd, db)

	r := NewRedisClient(addr, pwd, db)

	return &Env{S: r}
}
