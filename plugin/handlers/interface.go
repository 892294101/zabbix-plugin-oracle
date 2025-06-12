package handlers

import (
	"context"
	"git.zabbix.com/ap/plugin-support/log"
)

const (
	PingFailed = 0
	PingOk     = 1
)

var Logger log.Logger

type Database interface {
	Database(name string) Session
	Ping(ctx context.Context) error
}

type Session interface {
}
