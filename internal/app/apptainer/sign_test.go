// Copyright (c) Contributors to the Apptainer project, established as
//   Apptainer a Series of LF Projects LLC.
//   For website terms of use, trademark policy, privacy policy and other
//   project policies see https://lfprojects.org/policies
// Copyright (c) 2020-2022, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the LICENSE.md file
// distributed with the sources of this project regarding your rights to use or distribute this
// software.

package apptainer

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/apptainer/apptainer/pkg/sypgp"
	"github.com/apptainer/sif/v2/pkg/integrity"
)

// tempFileFrom copies the file at path to a temporary file, and returns a reference to it.
func tempFileFrom(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	pattern := "*"
	if ext := filepath.Ext(path); ext != "" {
		pattern = fmt.Sprintf("*.%s", ext)
	}

	tf, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer tf.Close()

	if _, err := io.Copy(tf, f); err != nil {
		return "", err
	}

	return tf.Name(), nil
}

func mockEntitySelector(t *testing.T) sypgp.EntitySelector {
	e := getTestEntity(t)

	return func(openpgp.EntityList) (*openpgp.Entity, error) {
		return e, nil
	}
}

func TestSign(t *testing.T) {
	mockEntityOpt := OptSignEntitySelector(mockEntitySelector(t))

	tests := []struct {
		name    string
		path    string
		opts    []SignOpt
		wantErr error
	}{
		{
			name:    "ErrNoKeyMaterial",
			path:    filepath.Join("testdata", "images", "one-group.sif"),
			wantErr: integrity.ErrNoKeyMaterial,
		},
		{
			name: "Defaults",
			path: filepath.Join("testdata", "images", "one-group.sif"),
			opts: []SignOpt{mockEntityOpt},
		},
		{
			name: "OptSignGroup",
			path: filepath.Join("testdata", "images", "one-group.sif"),
			opts: []SignOpt{mockEntityOpt, OptSignGroup(1)},
		},
		{
			name: "OptSignObjects",
			path: filepath.Join("testdata", "images", "one-group.sif"),
			opts: []SignOpt{mockEntityOpt, OptSignObjects(1)},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Signing modifies the file, so work with a temporary file.
			path, err := tempFileFrom(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(path)

			if got, want := Sign(path, tt.opts...), tt.wantErr; !errors.Is(got, want) {
				t.Errorf("got error %v, want %v", got, want)
			}
		})
	}
}
