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

import "strings"

func (c *Command) String() string {
	if c.Command == "" && c.Script == "" {
		return "(invalid command with neither Command nor Script set)"
	}
	if c.Command != "" && c.Script != "" {
		return "(invalid command with both Command and Script set)"
	}
	if c.Command != "" {
		return c.Command
	}
	return summarize(c.Script)
}

// summarize returns a short representation of a multi-line command.
func summarize(script string) string {
	var lines []string
	for i, line := range strings.Split(script, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i == 0 && strings.HasPrefix(line, "set -") {
			continue
		}
		lines = append(lines, line)
	}
	joined := strings.Join(lines, "\\n ")
	if len(joined) > 70 {
		return joined[:67] + "..."
	}
	return joined
}
