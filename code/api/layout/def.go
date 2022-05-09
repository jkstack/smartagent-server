package layout

type Main struct {
	Name    string `yaml:"name" json:"name"`
	Timeout uint   `yaml:"timeout" json:"timeout"`
	Tasks   []Task `yaml:"tasks" json:"tasks"`
}

type Task struct {
	// plugin info
	Exec `yaml:",inline" json:",inline"`
	File `yaml:",inline" json:",inline"`

	Name    string `yaml:"name" json:"name"`
	Plugin  string `yaml:"plugin" json:"plugin"`
	Auth    string `yaml:"auth" json:"auth"`
	If      string `yaml:"if" json:"if"`
	Timeout uint   `yaml:"timeout" json:"timeout"`
}

type Exec struct {
	Cmd    string `yaml:"cmd" json:"cmd"`
	Output string `yaml:"output" json:"output"`
}

type File struct {
	Action string `yaml:"action" json:"action"`
	Src    string `yaml:"src" json:"src"`
	Dst    string `yaml:"dst" json:"dst"`
}
