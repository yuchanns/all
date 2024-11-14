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
