package formatting

type NamedArgs map[string]interface{}

type Formatter interface {
	Expand(NamedArgs) string
}
