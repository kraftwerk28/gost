package formatting

type NamedArgs map[string]interface{}

type Formatter interface {
	Sprintf(s string, a NamedArgs) string
}
