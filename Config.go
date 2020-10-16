package role

import (
	"github.com/ssgo/db"
	"github.com/ssgo/log"
	"github.com/ssgo/redis"
	"time"
)

// 角色权限表
type TableAccess struct {
	Table   string // 表名
	Role    string // 角色字段名
	Module  string // 模块字段名
	Access  string // 权限字段名
	Tag     string // 标签字段名
	Deleted string // 是否删除字段名
	Version string // 版本号字段名
}

type Config struct {
	Redis        *redis.Redis  // Redis连接池
	DB           *db.DB        // 数据库连接池
	SyncInterval time.Duration // 数据同步周期
	TableAccess  TableAccess   // 数据库用户表配置
}

var defaultServe *Serve

func Init(config Config, logger *log.Logger) {
	defaultServe = NewServe(config, logger)
}

func Access(roles []string, module, access string) (accept bool, tags []string) {
	if defaultServe == nil {
		log.DefaultLogger.Error("no default serve for access")
		return false, nil
	}
	return defaultServe.Access(roles, module, access)
}
