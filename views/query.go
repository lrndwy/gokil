package views

import (
	"strconv"
	"strings"
)

// QueryInt parses a query parameter as int, returning fallback on missing/invalid input.
func (c *Context) QueryInt(name string, fallback int) int {
	raw := c.Query(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

// QueryInt64 parses a query parameter as int64, returning fallback on missing/invalid input.
func (c *Context) QueryInt64(name string, fallback int64) int64 {
	raw := c.Query(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return value
}

// QueryBool parses a query parameter as bool, returning fallback on missing/invalid input.
func (c *Context) QueryBool(name string, fallback bool) bool {
	raw := c.Query(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}

// MustParam returns 400 Bad Request when a route param is empty.
func (c *Context) MustParam(name string) (string, error) {
	value := c.Param(name)
	if strings.TrimSpace(value) == "" {
		return "", BadRequest(name + " is required")
	}
	return value, nil
}

// Pagination reads `page` and `limit` query params with safe defaults.
func (c *Context) Pagination(defaultLimit, maxLimit int) (page, limit, offset int) {
	if defaultLimit <= 0 {
		defaultLimit = 20
	}
	if maxLimit <= 0 {
		maxLimit = 100
	}
	page = c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}
	limit = c.QueryInt("limit", defaultLimit)
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	offset = (page - 1) * limit
	return page, limit, offset
}
