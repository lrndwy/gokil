package views

// Data-oriented response helpers.
//
// These helpers exist for handlers that prefer loading data first (via orm helpers,
// custom logic, or external calls) and then responding with a consistent envelope.

// Listed writes a 200 OK response for list endpoints.
//
// Example:
//   users, _ := orm.GetAll(...)
//   return views.Listed(ctx, users, "users retrieved")
func Listed(c *Context, data any, message string) error {
	return c.OK(message, data)
}

// Detailed writes a 200 OK response for detail endpoints.
func Detailed(c *Context, data any, message string) error {
	return c.OK(message, data)
}

// Created writes a 201 Created response.
func Created(c *Context, data any, message string) error {
	return c.Created(message, data)
}

// Updated writes a 200 OK response for update endpoints.
func Updated(c *Context, data any, message string) error {
	return c.OK(message, data)
}

// Deleted writes a 200 OK response for delete endpoints.
//
// If you prefer 204 No Content, use:
//   return ctx.NoContentResponse()
func Deleted(c *Context, data any, message string) error {
	return c.OK(message, data)
}

