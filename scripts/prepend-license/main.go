/*
 * Copyright (c) 2022, arivum.
 * All rights reserved.
 * SPDX-License-Identifier: MIT
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/MIT
 */

//#nosec
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

//go:generate go run gitlab.com/arifin/finnie/scripts/prepend-license

var (
	year, _, _ = time.Now().Date()
	regexps    = map[string]*regexp.Regexp{
		"yaml":         regexp.MustCompile(`(?s)^#.*?SPDX-License-Identifier.*?\n[^#]`),
		"go":           regexp.MustCompile(`(?s)^/\*.*?SPDX-License-Identifier.*?\*/\n\n`),
		"helmTemplate": regexp.MustCompile(`(?s)^{{/\*.*?SPDX-License-Identifier.*?\*/}}\n\n`),
	}
	licenses = map[string]string{
		"yaml": fmt.Sprintf(`#
# * Copyright (c) %d, arivum.
# * All rights reserved.
# * SPDX-License-Identifier: MIT
# * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/MIT
#

`, year),
		"go": fmt.Sprintf(`/*
 * Copyright (c) %d, arivum.
 * All rights reserved.
 * SPDX-License-Identifier: MIT
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/MIT
 */

`, year),
		"helmTemplate": fmt.Sprintf(`{{/*
 * Copyright (c) %d, arivum.
 * All rights reserved.
 * SPDX-License-Identifier: MIT
 * For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/MIT
*/}}

`, year),
	}
)

func getRoot(cwd string) (string, error) {
	var (
		newDir string
		dir    = cwd
		_, err = os.Stat(filepath.Join(dir, "go.mod"))
	)

	for err != nil {
		if newDir = filepath.Dir(dir); newDir == dir {
			return "", errors.New("could not find go.mod")
		} else {
			dir = newDir
		}
		_, err = os.Stat(filepath.Join(dir, "go.mod"))
	}
	return dir, nil
}

func prependLicenseToFile(fileName string, fileType string) error {
	var (
		err        error
		rawContent []byte
		matches    bool
	)

	if rawContent, err = ioutil.ReadFile(fileName); err != nil {
		return err
	}

	if matches = regexps[fileType].Match(rawContent); !matches {
		if err = os.WriteFile(fileName, append([]byte(licenses[fileType]), rawContent...), 0664); err != nil {
			return err
		}
	} else {
		rawContent = regexps[fileType].ReplaceAll(rawContent, []byte(licenses[fileType]))
		if err = os.WriteFile(fileName, rawContent, 0664); err != nil {
			return err
		}
	}

	return err
}

func prependLicense(goRoot string) error {
	return filepath.Walk(goRoot, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			switch {
			case filepath.Ext(info.Name()) == ".go":
				return prependLicenseToFile(path, "go")
			case strings.HasSuffix(filepath.Dir(path), "helm/templates"):
				return prependLicenseToFile(path, "helmTemplate")
			case filepath.Ext(info.Name()) == ".yaml":
				return prependLicenseToFile(path, "yaml")
			case info.Name() == "Dockerfile":
				return prependLicenseToFile(path, "yaml")
			case info.Name() == "_helpers.tpl":
				return prependLicenseToFile(path, "helmTemplate")
			}

		}
		return nil
	})
}

func main() {
	var (
		cwd    string
		err    error
		goRoot string
	)

	if cwd, err = os.Getwd(); err != nil {
		log.Fatal(err)
	}

	if goRoot, err = getRoot(cwd); err != nil {
		log.Fatal(err)
	}

	if err = prependLicense(goRoot); err != nil {
		log.Fatal(err)
	}
}
