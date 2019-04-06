package action

var (
	actions = make(map[string]Action)
)

type Action interface {
	Run() error
	Type() string
}

func Find(name string) Action {
	return actions[name]
}

func register(action Action) {
	actions[action.Type()] = action
}
