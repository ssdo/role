package role

import (
	"github.com/ssdo/utility"
	"github.com/ssgo/db"
	"github.com/ssgo/log"
	"sync"
	"time"
)

type accessInfo struct {
	access map[string]bool
	tags   []string
}

type Serve struct {
	utility.Starter
	config          *Config
	accessTable     map[string]map[string]*accessInfo
	accessTableLock sync.Mutex
	localVersion    uint64
}

func NewServe(config Config, logger *log.Logger) *Serve {

	if logger == nil {
		logger = log.DefaultLogger
	}

	if config.DB == nil {
		config.DB = db.GetDB("user", logger)
	}

	if config.SyncInterval == 0 {
		config.SyncInterval = time.Second * 5
	}

	if config.TableAccess.Table == "" {
		config.TableAccess.Table = "User"
	}
	if config.TableAccess.Role == "" {
		config.TableAccess.Role = "role"
	}
	if config.TableAccess.Module == "" {
		config.TableAccess.Module = "module"
	}
	if config.TableAccess.Access == "" {
		config.TableAccess.Access = "access"
	}
	if config.TableAccess.Tag == "" {
		config.TableAccess.Tag = "tag"
	}

	serve := Serve{
		Starter: utility.Starter{
			Interval: 0,
		},
		config:          &config,
		accessTable:     make(map[string]map[string]*accessInfo),
		accessTableLock: sync.Mutex{},
	}

	serve.Starter.Work = serve.sync

	return &serve
}
