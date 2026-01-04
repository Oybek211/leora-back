package utils

import (
	"fmt"
	"strconv"
)

// ParsePaginationParams parses page/limit query params with sane defaults.
func ParsePaginationParams(pageParam, limitParam string) (int, int, error) {
	page := 1
	limit := 20

	if pageParam != "" {
		parsed, err := strconv.Atoi(pageParam)
		if err != nil || parsed < 1 {
			return 0, 0, fmt.Errorf("page must be a positive integer")
		}
		page = parsed
	}
	if limitParam != "" {
		parsed, err := strconv.Atoi(limitParam)
		if err != nil || parsed < 1 {
			return 0, 0, fmt.Errorf("limit must be a positive integer")
		}
		limit = parsed
	}
	if limit > 100 {
		limit = 100
	}

	return page, limit, nil
}

// SliceBounds returns clamped slice bounds for pagination.
func SliceBounds(total, page, limit int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	start := (page - 1) * limit
	if start >= total {
		return total, total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return start, end
}

// TotalPages computes total pages given count/limit.
func TotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	return pages
}
