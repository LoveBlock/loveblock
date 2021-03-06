// Copyright 2016 The loveblock Authors
// This file is part of the loveblock library.
//
// The loveblock library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The loveblock library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the loveblock library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/LoveBlock/loveblock/internal/cmdtest"
	"github.com/docker/docker/pkg/reexec"
)

func tmpdir(t *testing.T) string {
	dir, err := ioutil.TempDir("", "loveblock-test")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

type testloveblock struct {
	*cmdtest.TestCmd

	// template variables for expect
	Datadir  string
	Lovebase string
}

func init() {
	// Run the app if we've been exec'd as "loveblock-test" in runGnetwork.
	reexec.Register("loveblock-test", func() {
		if err := app.Run(os.Args); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	})
}

func TestMain(m *testing.M) {
	// check if we have been reexec'd
	if reexec.Init() {
		return
	}
	os.Exit(m.Run())
}

// spawns loveblock with the given command line args. If the args don't set --datadir, the
// child g gets a temporary data directory.
func runGnetwork(t *testing.T, args ...string) *testloveblock {
	tt := &testloveblock{}
	tt.TestCmd = cmdtest.NewTestCmd(t, tt)
	for i, arg := range args {
		switch {
		case arg == "-datadir" || arg == "--datadir":
			if i < len(args)-1 {
				tt.Datadir = args[i+1]
			}
		case arg == "-networkbase" || arg == "--networkbase":
			if i < len(args)-1 {
				tt.Lovebase = args[i+1]
			}
		}
	}
	if tt.Datadir == "" {
		tt.Datadir = tmpdir(t)
		tt.Cleanup = func() { os.RemoveAll(tt.Datadir) }
		args = append([]string{"-datadir", tt.Datadir}, args...)
		// Remove the temporary datadir if somnetworking fails below.
		defer func() {
			if t.Failed() {
				tt.Cleanup()
			}
		}()
	}

	// Boot "loveblock". This actually runs the test binary but the TestMain
	// function will prevent any tests from running.
	tt.Run("loveblock-test", args...)

	return tt
}
