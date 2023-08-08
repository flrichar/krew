// Copyright 2019 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
)

// EnsureCloned will clone into the destination path, otherwise will return no error.
func EnsureCloned(uri, destinationPath string) error {

	if ok, err := IsGitCloned(destinationPath); err != nil {
		return err
	} else if !ok {
		// Clone the given repository to the given directory
		Info("git clone %s %s --recursive", url, directory)

		r, err := git.PlainClone(destinationPath, false, &git.CloneOptions{
			URL:               uri,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		CheckIfError(err)

		// ... retrieving the branch being pointed by HEAD
		ref, err := r.Head()
		CheckIfError(err)
		// ... retrieving the commit object
		commit, err := r.CommitObject(ref.Hash())
		CheckIfError(err)

		fmt.Println(commit)
		return err
	}
	return nil
}

// IsGitCloned will test if the path is a git dir.
func IsGitCloned(gitPath string) (bool, error) {
	f, err := os.Stat(filepath.Join(gitPath, ".git"))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil && f.IsDir(), err
}

// update will fetch origin and set HEAD to origin/HEAD
// and also will create a pristine working directory by removing
// untracked files and directories.
func updateAndCleanUntracked(destinationPath string) error {

	////// fetch-reset-clean goes here
	if _, err := Exec(destinationPath, "fetch", "-v"); err != nil {
		return errors.Wrapf(err, "fetch index at %q failed", destinationPath)
	}

	if _, err := Exec(destinationPath, "reset", "--hard", "@{upstream}"); err != nil {
		return errors.Wrapf(err, "reset index at %q failed", destinationPath)
	}

	_, err := Exec(destinationPath, "clean", "-xfd")
	return errors.Wrapf(err, "clean index at %q failed", destinationPath)
	/////
}

// EnsureUpdated will ensure the destination path exists and is up to date.
func EnsureUpdated(uri, destinationPath string) error {
	if err := EnsureCloned(uri, destinationPath); err != nil {
		return err
	}
	return updateAndCleanUntracked(destinationPath)
}

// GetRemoteURL returns the url of the remote origin
func GetRemoteURL(dir string) (string, error) {

	// We instantiate a new repository targeting the given path (the .git folder)
	r, err := git.PlainOpen(dir)
	CheckIfError(err)

	// Get the .git/config information
	Info("git config --get remote.origin.url")

	// Get the remote origin.
	origin, err := r.Remote("origin")
	CheckIfError(err)

	// Get the remote origin URL.
	originURL := origin.Config().URLs[0]

	// Print the remote origin URL.
	return fmt.Println(originURL)

}

////// replace Exec() with Info() & CheckIfError(), keeping krew conventions

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}
	errors.Wrapf(err, "error: %q")
}

// Info should be used to describe the example commands that are about to run.
func Info(format string, args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf(format, args...))
}

//func Exec(pwd string, args ...string) (string, error) {
//	klog.V(4).Infof("Going to run git %s", strings.Join(args, " "))
//	cmd := osexec.Command("git", args...)
//	cmd.Dir = pwd
//	buf := bytes.Buffer{}
//	var w io.Writer = &buf
//	if klog.V(2).Enabled() {
//		w = io.MultiWriter(w, os.Stderr)
//	}
//	cmd.Stdout, cmd.Stderr = w, w
//	if err := cmd.Run(); err != nil {
//		return "", errors.Wrapf(err, "command execution failure, output=%q", buf.String())
//	}
//	return strings.TrimSpace(buf.String()), nil
//}
