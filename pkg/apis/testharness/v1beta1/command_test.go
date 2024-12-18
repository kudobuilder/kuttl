/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommand_String(t *testing.T) {
	tests := map[string]struct {
		command string
		script  string
		want    string
	}{
		"empty": {
			want: "(invalid command with neither Command nor Script set)",
		},
		"both": {
			command: "foo",
			script:  "bar",
			want:    "(invalid command with both Command and Script set)",
		},
		"command": {
			command: "foo \\ \"; bar",
			want:    "foo \\ \"; bar",
		},
		"short script": {
			script: "make \\ \"; all",
			want:   "make \\ \"; all",
		},
		"long script": {
			script: `set -eou pipefail
first line
# comment which should be ignored, as should the above set and the following empty line

second line
AAAAAAAAAAAAAAAAA BBBBBBBBBBBBBBBBB CCCCCCCCCCCCCCCCC DDDDDDDDDDDDDDDDD
AAAAAAAAAAAAAAAAA BBBBBBBBBBBBBBBBB CCCCCCCCCCCCCCCCC DDDDDDDDDDDDDDDDD
`,
			want: "first line\\n second line\\n AAAAAAAAAAAAAAAAA BBBBBBBBBBBBBBBBB CCCC...",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := &Command{
				Command: tt.command,
				Script:  tt.script,
			}
			assert.Equalf(t, tt.want, c.String(), "String()")
		})
	}
}
