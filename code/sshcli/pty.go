package sshcli

type writer struct {
	data []byte
}

func (w *writer) Write(p []byte) (int, error) {
	w.data = append(w.data, p...)
	return len(p), nil
}

func (w writer) String() string {
	return string(w.data)
}
