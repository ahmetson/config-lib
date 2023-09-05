package pipeline

type PipeEnd struct {
	Id  string
	Url string
}

// NewHandlerEnd creates a pipe end with the given name as the handler
func NewHandlerEnd(end string) *PipeEnd {
	return &PipeEnd{
		Url: "",
		Id:  end,
	}
}

func NewThisServicePipeEnd() *PipeEnd {
	return NewHandlerEnd("")
}

func (end *PipeEnd) IsHandler() bool {
	return len(end.Id) > 0
}

func (end *PipeEnd) Pipeline(head []string) *Pipeline {
	return &Pipeline{
		End:  end,
		Head: head,
	}
}
