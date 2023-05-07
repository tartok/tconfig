package tconfig

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/tartok/tlog"
	"io"
	"io/fs"
	"sort"
	"strings"
)

//type UpdateConfig struct {
//	Scripts string   `json:"scripts"`
//	Files   []string `json:"files"`
//}

type Db struct {
	Driver         string `json:"driver"`
	Host           string `json:"host"`
	Port           int    `json:"port"`
	User           string `json:"user"`
	Password       string `json:"password"`
	DbName         string `json:"dbname"`
	DropDataBase   bool   `json:"drop_data_base"`
	CreateDataBase bool   `json:"create_data_base"`
	UpdateDataBase bool   `json:"update_data_base"`
	//DbUpdater      *UpdateConfig
}

func (d Db) ConnectString(dbName string) string {
	if dbName == "" {
		dbName = d.DbName
	}
	return fmt.Sprintf("host=%s port=%d user=%s password='%s' dbname=%s sslmode=disable",
		d.Host, d.Port, d.User, d.Password, dbName)
}

func PgUpdate(conf *Db, data fs.ReadDirFS, dataPatch string, logs *tlog.Loggers) error {
	do := func(dbName string, f func(db *sql.DB) (string, error)) error {
		conn, err := sql.Open("postgres", conf.ConnectString(dbName))
		if err != nil {
			if logs != nil && logs.Err != nil {
				logs.Err.Println(err)
			}
			return err
		}
		defer conn.Close()
		l, err := f(conn)
		if err != nil {
			if logs != nil && logs.Err != nil {
				logs.Err.Println(err)
			}
			return err
		}
		if l != "" && logs != nil && logs.Log != nil {
			logs.Log.Println(l)
		}
		return nil
	}

	if conf.DropDataBase {
		do("postgres", func(db *sql.DB) (string, error) {
			_, err := db.Exec(fmt.Sprintf(`DROP DATABASE "%s"`, conf.DbName))
			return fmt.Sprintf("db: %s dropped", conf.DbName), err
		})
	}
	if conf.CreateDataBase {
		do("postgres", func(db *sql.DB) (string, error) {
			_, err := db.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, conf.DbName))
			return fmt.Sprintf("db: %s created", conf.DbName), err
		})
	}
	if conf.UpdateDataBase {
		var uConfig struct {
			ConfigTable string   `json:"configTable"`
			Files       []string `json:"files"`
		}
		c, err := data.Open(dataPatch + "/config.json")
		if err != nil {
			panic(err)
		}
		defer c.Close()
		r := json.NewDecoder(c)
		err = r.Decode(&uConfig)
		if err != nil {
			return err
		}
		var vers, currentVersion Version
		do("", func(db *sql.DB) (string, error) {
			var s string
			err := db.QueryRow(fmt.Sprintf("select version from %s", uConfig.ConfigTable)).Scan(&s)
			if err == nil {
				currentVersion, _ = NewVersion(s)
			}
			return fmt.Sprintf("current db verstion %v", currentVersion), nil
		})
		vers = currentVersion
		defer func() {
			if currentVersion != vers {
				if logs != nil && logs.Log != nil {
					logs.Log.Printf("current db verstion %v", vers)
				}
			}
		}()
		de, err := data.ReadDir(dataPatch)
		if err != nil {
			return err
		}
		sort.Slice(de, func(i, j int) bool {
			v1, _ := NewVersion(de[i].Name())
			v2, _ := NewVersion(de[j].Name())
			return v1.Less(v2)
		})
		for _, entry := range de {
			if entry.IsDir() {
				newVers, _ := NewVersion(entry.Name())
				if !vers.Less(newVers) {
					continue
				}
				if logs != nil && logs.Log != nil {
					logs.Log.Printf("try v:%s", newVers)
				}
				err := do("", func(db *sql.DB) (string, error) {
					tx, err := db.Begin()
					if err != nil {
						return "", err
					}
					for _, fileName := range uConfig.Files {
						f, err := data.Open(strings.Join([]string{dataPatch, entry.Name(), fileName}, "/"))
						if err != nil {
							continue
						}
						err = func() error {
							defer f.Close()
							s, _ := io.ReadAll(f)
							_, err := tx.Exec(string(s))
							return err
						}()
						if err != nil {
							tx.Rollback()
							return "", err
						}
					}
					_, err = tx.Exec(fmt.Sprintf("update %s set version=$1", uConfig.ConfigTable), newVers.String())
					if err != nil {
						tx.Rollback()
						return "", err
					}
					err = tx.Commit()
					if err != nil {
						return "", err
					}
					vers = newVers
					return "", nil
				})
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
