package web

// Result is the result of a controller.
type Result interface {
	Render(ctx *Ctx) error
}
