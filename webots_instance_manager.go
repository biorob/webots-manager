package main

import (
	"archive/tar"
	"bufio"
	"compress/bzip2"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/nightlyone/lockfile"
)

type WebotsInstanceManager interface {
	Install(WebotsVersion) error
	Use(WebotsVersion) error
	IsUsed(WebotsVersion) bool
	Installed() []WebotsVersion
	ApplyAllTemplates() error
}

type SymlinkWebotsManager struct {
	basepath    string
	workpath    string
	installpath string
	usedpath    string
	lock        lockfile.Lockfile
	installed   WebotsVersionList
	inUse       *WebotsVersion
	archive     WebotsArchive
	templates   TemplateManager
	gid         int
}

func NewSymlinkManager(a WebotsArchive) (*SymlinkWebotsManager, error) {
	var err error
	res := &SymlinkWebotsManager{
		archive: a,
	}

	res.basepath, res.workpath, res.installpath, err = symlinkManagerPath()
	if err != nil {
		return nil, err
	}

	res.templates, err = NewHasHTemplateManager(path.Join(res.workpath, "templates"))
	if err != nil {
		return nil, err
	}

	res.usedpath = path.Join(res.workpath, "used")
	res.lock, err = lockfile.New(path.Join(res.workpath, "global.lock"))
	if err != nil {
		return nil, err
	}

	err = res.listInstalled()
	if err != nil {
		return nil, err
	}
	err = res.listUsed()
	if err != nil {
		return nil, err
	}

	//checks that we have the right gid
	res.gid, err = getGid("webots-manager")
	if err != nil {
		return nil, err
	}

	found := false
	userGroups, err := os.Getgroups()
	if err != nil {
		return nil, err
	}

	for _, g := range userGroups {
		if g == res.gid {
			found = true
			break
		}
	}

	if found == false {
		return nil, fmt.Errorf("Current use is not in 'webots-manager' group")
	}

	webotsHome := os.Getenv("WEBOTS_HOME")
	if len(webotsHome) == 0 {
		fmt.Printf("WEBOTS_HOME is not set, please consider exporting WEBOTS_HOME=%s", res.installpath)
	} else if webotsHome != res.installpath {
		return nil, fmt.Errorf("Invalid WEBOTS_HOME=%s, please use WEBOTS_HOME=%s", webotsHome, res.installpath)
	}

	return res, nil
}

func symlinkManagerPath() (string, string, string, error) {
	if runtime.GOOS != "linux" {
		return "", "", "", fmt.Errorf("Only linux is supported yet")
	}
	basepath := "/usr/local"
	workpath := path.Join(basepath, "webots_manager")
	installpath := path.Join(basepath, "webots")
	return basepath, workpath, installpath, nil

}

func (i *SymlinkWebotsManager) tryLock() error {
	if err := i.lock.TryLock(); err != nil {
		return fmt.Errorf("Could not lock %s: %s", i.lock, err)
	}
	return nil
}

func (i *SymlinkWebotsManager) unlock() {
	if err := i.lock.Unlock(); err != nil {
		panic(err)
	}
}

func (i *SymlinkWebotsManager) listInstalled() error {
	if err := i.tryLock(); err != nil {
		return err
	}
	defer i.unlock()

	files, err := ioutil.ReadDir(i.workpath)
	if err != nil {
		return err
	}

	i.installed = make(WebotsVersionList, 0, len(files))
	for _, fi := range files {
		if fi.IsDir() == false {
			continue
		}
		v, err := ParseWebotsVersion(fi.Name())
		if err != nil {
			continue
		}
		i.installed = append(i.installed, v)
	}

	sort.Sort(&i.installed)
	return nil
}

func (i *SymlinkWebotsManager) listUsed() error {
	fi, err := os.Stat(i.usedpath)
	if err != nil && os.IsNotExist(err) == false {
		return err
	}
	if os.IsNotExist(err) {
		i.inUse = nil
		return nil
	}

	if (fi.Mode() & os.ModeSymlink) == os.ModeSymlink {
		return fmt.Errorf("Set-up error, %s exist but it is not a symlink", i.usedpath)
	}

	dest, err := os.Readlink(i.usedpath)
	if err != nil {
		return err
	}

	v, err := ParseWebotsVersion(dest)
	if err != nil {
		return fmt.Errorf("Invalid %s target %s", i.installpath, dest)
	}
	found := false
	for _, vv := range i.installed {
		if v == vv {
			found = true
			break
		}
	}
	if found == false {
		return fmt.Errorf("Internal error, %s in use, but not listed as installed %v", v, i.installed)
	}
	i.inUse = &v
	return nil
}

type ProgressReader struct {
	reader   io.Reader
	progress chan int64
}

func (r *ProgressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.progress <- int64(n)
	}
	return n, err
}

func (m *SymlinkWebotsManager) extractFile(v WebotsVersion, h *tar.Header, r io.Reader) error {

	dest := strings.TrimPrefix(h.Name, "webots/")
	dest = path.Join(m.workpath, v.String(), dest)
	if dest == path.Join(m.workpath, v.String()) {
		return nil
	}
	switch h.Typeflag {
	case tar.TypeReg, tar.TypeRegA:
		destDir := path.Dir(dest)
		err := os.MkdirAll(destDir, 0775)
		if err != nil {
			return err
		}
		f, err := os.Create(dest)
		if err != nil {
			return err
		}
		_, err = io.Copy(f, r)
		if err != nil {
			return err
		}
	case tar.TypeDir:
		err := os.MkdirAll(dest, 0775)
		if err != nil {
			return err
		}
	case tar.TypeSymlink:
		destDir := path.Dir(dest)
		err := os.MkdirAll(destDir, 0775)
		if err != nil {
			return err
		}
		err = os.Symlink(h.Linkname, dest)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("Internal error, cannot handle file %s", h.Name)
	}

	err := os.Chtimes(dest, time.Now(), h.ModTime)
	if err != nil {
		return err
	}

	return nil
}

func (m *SymlinkWebotsManager) uncompressFromHttp(v WebotsVersion, addr string) error {
	dest := path.Join(m.workpath, v.String())
	err := os.RemoveAll(dest)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dest, 0775|os.ModeSetgid)
	if err != nil {
		return err
	}

	resp, err := http.Get(addr)
	if err != nil {
		return err
	}

	var netReader io.Reader = resp.Body
	if resp.ContentLength >= 0 {
		//adds a progress bar
		progress := make(chan int64)
		go func() {
			bar := pb.New64(resp.ContentLength)
			bar.Format("[=>_]")
			bar.Start()
			for n := range progress {
				bar.Add64(n)
			}
			bar.FinishPrint(fmt.Sprintf("Downloaded and extracted %s", addr))
		}()
		netReader = &ProgressReader{
			reader:   resp.Body,
			progress: progress,
		}
		defer close(progress)
	}

	tarReader := tar.NewReader(bzip2.NewReader(netReader))
	for {
		fileHeader, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		err = m.extractFile(v, fileHeader, tarReader)
		if err != nil {
			return fmt.Errorf("Cannot extract %s: %s", fileHeader.Name, err)
		}
	}
	return nil
}

func (m *SymlinkWebotsManager) install(v WebotsVersion) error {
	address, err := m.archive.GetUrl(v)
	if err != nil {
		return err
	}
	log.Printf("Downloading from %s", address)
	err = m.uncompressFromHttp(v, address)
	if err != nil {
		return err
	}

	found := false
	for _, vv := range m.installed {
		if vv == v {
			found = true
			break
		}
	}

	log.Printf("Installing templates for %s", v)
	err = m.templates.ApplyTemplates(path.Join(m.workpath, v.String()), v)
	if err != nil {
		return err
	}
	if found == false {
		log.Printf("Successfuly installed %s", v)
		m.installed = append(m.installed, v)
		sort.Sort(&m.installed)
	} else {
		log.Printf("Successfuly re-installed %s", v)
	}

	return nil
}

func (m *SymlinkWebotsManager) Install(v WebotsVersion) error {
	if err := m.tryLock(); err != nil {
		return err
	}
	defer m.unlock()
	return m.install(v)
}

func (m *SymlinkWebotsManager) Installed() []WebotsVersion {
	return []WebotsVersion(m.installed)
}

func (m *SymlinkWebotsManager) IsUsed(v WebotsVersion) bool {
	if m.inUse == nil {
		return false
	}
	return m.inUse.Major == v.Major && m.inUse.Minor == v.Minor && m.inUse.Patch == v.Patch
}

func getGid(g string) (int, error) {
	f, err := os.Open("/etc/group")
	if err != nil {
		return -1, err
	}
	r := bufio.NewReader(f)
	stopped := false
	groupRx := regexp.MustCompile(`^([a-z\-_A-Z0-9]+):[^:]*:([0-9]+):`)
	for stopped == false {
		l, err := r.ReadString('\n')
		if err == io.EOF {
			stopped = true
		} else if err != nil {
			return -1, err
		}
		m := groupRx.FindStringSubmatch(l)
		if m == nil {
			return -1, fmt.Errorf("Invalid line %s", l)
		}
		if m[1] == g {
			gid, err := strconv.ParseInt(m[2], 10, 0)
			if err != nil {
				return -1, err
			}
			return int(gid), nil
		}
	}
	return -1, nil
}

func SymlinkManagerSystemInit() error {
	if os.Getuid() != 0 || os.Geteuid() != 0 {
		return fmt.Errorf("need to be root")
	}

	gid, err := getGid("webots-manager")
	if err != nil {
		return fmt.Errorf("Could not check exitance of group webots-manager: %s", err)
	}
	if gid == -1 {
		cmd := exec.Command("addgroup", "webots-manager")
		err := cmd.Run()
		if err != nil {
			return err
		}
		fmt.Println("Added system group 'webots-manager'")
		gid, err = getGid("webots-manager")
		if err != nil {
			return err
		}
		if gid == -1 {
			return fmt.Errorf("Internal error, gid should not be -1")
		}
	}

	_, workpath, installpath, err := symlinkManagerPath()
	if err != nil {
		return err
	}

	err = os.MkdirAll(workpath, 0775|os.ModeSetgid)
	if err != nil {
		return err
	}

	err = os.Chown(workpath, 0, gid)
	if err != nil {
		return err
	}

	err = os.Chmod(workpath, 0775|os.ModeSetgid)
	if err != nil {
		return err
	}

	fi, err := os.Lstat(installpath)
	if err != nil && os.IsNotExist(err) == false {
		return err
	}
	usedpath := path.Join(workpath, "used")
	if os.IsNotExist(err) {
		err := os.Symlink(usedpath, installpath)
		if err != nil {
			return err
		}
	} else {
		if (fi.Mode() & os.ModeSymlink) != os.ModeSymlink {
			return fmt.Errorf("Webots seems already installed in %s, you should remove it.", installpath)
		}

		dest, err := os.Readlink(installpath)
		if err != nil {
			return err
		}
		if dest != path.Join(workpath, "used") {
			return fmt.Errorf("%s is a symling to incorrect target %s (require %s)", installpath, dest, usedpath)
		}
	}

	return nil
}

func (m *SymlinkWebotsManager) Use(v WebotsVersion) error {
	if err := m.tryLock(); err != nil {
		return err
	}
	defer m.unlock()

	found := false
	for _, vv := range m.installed {
		if vv == v {
			found = true
			break
		}
	}
	if found == false {
		log.Printf("Installing missing version %s", v)
		err := m.install(v)
		if err != nil {
			return err
		}
	}

	err := os.RemoveAll(m.usedpath)
	if err != nil {
		return err
	}
	err = os.Symlink(v.String(), m.usedpath)
	if err != nil {
		return err
	}

	m.inUse = &v

	return nil
}

func (m *SymlinkWebotsManager) ApplyAllTemplates() error {
	for _, v := range m.installed {
		err := m.templates.ApplyTemplates(path.Join(m.workpath, v.String()), v)
		if err != nil {
			return err
		}
	}
	return nil
}
