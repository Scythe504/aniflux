package utils

func Paginate[T any](data []T, page, perPage int) ([]T) {
	start := (page - 1) * perPage
	end := start + perPage
	
	if start >= len(data) {
		return []T{}
	}

	if end > len(data) {
		end = len(data)
	}

	return data[start:end]
}