package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type WebotsVersion struct {
	Major, Minor, Patch uint
}

var webotsVersionRx = regexp.MustCompile(`^([0-9]+)\.([0-9]+)\.([0-9]+)$`)

func ParseWebotsVersion(s string) (WebotsVersion, error) {
	m := webotsVersionRx.FindStringSubmatch(s)
	if m == nil {
		return WebotsVersion{}, fmt.Errorf("Invalid version syntax %s", s)
	}
	major, _ := strconv.ParseUint(m[1], 10, 0)
	minor, _ := strconv.ParseUint(m[2], 10, 0)
	patch, _ := strconv.ParseUint(m[3], 10, 0)
	return WebotsVersion{
		Major: uint(major),
		Minor: uint(minor),
		Patch: uint(patch),
	}, nil
}

func (v WebotsVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// A webots archive provie a lists of webots version and URL where to
// find them.
type WebotsArchive interface {
	AvailableVersions() []WebotsVersion
	GetUrl(WebotsVersion) (string, error)
}

type WebotsVersionList []WebotsVersion

func (l *WebotsVersionList) Len() int {
	return len(*l)
}

func (l *WebotsVersionList) Less(i, j int) bool {
	if (*l)[i].Major < (*l)[j].Major {
		return true
	} else if (*l)[i].Major > (*l)[j].Major {
		return false
	}
	if (*l)[i].Minor < (*l)[j].Minor {
		return true
	} else if (*l)[i].Minor > (*l)[j].Minor {
		return false
	}
	return (*l)[i].Patch < (*l)[j].Patch
}

func (l *WebotsVersionList) Swap(i, j int) {
	tmp := (*l)[i]
	(*l)[i] = (*l)[j]
	(*l)[j] = tmp
}

type HttpWebotsArchive struct {
	baseurl  string
	arch     string
	versions WebotsVersionList
}

func NewWebotsHttpArchive(basepath string) (*HttpWebotsArchive, error) {
	res := &HttpWebotsArchive{}
	var err error
	res.baseurl, err = res.osPath(basepath)
	if err != nil {
		return nil, err
	}
	res.arch, err = res.archSuffix()
	if err != nil {
		return nil, err
	}

	err = res.load()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (a *HttpWebotsArchive) osPath(basepath string) (string, error) {
	//get the right suffix
	var suffix string
	switch runtime.GOOS {
	case "darwin":
		suffix = "mac"
	case "windows", "linux":
		suffix = runtime.GOOS
	}
	if len(suffix) == 0 {
		return "", fmt.Errorf("Unsupported system %s", runtime.GOOS)
	}

	if strings.HasSuffix(basepath, "/") {
		return basepath + suffix, nil
	}

	return basepath + "/" + suffix, nil
}

func (a *HttpWebotsArchive) archSuffix() (string, error) {
	if runtime.GOOS != "linux" {
		return "", fmt.Errorf("%s is not supported yet", runtime.GOOS)
	}

	switch runtime.GOARCH {
	case "386":
		return "i386", nil
	case "amd64":
		return "x86-64", nil
	}

	return "", fmt.Errorf("Unsupported architecture %s for webots", runtime.GOARCH)
}

func (a *HttpWebotsArchive) fileSuffix() string {
	return fmt.Sprintf("-%s.tar.bz2", a.arch)
}

func (a *HttpWebotsArchive) load() error {
	resp, err := http.Get(a.baseurl)
	if err != nil {
		return err
	}

	tokenizer := html.NewTokenizer(resp.Body)

	nameRx := regexp.MustCompile(fmt.Sprintf(`^webots-(.*)-%s.tar.bz2$`, a.arch))

	for {
		t := tokenizer.Next()
		if t == html.ErrorToken {
			err := tokenizer.Err()
			if err == io.EOF {
				break
			}
			return err

		}

		if t != html.StartTagToken {
			continue
		}

		tName, hasAttrib := tokenizer.TagName()
		if string(tName) != "a" {
			continue
		}

		if hasAttrib == false {
			continue
		}

		stopped := false
		for stopped == false {
			key, val, next := tokenizer.TagAttr()
			if string(key) != "href" {
				continue
			}
			stopped = !next
			// we got a link, test if it has the right prefix
			matches := nameRx.FindStringSubmatch(string(val))
			if matches == nil {
				continue
			}

			v, err := ParseWebotsVersion(matches[1])
			if err != nil {
				return err
			}

			a.versions = append(a.versions, v)
		}
	}

	sort.Sort(&a.versions)

	return nil
}

func (a *HttpWebotsArchive) AvailableVersions() []WebotsVersion {
	return []WebotsVersion(a.versions)
}

func (a *HttpWebotsArchive) GetUrl(v WebotsVersion) (string, error) {
	for _, vv := range a.versions {
		if v != vv {
			continue
		}

		return fmt.Sprintf("%s/webots-%s-%s.tar.bz2", a.baseurl, v, a.arch), nil
	}
	return "", fmt.Errorf("Version %s not found", v)
}
