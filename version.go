package tconfig

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	Major   int
	Minor   int
	Release int
	Build   int
}

func NewVersion(s string) (Version, error) {
	v := Version{}
	var err error
	ss := strings.Split(s, ".")
	v.Major, err = strconv.Atoi(ss[0])
	if err != nil {
		return Version{}, err
	}
	v.Minor, err = strconv.Atoi(ss[1])
	if err != nil {
		return Version{}, err
	}
	v.Release, err = strconv.Atoi(ss[2])
	if err != nil {
		return Version{}, err
	}
	v.Build, err = strconv.Atoi(ss[3])
	if err != nil {
		return Version{}, err
	}
	return v, nil
}
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Release, v.Build)
}
func (v Version) Less(version Version) bool {
	if v.Major < version.Major {
		return true
	}
	if v.Minor < version.Minor {
		return true
	}
	if v.Release < version.Release {
		return true
	}
	if v.Build < version.Build {
		return true
	}
	return false
}
