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
	"text/template"

	"github.com/dvln/mapstructure"
	"github.com/dvln/out"
	"github.com/dvln/pkg"
	"github.com/dvln/util/file"
	globs "github.com/dvln/viper"
)

// Defn is the code base defintion file: what known pkgs are available to
// this codebase at startup time and what, if any, detailed config/description
// and other info do we have on this codebase and constituent packages.
type Defn struct {
	Name     string                       `json:"name"`
	Desc     string                       `json:"desc"`
    Pkg      pkg.Revision
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
	// LocalDir for pkg repo available via local host accessible dir
	LocalDir Locality = 1 << iota
	// RemoteURL indicates we're identifying the repo via a network URL
	RemoteURL
	// NonExistent indicates we couldn't find the repo, sorry
	NonExistent
	// Anywhere "combines" local and remote and just indicates the repo exists
	Anywhere = LocalDir | RemoteURL
)

// Codebase is an interface for working with 'dvln' codebases, not that it
// is of much use today it may help with testing and other fun over time.
// maybe ReadWriteGetter would be a better name (?)... or "Manipulater"
// - currently not used, playing with ideas
type Codebase interface {
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

/* FIXME: erik: consider a way to get at fields generically, perhaps:
type CodebaseItem interface {
    UnmarshalCodebase(cbDefn *Defn) error
}

func (c *Defn) Item(i CodebaseItem) error {
    return c.UnmarshalCodebase(i)
}
*/

// Exists checks for codebase existence by using the basic pkg rev Exists()
// routine to see if the specified codebase (at any specified rev, or from
// the default VCS rev) can be found.  If so it'll return the fullest URI
// it can to that codebase pkg repo, otherwise an error is returned.
// Config file and env var settings can help to find the codebase, ie:
// * DVLN_CODEBASE_PATH in env or codebase_path in cfg file
// -  > space separated list of paths to prepend to the codebase name
// -  > eg: "github.com/dvln git+ssh://host.com/path/to/clones /some/local/dir"
// -     - note: 'hub' and 'hub:<uri>' reserved for future user/central dvln hub
// -  > all values are *always* forward slash regardless of platform
// -  > 'dvln' as the codebase would result in looking, in order, here:
// -     > "github.com/dvln/dvln[.git]"
// -     > "git+ssh://host.com/clone/path/dvln[.git]"
// -     > "/some/local/dir/dvln[.git]"
// -     > "dvln"
// -  > final fallback will be on the individual dvln itself
// Based on the above existence checks it'll return:
// - string: URI: full path (on filesystem) or URL (if remote), "" if not found
// - locality: where it exists (LocalDir, RemoteURL, NonExistent)
// - error: any error that is detected in scanning for the workspace
func (cb *Defn) Exists(codebaseVerSel string) (string, Locality, error) {
	//eriknow, make some choices around codebase existence checks...
	//         should be a challenge, that's for sure
	var err error
	exists := false
	if codebaseVerSel != "" {
		exists, err = file.Exists(codebaseVerSel)
	}
	wkspcRootDir := ""
	if !exists {
		wkspcRootDir = globs.GetString("wkspcRootDir")
	}
	if wkspcRootDir == "" {
		wkspcRootDir = globs.GetString("dvlnCfgDir")
	}
	if codebaseVerSel == "" || !exists || err != nil {
		if err == nil {
			out.Debugln("No codebase file to read, skipping (considered normal)")
		} else {
			out.Debugf("No codebase file to read, skipping (abnormal, unexpected err: %s)\n", err)
		}
		return "", NonExistent, err
	}
	return codebaseVerSel, LocalDir, nil
}

func (cb *Defn) applyVarsToField(desc, fieldName, fieldValue string) (string, error) {
	t := template.New(fieldName)
	t, err := t.Parse(fieldValue)
	if err != nil {
		return "", out.WrapErrf(err, 3004, "Parsing problem applying templates to:\n%v\n  URI Template: %v", desc, fieldValue)
	}
	buf := new(bytes.Buffer)
	err = t.Execute(buf, cb.Vars)
	if err != nil {
		return "", out.WrapErrf(err, 3005, "Template execute problem with codebase repo definition\n%v\n  URI Template: %v", desc, fieldValue)
	}
	result := fmt.Sprintf("%s", buf)
	return result, nil
}

// expandVarUse basically takes any variables (Vars) defined in the
// codebase definition file in a codebase wide section like this:
//   ...
//   "vars" : {
//     "dvln": "http://github.com/dvln",
//     "joe": "http://github.com/joe",
//     "spf13": "http://github.com/spf13"
//   },
//   ...
// So that anywhere we see "{{<.varname>/etc}}" so that something like
// these (defined for a single package in the codebase):
//       "repo" : { "rw": "{{.dvln}}/viper" },
//       "remotes" : {
//         "vendor": { "r": "{{.spf13}}/viper" },
//         "joe": { "read": "{{.joe}}/viper" }
//       },
// Would turn into:
//       "repo" : { "rw": "http://github.com/dvln/viper" },
//       "remotes" : {
//         "vendor": { "r": "http://github.com/spf13/viper" },
//         "joe": { "read": "http://github.com/joe/viper" }
//       },
// The other codebase level access would turn from this:
//   "access" : {
//     "m,^{{.dvln}}/*, && Vendor!=True": {
//       "read" : "open",
//       "write": "http://dvln.org/api/v1/access?codebase={{.TopCodebase}}&pkg={{.Pkg}}&branch={{.Branch}}&devline={{.Devline}}&user={{.UserID}}&type=write"
//     }
//   },
// Into this (only focuses on the condiitonal part):
//   "access" : {
//     "m,^http://github.com/dvln/*, && Vendor!=True": {
//       "read" : "open",
//       "write": "http://dvln.org/api/v1/access?codebase={{.TopCodebase}}&pkg={{.Pkg}}&branch={{.Branch}}&devline={{.Devline}}&user={{.UserID}}&type=write"
//     }
//   },
// Right now it's kind of cheesy and only adjusts "access" (codebase-wide),
// remotes" and "repo" related codebase fields (currently only place that
// var's can be used):
func (cb *Defn) expandVarUse() error {
	// Deal with the codebase level access conditional here, expanding any
	// vars used in templates there
	for conditional, accessMap := range cb.Access {
		desc := fmt.Sprintf("  Access Conditional for Codebase: %s\n", cb.Name)
		result, err := cb.applyVarsToField(desc, "codebaseAccess", conditional)
		if err != nil {
			return err
		}
		if result != conditional {
			cb.Access[result] = accessMap
			delete(cb.Access, conditional)
		}
	}

	// Deal with package settings that can use vars in templates here:
	for _, pkg := range cb.Pkgs {
		// first deal with any repo settings using vars
		for access, repoURI := range pkg.Repo {
			desc := fmt.Sprintf("  Pkg: %s\n  Tgt: %s", pkg.Name, access)
			result, err := cb.applyVarsToField(desc, "repoURI", repoURI)
			if err != nil {
				return err
			}
			pkg.Repo[access] = result
		}
		// then deal with any remotes definitions set up using vars
		for remName, remURLMap := range pkg.Remotes {
			for access, repoURI := range remURLMap {
				desc := fmt.Sprintf("  Pkg: %s\n  Remote: %s\nTgt: %s", pkg.Name, remName, access)
				result, err := cb.applyVarsToField(desc, "remoteURI", repoURI)
				if err != nil {
					return err
				}
				remURLMap[access] = result
			}
		}
	}
	return nil
}

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
	// codebase definitions can be reduced in size if the person defining that
	// file uses variables to identify common repo references and such, lets
	// examine those vars and, if any, make sure we "expand" them so the codebase
	// definition is complete.  Write() will "smart" subtitute the vars back.
	if err = cb.expandVarUse(); err != nil {
		return err
	}
	return nil
}

// Write will, given an io.Writer, attempt to write the codebase out to
// it's local file representation, note that it will take any settings
// matching a "Var" and "re-compact" it so the "{{.<var>}}" is used for
// any exact matches in fields that support var expansion.
func (cb *Defn) Write(r io.Writer) error {
	//FIXME: erik: make this thing work
	return nil
}

// New returns a pointer to a codebase definition, empty at this point, see Get
// and/or Read
func New() *Defn {
	cb := &Defn{}
	return cb
}

// Get basically tries to find, get and read a codebase if it can, it
// returns a codebase definition and any error's that may have occurred
func (cb *Defn) Get(codebaseVerSel string) error {
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
	fileName := fmt.Sprintf("/Users/brady/.dvlncfg/%s.codebase", codebaseVerSel)
	cbFile, locality, err := cb.Exists(fileName)
	if locality == NonExistent {
		cb.Name = "generated"
		cb.Desc = "Dynamically generated development line"
		return err // note, err may be nil as it's ok for codebase not to exist
	}
	//FIXME: erik: normally we would check and see if locality was 'RemoteURL' as well
	//       and, if so, use the Get() routine to bring it down to our workspace
	//       if we have one and to a tmp location if not (wsroot should be set
	//       in viper if we have a workspace, note that if it's nested it may
	//       be a child workspace root but that should be normal I think, consider)
	fileContents, err := ioutil.ReadFile(cbFile)
	if err != nil {
		msg := fmt.Sprintf("Codebase file \"%s\" read failed\n", cbFile)
		return out.WrapErr(err, msg, 3000)
	}
	err = cb.Read(bytes.NewReader(fileContents))
	return err
}
