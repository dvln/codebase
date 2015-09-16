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

// Package test: out
//      Simple testing for the 'out' package ... could use more around testing
//      new prefix setting, adjusting of formats, adding in checks to make sure
//      the date/time is in the log output but not the screen output

package codebase

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/dvln/out"
	"github.com/dvln/pretty"
)

var codebaseExample = []byte(` { "name" : "dvln",
   "desc" : "Multi-package and workspace management tool",
   "contacts" : {
   	 "authors" : [ "Erik Brady <brady@dvln.org>" ]
   },
   "attrs" : {
     "linkalias" : "True",
     "jobs" : "4",
     "codebasePkgDef" : "required",
     "key1" : "val2"
   },
   "pathing": {
   	 "wkspc_pfx_dir" : "{{if .GoPkg}}src{{end}}"
   },
   "vars" : {
     "dvln": "http://github.com/dvln",
     "joe": "http://github.com/joe",
     "spf13": "http://github.com/spf13",
     "dhowlett": "http://github.com/dhowlett"
   },
   "access" : {
     "m,^{{.dvln}}/*, && Vendor!=True": {
       "read" : "open",
       "write": "http://dvln.org/api/v1/access?codebase={{.TopCodebase}}&pkg={{.Pkg}}&branch={{.Branch}}&devline={{.Devline}}&user={{.UserID}}&type=write"
     }
   },
   "pkgs" : [
     { "id" : "22",
       "name" : "dvln/lib/3rd/viper",
       "desc" : "dvln project copy of spf13/viper package",
       "ws" : "src/dvln/lib/3rd/viper",
       "aliases" : { "dvln/lib/olddir/viper": "src/dvln/lib/olddir/viper",
                     "dvln/reallyolddir/viper": "src/dvln/reallyolddir/viper" },
       "vcs" : [ {
       	   "type" : "git",
       	   "fmts" : [ "vcs", "src", "source" ],
       	   "repo" : { "rw": "{{.dvln}}/viper" },
           "remotes" : {
              "vendor,spf13": { "r": "{{.spf13}}/viper" },
              "joe": { "read": "{{.joe}}/viper" }
           }
       } ],
       "attrs" : {
         "GoPkg" : "True",
         "Vendor": "True",
         "Owners": "jessie@co.com",
         "Readers" : "anyone",
         "Committers" : "team1@co.com"
       },
       "status" : "active"
     },
     { "id" : "23",
       "name" : "dvln/lib/out",
       "desc" : "dvln project copy of spf13/viper package",
       "license" : "Apache-2.0",
       "ws" : "src/dvln/lib/out",
       "aliases" : { "dvln/lib/oldoutname": "src/dvln/lib/oldoutname" },
       "vcs" : [ {
       	   "type" : "git",
       	   "fmts" : [ "vcs", "src", "source" ],
           "repo" : { "rw": "{{.dvln}}/out" },
           "remotes" : { "dhowlett": { "r": "{{.dhowlett}}/out" } }
       } ],
       "attrs" : {
         "GoPkg" : "True",
         "Owners": "dvln@dvln.org",
         "Readers" : "anyone",
         "Committers" : "team_dvln@dvln.org"
       },
       "access" : {
         "read" : "open",
         "write": "http://dvln.org/api/v1/access?codebase={{.TopCodebase}}&pkg={{.Pkg}}&branch={{.Branch}}&devline={{.Devline}}&user={{.UserID}}&type=write"
       },
       "status" : "active"
     },
     { "id" : "24",
       "name" : "dvln/web/hugo",
       "desc" : "a fast static website server",
       "class" : "codebase",
       "ws" : "src/dvln/web/hugo",
       "aliases" : { "github.com/spf13/hugo": "src/github.com/spf13/hugo" },
       "vcs" : [ {
       	   "type" : "git",
       	   "fmts" : [ "vcs", "src", "source" ],
           "repo" : { "rw": "{{.dvln}}/hugo" },
           "remotes" : { "vendor,spf13": { "r": "{{.spf13}}/hugo" } }
       } ],
       "attrs" : {
         "GoPkg" : "True",
         "Owners": "spf13@spf13.com",
         "Readers" : "anyone"
       },
       "status" : "active"
     }
   ]
 }
`)

var codebaseExpanded = []byte(` { "name" : "dvln",
   "desc" : "Multi-package and workspace management tool",
   "contacts" : {
   	 "authors" : [ "Erik Brady <brady@dvln.org>" ]
   },
   "attrs" : {
     "linkalias" : "True",
     "jobs" : "4",
     "codebasePkgDef" : "required",
     "key1" : "val2"
   },
   "pathing": {
   	 "wkspc_pfx_dir" : "{{if .GoPkg}}src{{end}}"
   },
   "vars" : {
     "dvln": "http://github.com/dvln",
     "joe": "http://github.com/joe",
     "spf13": "http://github.com/spf13",
     "dhowlett": "http://github.com/dhowlett"
   },
   "access" : {
     "m,^http://github.com/dvln/*, && Vendor!=True": {
       "read" : "open",
       "write": "http://dvln.org/api/v1/access?codebase={{.TopCodebase}}&pkg={{.Pkg}}&branch={{.Branch}}&devline={{.Devline}}&user={{.UserID}}&type=write"
     }
   },
   "pkgs" : [
     { "id" : "22",
       "name" : "dvln/lib/3rd/viper",
       "desc" : "dvln project copy of spf13/viper package",
       "ws" : "src/dvln/lib/3rd/viper",
       "aliases" : {
         "dvln/lib/olddir/viper": "src/dvln/lib/olddir/viper",
         "dvln/reallyolddir/viper": "src/dvln/reallyolddir/viper"
       },
       "vcs" : [ {
       	   "type" : "git",
       	   "fmts" : [ "vcs", "src", "source" ],
       	   "repo" : { "rw": "http://github.com/dvln/viper" },
           "remotes" : {
              "vendor,spf13": { "r": "http://github.com/spf13/viper" },
              "joe": { "read": "http://github.com/joe/viper" }
           }
       } ],
       "attrs" : {
         "GoPkg" : "True",
         "Vendor": "True",
         "Owners": "jessie@co.com",
         "Readers" : "anyone",
         "Committers" : "team1@co.com"
       },
       "status" : "active"
     },
     { "id" : "23",
       "name" : "dvln/lib/out",
       "desc" : "dvln project copy of spf13/viper package",
       "license" : "Apache-2.0",
       "ws" : "src/dvln/lib/out",
       "aliases" : {
         "dvln/lib/oldoutname": "src/dvln/lib/oldoutname"
       },
       "attrs" : {
         "GoPkg" : "True",
         "Owners": "dvln@dvln.org",
         "Readers" : "anyone",
         "Committers" : "team_dvln@dvln.org"
       },
       "vcs" : [ {
       	   "type" : "git",
       	   "fmts" : [ "vcs", "src", "source" ],
           "repo" : { "rw": "http://github.com/dvln/out" },
           "remotes" : { "dhowlett": { "r": "http://github.com/dhowlett/out" } }
       } ],
       "access" : {
         "read" : "open",
         "write": "http://dvln.org/api/v1/access?codebase={{.TopCodebase}}&pkg={{.Pkg}}&branch={{.Branch}}&devline={{.Devline}}&user={{.UserID}}&type=write"
       },
       "status" : "active"
     },
     { "id" : "24",
       "name" : "dvln/web/hugo",
       "desc" : "a fast static website server",
       "class" : "codebase",
       "ws" : "src/dvln/web/hugo",
       "aliases" : {
         "github.com/spf13/hugo": "src/github.com/spf13/hugo"
       },
       "vcs" : [ {
       	   "type" : "git",
       	   "fmts" : [ "vcs", "src", "source" ],
           "repo" : { "rw": "http://github.com/dvln/hugo" },
           "remotes" : { "vendor,spf13": { "r": "http://github.com/spf13/hugo" } }
       } ],
       "attrs" : {
         "GoPkg" : "True",
         "Owners": "spf13@spf13.com",
         "Readers" : "anyone"
       },
       "status" : "active"
     }
   ]
 }
`)

// This example has a missing } in a "repo" field template
var badCodebaseTemplateExample = []byte(` { "name" : "dvln",
   "desc" : "Multi-package and workspace management tool",
   "contacts" : {
   	 "authors" : [ "Erik Brady <brady@dvln.org>" ]
   },
   "attrs" : {
     "linkalias" : "True",
     "jobs" : "4",
     "codebasePkgDef" : "required",
     "key1" : "val2"
   },
   "pathing": {
   	 "wkspc_pfx_dir" : "{{if .GoPkg}}src{{end}}"
   },
   "vars" : {
     "dvln": "http://github.com/dvln",
     "joe": "http://github.com/joe",
     "spf13": "http://github.com/spf13",
     "dhowlett": "http://github.com/dhowlett"
   },
   "access" : {
     "m,^{{.dvln}}/*, && Vendor!=True": {
       "read" : "open",
       "write": "http://dvln.org/api/v1/access?codebase={{.TopCodebase}}&pkg={{.Pkg}}&branch={{.Branch}}&devline={{.Devline}}&user={{.UserID}}&type=write"
     }
   },
   "pkgs" : [
     { "id" : "22",
       "name" : "dvln/lib/3rd/viper",
       "desc" : "dvln project copy of spf13/viper package",
       "ws" : "src/dvln/lib/3rd/viper",
       "aliases" : {
         "dvln/lib/olddir/viper": "src/dvln/lib/olddir/viper",
         "dvln/reallyolddir/viper": "src/dvln/reallyolddir/viper"
       },
       "vcs" : [ {
       	   "type" : "git",
       	   "fmts" : [ "vcs", "src", "source" ],
       	   "repo" : { "rw": "{{.dvln}/viper" },
           "remotes" : {
              "vendor,spf13": { "r": "{{.spf13}}/viper" },
              "joe": { "read": "{{.joe}}/viper" }
           }
       } ],
       "attrs" : {
         "GoPkg" : "True",
         "Vendor": "True",
         "Owners": "jessie@co.com",
         "Readers" : "anyone",
         "Committers" : "team1@co.com"
       },
       "status" : "active"
     },
     { "id" : "23",
       "name" : "dvln/lib/out",
       "desc" : "dvln project copy of spf13/viper package",
       "license" : "Apache-2.0",
       "ws" : "src/dvln/lib/out",
       "aliases" : {
         "dvln/lib/oldoutname": "src/dvln/lib/oldoutname"
       },
       "vcs" : [ {
       	   "type" : "git",
       	   "fmts" : [ "vcs", "src", "source" ],
           "repo" : { "rw": "{{.dvln}}/out" },
           "remotes" : { "dhowlett": { "r": "{{.dhowlett}}/out" } }
       } ],
       "attrs" : {
         "GoPkg" : "True",
         "Owners": "dvln@dvln.org",
         "Readers" : "anyone",
         "Committers" : "team_dvln@dvln.org"
       },
       "access" : {
         "read" : "open",
         "write": "http://dvln.org/api/v1/access?codebase={{.TopCodebase}}&pkg={{.Pkg}}&branch={{.Branch}}&devline={{.Devline}}&user={{.UserID}}&type=write"
       },
       "status" : "active"
     },
     { "id" : "24",
       "name" : "dvln/web/hugo",
       "desc" : "a fast static website server",
       "class" : "codebase",
       "ws" : "src/dvln/web/hugo",
       "aliases" : {
         "github.com/spf13/hugo": "src/github.com/spf13/hugo"
       },
       "vcs" : [ {
       	   "type" : "git",
       	   "fmts" : [ "vcs", "src", "source" ],
           "repo" : { "rw": "{{.dvln}}/hugo" },
           "remotes" : { "vendor,spf13": { "r": "{{.spf13}}/hugo" } }
       } ],
       "attrs" : {
         "GoPkg" : "True",
         "Owners": "spf13@spf13.com",
         "Readers" : "anyone"
       },
       "status" : "active"
     }
   ]
 }
`)

// Example of more standard Go test infra usage (if removing assert)
//	if tracer == nil {
//		t.Error("Return from New should not be nil")
//	}
//	tracer.Trace("Hello trace package.")
//	if buf.String() != "Hello trace package.\n" {
//		t.Errorf("Trace should not write '%s'.", buf.String())
//	}

/*
func TestLevels(t *testing.T) {
	SetThreshold(LevelIssue, ForScreen)
	assert.Equal(t, Threshold(ForScreen), LevelIssue)
	SetThreshold(LevelError, ForLogfile)
	assert.Equal(t, Threshold(ForLogfile), LevelError)
	assert.NotEqual(t, Threshold(ForScreen), LevelError)
	SetThreshold(LevelNote, ForScreen)
	assert.Equal(t, Threshold(ForScreen), LevelNote)
}
*/

func TestCodebaseFileRead(t *testing.T) {
	//eriknow: fix this up, need to write the above jsontemplate to a tmp
	//         file and then have this routine read that file (?), right
	//         now this relies upon the 'dvln.codebase' file in my home dir
	codebaseDefn := New()
	err := codebaseDefn.Get("dvln")
	if err != nil {
		t.Fatal(err)
	}
	if codebaseDefn == nil {
		t.Fatal("The codebase file read should have had data, there is none")
	}
	if codebaseDefn.Name != "dvln" {
		t.Fatal("The codebase should have the name \"dvln\" but does not")
	}
	//results := fmt.Sprintf("%# v\n", pretty.Formatter(codebaseDefn))
	//fmt.Printf("Codebase contents:\n%v", results)
}

func TestCodebaseParse(t *testing.T) {
	b := bytes.NewBuffer(codebaseExample)
	codebaseDefn := New()
	err := codebaseDefn.Read(b)
	if err != nil {
		t.Error(err)
		t.Fatal("Error reading pre-defined codebase JSON")
	}
	e := bytes.NewBuffer(codebaseExpanded)
	expandedDefn := &Defn{}
	err = expandedDefn.Read(e)
	if err != nil {
		t.Error(err)
		t.Fatal("Error reading pre-defined \"expanded\" codebase JSON")
	}
	// see if the read/expanded and pre/expanded results match, basically
	// this checks if the template expansion is working, not much else
	eq := reflect.DeepEqual(codebaseDefn, expandedDefn)
	if !eq {
		t.Error("Raw codebase read/expansion failed to match pre-expanded copy")
		t.Error("Examine the output differences but ignore hash ordering differences,")
		t.Error("the non-hash ordering differences will highlight whatever the issue is:")
		results := fmt.Sprintf("%# v\n", pretty.Formatter(codebaseDefn))
		found := strings.Split(results, "\n")
		preDefined := fmt.Sprintf("%# v\n", pretty.Formatter(expandedDefn))
		expected := strings.Split(preDefined, "\n")
		for index, line := range expected {
			if line != found[index] {
				t.Errorf("Expected Line #: %d\n  Expected: \"%s\"\n  Results : \"%s\"\n", index, line, found[index])
			}
		}
		t.Fatalf("Please determine why the code base read/expansion failed, full results:\n%s\n", results)
	}
	// sample a few fields and make sure they are as expected
	if codebaseDefn.Name != "dvln" {
		t.Fatalf("Codebase name failed to match test name \"dvln\", found: %s", codebaseDefn.Name)
	}
	if codebaseDefn.Desc != "Multi-package and workspace management tool" {
		t.Fatalf("Codebase description failed to match test desc, found: %s", codebaseDefn.Desc)
	}
	if codebaseDefn.Attrs["jobs"] != "4" {
		t.Fatalf("Codebase attributes jobs field should be set to 4, found %s", codebaseDefn.Attrs["jobs"])
	}
	if codebaseDefn.Vars["dvln"] != "http://github.com/dvln" {
		t.Fatalf("Codebase vars setting for \"dvln\" is not set correctly, found: %s", codebaseDefn.Vars["dvln"])
	}
	if len(codebaseDefn.Pkgs) < 3 {
		t.Fatalf("Codebase sample should have had 3 or more packages... but does not")
	}
	results := fmt.Sprintf("%# v\n", pretty.Formatter(codebaseDefn))
	fmt.Printf("Codebase contents:\n%v", results)
}

func TestBadCodebaseParse(t *testing.T) {
	b := bytes.NewBuffer(badCodebaseTemplateExample)
	codebaseDefn := &Defn{}
	err := codebaseDefn.Read(b)
	if err == nil {
		t.Fatal("Bad codebase example wasn't detected as bad, we have a problem")
	}
	if !out.IsError(err, nil, 3004) {
		t.Error(out.DefaultError(err.(out.DetailedError), false, false, true))
		t.Fatal("Failed to match correct error from codebase read, expected 3004")
	}
}
