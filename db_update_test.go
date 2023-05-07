package tconfig

import (
	"embed"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/tartok/tlog"
	"os"
	"testing"
)

var (
	//go:embed test_data/db_scripts
	db_scripts embed.FS
)

func TestDbCreate(t *testing.T) {
	tlog.InitLog("", os.Stdout)
	tlog.InitErr("", os.Stdout)
	conf, err := Load("./test_data/config.json")
	if err != nil {
		panic(err)
	}
	err = PgUpdate(conf.Db, db_scripts, "test_data/db_scripts", &tlog.DefLoggers)
	fmt.Println(conf, err)
}
