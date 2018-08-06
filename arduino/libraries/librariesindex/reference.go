/*
 * This file is part of arduino-cli.
 *
 * Copyright 2018 ARDUINO AG (http://www.arduino.cc/)
 *
 * arduino-cli is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program; if not, write to the Free Software
 * Foundation, Inc., 51 Franklin St, Fifth Floor, Boston, MA  02110-1301  USA
 *
 * As a special exception, you may use this file as part of a free software
 * library without restriction.  Specifically, if other files instantiate
 * templates or use macros or inline functions from this file, or you compile
 * this file and link it with other files to produce an executable, this
 * file does not by itself cause the resulting executable to be covered by
 * the GNU General Public License.  This exception does not however
 * invalidate any other reasons why the executable file might be covered by
 * the GNU General Public License.
 */

package librariesindex

import (
	"fmt"
	"strings"

	semver "go.bug.st/relaxed-semver"
)

// Reference uniquely identify a Library in the library index
type Reference struct {
	Name    string          // The name of the parsed item.
	Version *semver.Version // The Version of the parsed item.
}

func (r *Reference) String() string {
	if r.Version == nil {
		return r.Name
	}
	return r.Name + "@" + r.Version.String()
}

// ParseArgs parses a sequence of "item@version" tokens and returns a Name-Version slice.
//
// If version is not present it is assumed as "latest" version.
func ParseArgs(args []string) ([]*Reference, error) {
	res := []*Reference{}
	for _, item := range args {
		tokens := strings.SplitN(item, "@", 2)
		if len(tokens) == 2 {
			version, err := semver.Parse(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("invalid version %s: %s", version, err)
			}
			res = append(res, &Reference{Name: tokens[0], Version: version})
		} else {
			res = append(res, &Reference{Name: tokens[0]})
		}
	}
	return res, nil
}
