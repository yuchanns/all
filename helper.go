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

import "slices"

// Collect is a helper function that collects results from the executor
// and returns them in either ascending or descending order based on the input flag.
//
// Parameters:
//
//	x - The executor instance that provides the concurrent results.
//	descending - A boolean flag indicating whether the results should be sorted in descending order.
//
// Returns:
//
//	results - A slice of results collected from the executor, sorted in the specified order.
//	err - An error if any occurred during the collection process.
func Collect[T any, R any](x *executor[T, R], descending bool) (results []R, err error) {
	var orderedResults []ordered[R]
	for x.Next() {
		orderedResults = append(orderedResults, ordered[R]{index: x.Index(), data: x.Each()})
	}
	slices.SortStableFunc(orderedResults, func(i, j ordered[R]) int {
		if descending {
			return int(j.index - i.index)
		}
		return int(i.index - j.index)
	})
	for _, v := range orderedResults {
		results = append(results, v.data)
	}
	return results, x.Error()
}

// Persist allows the executor to continue collecting results despite errors.
func Persist[T any, R any](x *executor[T, R]) {
	x.abortOnError = false
}
