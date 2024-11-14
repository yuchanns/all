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

package all

import "strings"

type errors struct {
	errs []error
}

// Implement the error interface for the errors type.
func (e *errors) Error() string {
	var desc []string
	for _, err := range e.errs {
		desc = append(desc, err.Error())
	}
	return strings.Join(desc, ",")
}
