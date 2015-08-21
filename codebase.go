// Copyright © 2015 Erik Brady <brady@dvln.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package codebase works with 'dvln' code bases.  This typically means
// getting, pulling/updating, reading and writingcodebase definitions or
// handling situations where one is defined dynamically.  Currently this
// is focused on a JSON definition for the codebase file.
package codebase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/dvln/mapstructure"
	"github.com/dvln/out"
	"github.com/dvln/pkg"
	"github.com/dvln/util/file"
	globs "github.com/dvln/viper"
)

// Defn is the code base defintion: what pkgs are available to this codebase
// and what, if any, configuration/description/etc information do we have
// related to this code base.
type Defn struct {
	Name     string                       `json:"name"`
	Desc     string                       `json:"desc"`
	Contacts []string                     `json:"contacts.omitempty"`
	Attrs    map[string]string            `json:"attrs,omitempty"`
	Vars     map[string]string            `json:"vars,omitempty"`
	Pathing  map[string]string            `json:"pathing,omitempty"`
	Access   map[string]map[string]string `json:"access,omitempty"`
	Pkgs     []pkg.Defn                   `json:"pkgs" mapstructure:",squash"`
}

// Locality indicates to the codebase existence checker if a pkg/repo exists
// on the local host (in any expected area) if it is 'Locally" and if it's
// in the "cloud"
type Locality int

const (
	// LocalPath indicates a pkg (repo) is available via locally accessible path
	LocalPath Locality = 1 << iota
	// Remote indicates no local path found, but available via remote/network
	Remote
	// NonExistent indicates we couldn't find the repo, sorry
	NonExistent
	// Anywhere "combines" local and remote and just indicates the repo exists
	Anywhere = LocalPath | Remote
)

/* FIXME: erik: mucking around with ideas here
type CodeBase interface {
    Exists(name string, locality Locality) (string, locality, error)
    Get(name string, dest string) error
    Read(r io.Reader) error
    Write(w io.Writer) error
    //PkgExists(pkgSel string) (*Pkg, error)
    //Pkg(pkgSel string) (*pkg.Pkg, error)
    //AddPkg(p *pkg.Pkg) error
    //RmPkg(p *pkg.Pkg) error
    //Pkgs(pkgSel ...string) []*pkg.Pkg
    //Item(item CodebaseItem) error
    //SetItem(item CodebaseItem) error
}
*/

/* FIXME: erik: consider a way to get at fields generically, perhaps:
type CodebaseItem interface {
    UnmarshalCodebase(cbDefn *Defn) error
}

func (c *Defn) Item(i CodebaseItem) error {
    return c.UnmarshalCodebase(i)
}
*/

// Read will, given an io.Reader, attempt to scan in the codebase contents
// and "fill out" the given Defn structure for you.  What could
// go wrong?  If anything a non-nil error is returned.
func (cb *Defn) Read(r io.Reader) error {
	codebaseMap := make(map[string]interface{})
	decJSON := json.NewDecoder(r)
	err := decJSON.Decode(&codebaseMap)
	if err != nil {
		return out.WrapErr(err, "Failed to decode codebase JSON file", 3001)
	}

	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           cb,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return out.WrapErr(err, "Failed to prepare codebase file decoder", 3002)
	}

	err = decoder.Decode(codebaseMap)
	if err != nil {
		return out.WrapErr(err, "Failed to decode codebase file contents", 3003)
	}
	return nil
}

// Exists scans to see if the given codebase exists, and one can
// narrow that using Locality (LocalPath, Remote or Anywhere) and
// if it exists you will be told where (path or URL) and if that
// is a local host accessible path or remote (via a returned locality)
// and if there was some error checking existence (if no error and
// doesn't exist then the where it exists will be "" and locality
// will be the NonExistent constant)
func (cb *Defn) Exists(codebaseSel string, locality Locality) (string, Locality, error) {
	var err error
	exists := false
	if codebaseSel != "" {
		exists, err = file.Exists(codebaseSel)
	}
	ws_root_dir := ""
	if !exists {
		ws_root_dir = globs.GetString("wsRootDir")
	}
	if ws_root_dir == "" {
		ws_root_dir = globs.GetString("dvlnCfgDir")
	}
	if codebaseSel == "" || !exists || err != nil {
		if err == nil {
			out.Debugln("No codebase file to read, skipping (considered normal)")
		} else {
			out.Debugf("No codebase file to read, skipping (abnormal, unexpected err: %s)\n", err)
		}
		return "", NonExistent, err
	}
	return codebaseSel, LocalPath, nil
}

// FindAndRead does just that, it finds the codebase related to the given
// name (can be a name, URL or any valid CodeBase selector) and it will
// grab it if it is not local already (assuming it exists) and it will
// read it into the codebase data structure.
func FindAndRead(codebaseSel string) (*Defn, error) {
	cb := &Defn{}
	//eriknow: normally we would do any smart discovery of the code base
	//         definition file here via 'findCodebase()' or something which
	//         would be able to get it via local file (support RCS versioned),
	//         or, more likely, the file is in a git clone with the codebase
	//         definition in it which may also have devline definitions and
	//         possibly code as well
	//         - the name could be a partial or full URL, eg:
	//           "dvln" or "github.com/dvln/dvln" or "http://github.com/dvln/dvln"
	//           or "file:://path/to/somefile.json" or "/path/to/somefile.json" or
	//           might use OS friendly paths like "\path\to\somefile.toml"
	//         - if we think it's in a VCS we need to bring that clone into the
	//           workspace if we have a workspace, if we don't have a workspace
	//           we should bring it into a temp location or into ~/.dvlncfg/codebase/
	//           or something like that (flattened full path?)... we can "smart local clone"
	//           this pkg into the workspace with hard links or whatever if later "get" of it

	// normally Exists() would do this part and try and get us a "real" name for the
	// codebase (full URL/etc... but the name should be simple in the file even if
	// the "full" name is a URL and such)
	fileName := fmt.Sprintf("/Users/brady/.dvlncfg/%s.codebase", codebaseSel)
	cbFile, locality, err := cb.Exists(fileName, LocalPath)
	if locality == NonExistent {
		cb = &Defn{Name: "generated", Desc: "Dynamically generated development line"}
		return cb, err // note, err may be nil as it's ok for codebase not to exist
	}
	//FIXME: erik: normally we would check and see if locality was 'Remote' as well
	//       and, if so, use the Get() routine to bring it down to our workspace
	//       if we have one and to a tmp location if not (wsroot should be set
	//       in viper if we have a workspace, note that if it's nested it may
	//       be a child workspace root but that should be normal I think, consider)
	fileContents, err := ioutil.ReadFile(cbFile)
	if err != nil {
		msg := fmt.Sprintf("Codebase file \"%s\" read failed\n", cbFile)
		return cb, out.WrapErr(err, msg, 3000)
	}
	out.Tracef("---- Codebase file '%s' contents ----\n%s---- END ----\n", cbFile, fileContents)
	err = cb.Read(bytes.NewReader(fileContents))
	return cb, err
}
