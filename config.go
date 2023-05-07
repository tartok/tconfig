package tconfig

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
)

type (
	Log struct {
		ToFile bool `json:"to_file"`
		Debug  bool `json:"debug"`
		Log    bool `json:"log"`
	}

	App struct {
		Db  *Db
		Api struct {
			Listen string `json:"listen"`
		} `json:"api"`
		Log *Log `json:"log"`
	}
)

func Load(fileName string) (*App, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	var d []byte
	for {
		line, _, err := r.ReadLine()
		if len(line) == 0 && err != nil {
			break
		}
		if !bytes.HasPrefix(bytes.Trim(line, " "), []byte("//")) {
			d = append(d, line...)
		}
	}
	if err != nil {
		panic(err)
	}
	c := App{}
	err = json.Unmarshal(d, &c)
	if err != nil {
		return nil, err
	}
	return &c, err
}
