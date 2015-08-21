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
	"fmt"
	"testing"

	"github.com/dvln/pretty"
)

var jsonExample = []byte(` { "name" : "dvln",
   "desc" : "Multi-package and workspace management tool",
   "contacts" : [ "Erik Brady <brady@dvln.org>" ],
   "attrs" : {
     "linkalias" : "True",
     "jobs" : "4",
     "key2" : "val2",
     "key3" : "True"
   },
   "pathing": {
   	 "ws_pfx_dir" : "{{if .GoPkg}}src{{end}}"
   },
   "vars" : {
     "dvln": "http://github.com/dvln",
     "joe": "http://github.com/joe",
     "spf13": "http://github.com/spf13"
   },
   "access" : {
     "m,^{{.dvln}}/.*, && Vendor!=True": {
       "read" : "open",
       "write": "http://dvln.org/api/v1/access?codebase={{.TopCodebase}}&pkg={{.Pkg}}&branch={{.Branch}}&devline={{.Devline}}&user={{.UserID}}&type=write"
     }
   },
   "pkgs" : [
     { "id" : "22",
       "name" : "dvln/lib/3rd/viper",
       "desc" : "dvln project copy of spf13/viper package",
       "vcs" : "git",
       "ws" : "src/dvln/lib/3rd/viper",
       "repo" : { "rw": "{{.dvln}}/viper" },
       "aliases" : {
         "dvln/lib/olddir/viper": "src/dvln/lib/olddir/viper",
         "dvln/reallyolddir/viper": "src/dvln/reallyolddir/viper"
       },
       "attrs" : {
         "GoPkg" : "True",
         "Vendor": "True",
         "Owners": "jessie@co.com",
         "Readers" : "anyone",
         "Committers" : "team1@co.com"
       },
       "remotes" : {
         "vendor": { "r": "{{.spf13}}/viper" },
         "joe": { "read": "{{.joe}}/viper" }
       },
       "status" : "active"
     },
     { "id" : "23",
       "name" : "dvln/lib/out",
       "desc" : "dvln project copy of spf13/viper package",
       "vcs" : "git",
       "ws" : "src/dvln/lib/out",
       "repo" : { "rw": "{{.dvln}}/out" },
       "aliases" : {
         "dvln/lib/oldoutname": "src/dvln/lib/oldoutname"
       },
       "attrs" : {
         "GoPkg" : "True",
         "Owners": "dvln@dvln.org",
         "Readers" : "anyone",
         "Committers" : "team_dvln@dvln.org"
       },
       "remotes" : {
         "dhowlett" : { "rw": "{{dhowlett}}/out" }
       },
       "access" : {
         "read" : "open",
         "write": "http://dvln.org/api/v1/access?codebase={{.TopCodebase}}&pkg={{.Pkg}}&branch={{.Branch}}&devline={{.Devline}}&user={{.UserID}}&type=write"
       },
       "status" : "active"
     },
     { "id" : "24",
       "name" : "dvln/web/hugo",
       "desc" : "dvln project copy of spf13/viper package",
       "vcs" : "dvln",
       "codebase" : "hugo",
       "ws" : "src/dvln/web/hugo",
       "repo" : { "rw": "{{.dvln}}/hugo" },
       "aliases" : {
         "github.com/spf13/hugo": "src/github.com/spf13/hugo"
       },
       "attrs" : {
         "GoPkg" : "True",
         "Owners": "spf13@spf13.com",
         "Readers" : "anyone"
       },
       "remotes" : {
         "vendor,spf13": { "r": "{{.spf13}}/hugo" }
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

func TestCodebaseParse(t *testing.T) {
	fmt.Println("About to test codebase parsing")
	codebaseDefn, err := FindAndRead("dvln")
	if err != nil {
		fmt.Printf("Error:\n%s\n", err)
	}
	//fmt.Printf("Results:\n%+v\n", codebaseDefn)
	fmt.Printf("Results:\n%# v\n", pretty.Formatter(codebaseDefn))
	//eriknow
	/*
		assert.NotContains(t, screenBuf.String(), "trace info")
		assert.NotContains(t, screenBuf.String(), "debugging info")
		assert.NotContains(t, screenBuf.String(), "verbose info")
		assert.Contains(t, screenBuf.String(), "information")
		assert.Contains(t, screenBuf.String(), "key note")
	*/
}
