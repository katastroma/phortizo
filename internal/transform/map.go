//revive:disable:package-comments
package transform

// Map transforms each element of a slice using a function that can fail.
func Map[From, To any](items []From, fn func(From) (To, error)) ([]To, error) {
	results := make([]To, 0, len(items))
	for _, item := range items {
		result, err := fn(item)
		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}
