package all

import "slices"

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

func Persist[T any, R any](x *executor[T, R]) {
	x.abortOnError = false
}
