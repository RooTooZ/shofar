// Package input предоставляет ввод текста в активное поле.
package input

// Typer вводит текст в активное поле ввода.
type Typer interface {
	// Type вводит текст в текущее активное поле.
	Type(text string) error
}

// New создаёт платформо-специфичный Typer.
func New() (Typer, error) {
	return newTyper()
}
