// Command migrate 使用 golang-migrate 对 PostgreSQL 执行 migrations/ 下版本化 SQL。
//
// 前端类比：类似在部署前跑「数据库 schema 的 npm 脚本」——把 .up.sql / .down.sql 按版本顺序应用到库里，
// 保证本地、测试、生产库的表结构一致；只是这里用 Go 二进制 + DSN，而不是只在前端项目里跑。
//
// 工作目录需为 server/（与 go run ./cmd/api 一致），以便 configs 能加载 .env、.env.dev 与 configs 包里的默认路径。
//
// 用法：
//
//	go run ./cmd/migrate -op up
//	go run ./cmd/migrate -op down
//	go run ./cmd/migrate -op version
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"vehivle/configs"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func main() {
	// FE analogy: 类似 `process.argv` / yargs 解析 CLI 参数。
	// Go detail: flag 在 main 开头 Parse；指针是因为 flag 包用指针回填解析结果。
	op := flag.String("op", "up", "up | down | version")
	steps := flag.Int("steps", 1, "down 时回滚的迁移步数（仅 op=down 生效）")
	flag.Parse()

	// FE analogy: 像读取 `import.meta.env` / dotenv，得到数据库连接串等配置。
	// Go detail: 失败直接 Fatalf 退出进程；迁移工具通常不做 HTTP 重试，配置错了就应立刻失败。
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("configs: %v", err)
	}
	if cfg.Database.DSN == "" {
		log.Fatal("database DSN is empty (set VEHIVLE_DATABASE_DSN)")
	}

	// FE analogy: 类似 `process.cwd()`，用来拼出 migrations 文件夹的绝对路径。
	// Go detail: 迁移文件从磁盘读，路径依赖当前工作目录，所以文档要求从 server/ 目录执行。
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	migDir := filepath.Join(wd, "migrations")
	if st, err := os.Stat(migDir); err != nil || !st.IsDir() {
		log.Fatalf("migrations dir not found: %s", migDir)
	}

	// FE analogy: 像用 `fs.readFile` 读打包进内存的目录树，再交给迁移库；避免用 file:// URL 字符串。
	// Go detail: iofs + os.DirFS 把 migrations/ 当作 embed 风格的文件源；Windows 下 file:// 常被误解析，此写法更稳。
	src, err := iofs.New(os.DirFS(migDir), ".")
	if err != nil {
		log.Fatalf("migrate source: %v", err)
	}

	// FE analogy: 同时传入「SQL 文件从哪来」和「连哪个数据库」，类似数据源 + API baseURL 两处配置。
	// Go detail: NewWithSourceInstance 第一个参数是 source driver 名；DSN 是 postgres 驱动识别的连接串。
	m, err := migrate.NewWithSourceInstance("iofs", src, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("migrate new: %v", err)
	}
	// FE analogy: 像 `useEffect` 的 cleanup，进程退出前关掉迁移句柄。
	// Go detail: 忽略 Close 的返回值是常见写法；失败时进程已在退出路径上。
	defer func() { _, _ = m.Close() }()

	switch *op {
	case "up":
		// 执行所有未应用的 *.up.sql，直到最新版本。
		if err := m.Up(); err != nil {
			// FE analogy: 像 git「已经是最新」不是错误；这里同样用哨兵错误区分。
			// Go detail: errors.Is 用于判断是否是 migrate 包定义的 ErrNoChange。
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("migrate: already up to date")
				return
			}
			log.Fatalf("migrate up: %v", err)
		}
		fmt.Println("migrate: up ok")
	case "down":
		if *steps < 1 {
			log.Fatal("steps must be >= 1")
		}
		// FE analogy: 按「步数」撤销迁移；负数表示向下回滚。
		// Go detail: Steps(-N) 回滚 N 个版本，与 Down() 一次回滚到底不同。
		if err := m.Steps(-*steps); err != nil {
			log.Fatalf("migrate down: %v", err)
		}
		fmt.Println("migrate: down ok")
	case "version":
		// 打印当前 schema_migrations 表中的版本号，以及是否 dirty（中途失败会标记 dirty，需人工处理）。
		v, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				fmt.Println("migrate: version (no migration applied yet)")
				return
			}
			log.Fatalf("migrate version: %v", err)
		}
		fmt.Printf("migrate: version=%d dirty=%v\n", v, dirty)
	default:
		log.Fatal("op must be up | down | version")
	}
}
