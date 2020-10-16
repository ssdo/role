package role

import (
	"fmt"
	"github.com/ssgo/u"
	"strings"
)

func (serve *Serve) Access(roles []string, module, access string) (accept bool, tags []string) {
	allTags := make([]string, 0)

	serve.accessTableLock.Lock()
	accessTable := serve.accessTable
	serve.accessTableLock.Unlock()

	for _, role := range roles {
		roleAccept := false
		roleModules := accessTable[role]
		if roleModules == nil {
			continue
		}
		modules := strings.Split(module, "-")
		for i := len(modules); i >= 1; i-- {
			moduleAccesses := roleModules[strings.Join(modules[0:i], "-")]
			if moduleAccesses == nil {
				continue
			}
			if moduleAccesses.access[access] {
				roleAccept = true
				if moduleAccesses.tags != nil {
					allTags = u.AppendUniqueStrings(allTags, moduleAccesses.tags)
				}
			}
			if moduleAccesses.access["!"+access] {
				// 遇到一个拒绝立刻跳过此角色的比较
				continue
			}
		}

		// 此角色比对成功，即宣告成功但仍需处理完所有角色的比对以获得完整的 tags
		if roleAccept {
			accept = true
		}
	}

	if accept {
		return true, allTags
	} else {
		return false, nil
	}
}

func (serve *Serve) sync() {
	table := serve.config.TableAccess

	// 如果设置了 Deleted，增加筛选条件
	where := "1"
	args := make([]interface{}, 0)
	if table.Deleted != "" {
		where = fmt.Sprint("`", table.Deleted, "`=?")
		args = append(args, 1)
	}

	// 如果设置了 Version，只获取增量部分数据
	if table.Version != "" {
		maxVersion := uint64(serve.config.DB.Query(fmt.Sprint("SELECT MAX(`", table.Version, "`) FROM `", table.Table, "`")).IntOnR1C1())
		where += fmt.Sprint(" AND `", table.Version, "` BETWEEN ? AND ?")
		args = append(args, serve.localVersion, maxVersion)
		serve.localVersion = maxVersion
	}

	// 从数据库中获得数据
	results := serve.config.DB.Query(fmt.Sprint("SELECT `", table.Role, "`,`", table.Module, "`,`", table.Access, "`,`", table.Tag, "`,`", table.Deleted, "` FROM `", table.Table, "` WHERE ", where, args)).StringMapResults()

	if table.Version != "" {
		// 增量更新
		serve.accessTableLock.Lock()
		mergeAccess(serve.accessTable, results, &table)
		serve.accessTableLock.Unlock()
	} else {
		// 全量更新
		accessTable := make(map[string]map[string]*accessInfo)
		mergeAccess(accessTable, results, &table)
		serve.accessTableLock.Lock()
		serve.accessTable = accessTable
		serve.accessTableLock.Unlock()
	}
}

func mergeAccess(to map[string]map[string]*accessInfo, from []map[string]string, table *TableAccess) {
	for _, r := range from {
		role := r[table.Role]
		module := r[table.Module]

		// 在增量更新模式下处理删除标记
		deleted := false
		if table.Deleted != "" {
			deleted = r[table.Deleted] == "1"
		}

		if to[role] == nil {
			to[role] = make(map[string]*accessInfo)
		}

		if deleted {
			// 新版本的数据中已经删除此条
			if to[role][module] != nil {
				delete(to[role], module)
			}
		} else {
			// 添加或覆盖数据
			access := r[table.Access]
			tags := r[table.Tag]
			info := accessInfo{access: map[string]bool{access: true}}
			if tags != "" {
				info.tags = strings.Split(tags, ",")
			}
			to[role][module] = &info
		}
	}
}
