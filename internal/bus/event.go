package bus

type Event struct {
	Action string
	Topic  string
	Data   any
}
