package task2make

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"gopkg.in/yaml.v3"
)

type Taskfile struct {
	Version string `yaml:"version"`
	Vars    Vars   `yaml:"vars"`
	// Includes map[string]*TaskfileRef `yaml:"includes"`
	Tasks Tasks `yaml:"tasks"`
}

func (t Taskfile) WriteMakefile(w io.Writer) error {
	// err := t.Vars.WriteMakefile(w)
	// if err != nil {
	// 	return fmt.Errorf("write vars: %w", err)
	// }
	for _, v := range t.Vars {
		fmt.Fprintf(w, "%s=", v.Name)
		err := v.Body.WriteMakefile(w)
		if err != nil {
			return fmt.Errorf("write %q var: %w", v.Name, err)
		}
		fmt.Fprintln(w)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w)

	err := t.Tasks.WriteMakefile(w)
	if err != nil {
		return fmt.Errorf("write tasks: %w", err)
	}

	return nil
}

type Vars Map2List[Scalarable[VarContent]]

// func (v Vars) WriteMakefile(w io.Writer) error {
// 	for _, vr := range v {
// 		fmt.Fprintf(w, "%s=", vr.Name)
// 		err := vr.Body.WriteMakefile(w)
// 		if err != nil {
// 			return fmt.Errorf("write var body: %w", err)
// 		}
// 		fmt.Fprintln(w)

// 		// err := vr.WriteMakefile(w)
// 		// if err != nil {
// 		// 	return fmt.Errorf("write %q var: %w", vr.Varname, err)
// 		// }
// 	}
// 	return nil
// }

var _ yaml.Unmarshaler = (*Vars)(nil)

func (v *Vars) UnmarshalYAML(n *yaml.Node) error {
	return (*Map2List[Scalarable[VarContent]])(v).UnmarshalYAML(n)
}

// type Vars []*Variable

type Map2List[T any] []*Named[T]

type Named[T any] struct {
	Name string
	Body T
}

// func (n *Named[T]) DecodeNameContent() (name, content interface{})

var _ yaml.Unmarshaler = (*Map2List[struct{}])(nil)

func (m *Map2List[T]) UnmarshalYAML(n *yaml.Node) error {
	for i := 0; i+1 < len(n.Content); i += 2 {
		var v *Named[T] = &Named[T]{}
		// pt := &v
		// name, content := v.DecodeNameContent()

		err := n.Content[i].Decode(&v.Name)
		if err != nil {
			return fmt.Errorf("unmarshal name: %w", err)
		}
		// log.Printf("task: %+v", v)

		err = n.Content[i+1].Decode(&v.Body)
		if err != nil {
			log.Printf("unmarshal body: %+v", n.Content[i+1])
			return fmt.Errorf("unmarshal body: %w", err)
		}

		*m = append(*m, v)
	}

	// log.Println("name", n.Content[0].Value)
	// log.Printf("value %+v\n", n.Content[1])
	// for _, c := range n.Content {
	// 	fmt.Printf("%q\n", c.Value)
	// }
	return nil
}

func unmarshalMap2SliceYAML[F any, T interface {
	DecodeNameContent() (name, content interface{})
	*F
}, S ~[]T](n *yaml.Node, slice *S) error {
	for i := 0; i+1 < len(n.Content); i += 2 {
		var v T = new(F)
		// pt := &v
		name, content := v.DecodeNameContent()

		err := n.Content[i].Decode(name)
		if err != nil {
			return fmt.Errorf("unmarshal name: %w", err)
		}
		// log.Printf("task: %+v", v)

		err = n.Content[i+1].Decode(content)
		if err != nil {
			log.Printf("n.Content[i+1]: %+v", n.Content[i+1])
			return fmt.Errorf("unmarshal content: %w", err)
		}

		*slice = append(*slice, v)
	}

	// log.Println("name", n.Content[0].Value)
	// log.Printf("value %+v\n", n.Content[1])
	// for _, c := range n.Content {
	// 	fmt.Printf("%q\n", c.Value)
	// }
	return nil
}

// var _ yaml.Unmarshaler = (*Vars)(nil)

// func (v *Vars) UnmarshalYAML(n *yaml.Node) error {
// 	err := unmarshalMap2SliceYAML(n, v)
// 	if err != nil {
// 		return fmt.Errorf("unmarshal vars: %w", err)
// 	}
// 	return nil
// }

// type Variable struct {
// 	Varname
// 	VarContent Scalarable[*VarContent]
// }

// func (v *Variable) DecodeNameContent() (name, content interface{}) {
// 	return &v.Varname, &v.VarContent
// }

// func (v Variable) WriteMakefile(w io.Writer) error {
// 	fmt.Fprintf(w, "%s=", v.Varname)
// 	return v.VarContent.WriteMakefile(w)
// 	// if v.VarContent.Scalar != "" {
// 	// 	fmt.Fprint(w, v.VarContent.Scalar)
// 	// } else {
// 	// 	fmt.Fprintf(w, "$(shell %s)", v.VarContent.T.ShellCommand)
// 	// }
// 	// return nil
// }

type Varname string

type VarContent struct {
	ShellCommand ShellCommand `yaml:"sh"`
}

func (v VarContent) WriteMakefile(w io.Writer) error {
	// fmt.Fprintf(w, "$(shell %s)", v.ShellCommand)
	return v.ShellCommand.WriteMakefile(w)
}

type Tasks Map2List[TaskContent]

var _ yaml.Unmarshaler = (*Tasks)(nil)

func (t *Tasks) UnmarshalYAML(n *yaml.Node) error {
	// err := unmarshalMap2SliceYAML(n, t)
	// if err != nil {
	// 	return fmt.Errorf("unmarshal tasks: %w", err)
	// }
	return (*Map2List[TaskContent])(t).UnmarshalYAML(n)
}

func (t Tasks) WriteMakefile(w io.Writer) error {
	for _, task := range t {
		err := writeTask(w, task.Name, task.Body)
		// err := task.WriteMakefile(w)
		if err != nil {
			return fmt.Errorf("write %q task: %w", task.Name, err)
		}

		fmt.Fprintln(w)
	}
	return nil
}

type TaskContent struct {
	Deps        []Taskname    `yaml:"deps"`
	Description string        `yaml:"desc"`
	Dir         string        `yaml:"dir"`
	Envs        Envs          `yaml:"env"`
	Vars        Vars          `yaml:"vars"`
	Commands    []TaskCommand `yaml:"cmds"`
}

func writeTask(w io.Writer, taskname string, t TaskContent) error {
	// comment
	if t.Description != "" {
		fmt.Fprintf(w, "# %s\n", t.Description)
	}

	// export envs
	for _, e := range t.Envs {
		if string(e.Body.T.ShellCommand) == fmt.Sprintf("echo {{.%s}}", e.Name) {
			continue
		}
		fmt.Fprintf(w, "%s: export %s=", taskname, e.Name)
		err := e.Body.WriteMakefile(w)
		if err != nil {
			return fmt.Errorf("write %s env: %w", e.Name, err)
		}
		fmt.Fprintln(w)
	}

	// vars
	// err := t.Vars.WriteMakefile(w)
	// if err != nil {
	// 	return fmt.Errorf("write vars: %w", err)
	// }
	for _, v := range t.Vars {
		if string(v.Body.T.ShellCommand) == fmt.Sprintf("echo $%s", v.Name) {
			continue
		}
		fmt.Fprintf(w, "%s: %s=", taskname, v.Name)
		err := v.Body.WriteMakefile(w)
		if err != nil {
			return fmt.Errorf("write %s var: %w", v.Name, err)
		}
		fmt.Fprintln(w)
	}

	if len(t.Vars) > 0 {
		w = ReplaceWriter{
			Vars: t.Vars,
			w:    w,
		}
	}

	// target: deps
	fmt.Fprintf(w, "%s:", taskname)
	for _, deps := range t.Deps {
		fmt.Fprintf(w, " %s", deps)
	}
	fmt.Fprintln(w)

	// commands
	if len(t.Commands) > 0 {
		for i, command := range t.Commands {
			fmt.Fprint(w, "\t")
			if command.Cmd.IsSilent {
				fmt.Fprint(w, "@ ")
			}
			if t.Dir != "" {
				fmt.Fprintf(w, "cd %s; ", t.Dir)
			}
			err := command.WriteMakefile(w)
			if err != nil {
				return fmt.Errorf("write %d command: %w", i, err)
			}
			fmt.Fprintln(w)
		}
	}
	return nil
}

type Taskname string

type Task struct {
	Taskname
	TaskContent
}

// func (t *Task) DecodeNameContent() (name, content interface{}) {
// 	return &t.Taskname, &t.TaskContent
// }

// func (t Task) WriteMakefile(w io.Writer) error {
// 	// comment
// 	if t.Description != "" {
// 		fmt.Fprintf(w, "# %s\n", t.Description)
// 	}

// 	// export envs
// 	for _, e := range t.Envs {
// 		fmt.Fprintf(w, "%s: export %s=", t.Taskname, e.Name)
// 		err := e.WriteMakefile(w)
// 		if err != nil {
// 			return fmt.Errorf("write %s env: %w", e.Name, err)
// 		}
// 		fmt.Fprintln(w)
// 	}

// 	// vars
// 	err := t.Vars.WriteMakefile(w)
// 	if err != nil {
// 		return fmt.Errorf("write vars: %w", err)
// 	}
// 	// for _, v := range t.Vars {
// 	// 	fmt.Fprintf(w, "%s: ", t.Taskname)
// 	// 	err := v.WriteMakefile(w)
// 	// 	if err != nil {
// 	// 		return fmt.Errorf("write %s var: %w", v.Varname, err)
// 	// 	}
// 	// 	fmt.Fprintln(w)
// 	// }

// 	if len(t.Vars) > 0 {
// 		w = ReplaceWriter{
// 			Vars: t.Vars,
// 			w:    w,
// 		}
// 	}

// 	// target: deps
// 	fmt.Fprintf(w, "%s:", t.Taskname)
// 	for _, deps := range t.Deps {
// 		fmt.Fprintf(w, " %s", deps)
// 	}
// 	fmt.Fprintln(w)

// 	// commands
// 	if len(t.Commands) > 0 {
// 		for i, command := range t.Commands {
// 			fmt.Fprint(w, "\t")
// 			if command.Cmd.IsSilent {
// 				fmt.Fprint(w, "@ ")
// 			}
// 			if t.Dir != "" {
// 				fmt.Fprintf(w, "cd %s; ", t.Dir)
// 			}
// 			err := command.WriteMakefile(w)
// 			if err != nil {
// 				return fmt.Errorf("write %d command: %w", i, err)
// 			}
// 			fmt.Fprintln(w)
// 		}
// 	}
// 	return nil
// }

type ReplaceWriter struct {
	Vars
	w io.Writer
}

func (r ReplaceWriter) Write(p []byte) (n int, _ error) {
	log.Printf("p: %+v", string(p))
	for _, v := range r.Vars {
		from := []byte(fmt.Sprintf("{{.%s}}", v.Name))
		log.Printf("from: %+v", string(from))
		to := []byte(fmt.Sprintf("${%s}", v.Name))
		log.Printf("to: %+v", string(to))
		p = bytes.ReplaceAll(p, from, to)
	}
	return r.w.Write(p)
}

type Envs Map2List[Scalarable[EnvContent]]

var _ yaml.Unmarshaler = (*Envs)(nil)

func (e *Envs) UnmarshalYAML(n *yaml.Node) error {
	// err := unmarshalMap2SliceYAML(n, e)
	// if err != nil {
	// 	return fmt.Errorf("unmarshal envs: %w", err)
	// }

	return (*Map2List[Scalarable[EnvContent]])(e).UnmarshalYAML(n)
}

// func (e *Envs) WriteMakefile(w io.Writer) error {
// 	for _, env := range *e {
// 		fmt.Fprintf(w, "%s=", env.Name)
// 		err := env.Body.WriteMakefile(w)
// 		if err != nil {
// 			return fmt.Errorf("write %q env: %w", env.Name, err)
// 		}
// 	}
// 	return nil
// }

type Env struct {
	Name       string `yaml:"-"`
	EnvContent Scalarable[EnvContent]
}

// var _ yaml.Unmarshaler = (*Env)(nil)

// func (e *Env) UnmarshalYAML(n *yaml.Node) error {
// }

// func (e *Env) DecodeNameContent() (name, content interface{}) {
// 	return &e.Name, &e.EnvContent
// }

// func (e *Env) WriteMakefile(w io.Writer) error {
// 	fmt.Fprintf(w, "%s=", e.Name)
// 	return e.EnvContent.WriteMakefile(w)
// }

// var _ yaml.Unmarshaler = (*Env)(nil)

// func (e *Env) UnmarshalYAML(n *yaml.Node) error {
// 	switch n.Kind {
// 	case yaml.ScalarNode:
// 		e.Value. = n.Value
// 		return nil
// 	default:
// 	}
// 	err := n.Decode(&e.EnvContent)
// 	if err != nil {
// 		log.Printf("%+v", n)
// 		log.Println(n.Content)
// 		for _, c := range n.Content {
// 			log.Printf("%v", c.Value)
// 		}
// 		log.Printf("%q", n.Value)
// 		panic("not implemented")
// 	}
// 	return nil
// }

type EnvContent struct {
	ShellCommand ShellCommand `yaml:"sh"`
}

func (e EnvContent) WriteMakefile(w io.Writer) error {
	// fmt.Fprintf(w, "$(shell %s)", e.ShellCommand)
	return e.ShellCommand.WriteMakefile(w)
}

type ShellCommand string

func (s ShellCommand) WriteMakefile(w io.Writer) error {
	s = ShellCommand(strings.ReplaceAll(string(s), "$", "$$"))
	s = ShellCommand(strings.ReplaceAll(string(s), "{{.", "${"))
	s = ShellCommand(strings.ReplaceAll(string(s), "}}", "}"))
	fmt.Fprintf(w, "$(shell %s)", s)
	return nil
}

type TaskCommand struct {
	Cmd TaskCommandCmd
}

func (t TaskCommand) WriteMakefile(w io.Writer) error {
	switch {
	case t.Cmd.Command != "":
		// cmd := strings.ReplaceAll(t.Cmd.Command, "\n", ";\\\n\t")
		return t.Cmd.Command.WriteMakefile(w)
		// fmt.Fprintf(w, "%s", cmd)
	case t.Cmd.TaskRef != "":
		fmt.Fprintf(w, "$(MAKE) %s", t.Cmd.TaskRef)
	}
	return nil
}

var _ yaml.Unmarshaler = (*TaskCommand)(nil)

func (t *TaskCommand) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		t.Cmd.Command = MakeCommand(n.Value)
		return nil
	default:
	}
	err := n.Decode(&t.Cmd)
	if err != nil {
		log.Printf("%+v", n)
		log.Println(n.Content)
		for _, c := range n.Content {
			log.Printf("%v", c.Value)
		}
		log.Printf("%q", n.Value)
		panic("not implemented")
	}
	return nil
}

type TaskCommandCmd struct {
	Command  MakeCommand `yaml:"cmd"`
	IsSilent bool        `yaml:"silent"`
	TaskRef  string      `yaml:"task"`
}

type MakeCommand string

func (m MakeCommand) WriteMakefile(w io.Writer) error {
	s := string(m)
	for {
		ifIdx := strings.Index(s, "{{if")
		if ifIdx == -1 {
			break
		}
		ifEndIdx := strings.Index(s[ifIdx:], "}}") + ifIdx
		elseIdx := strings.Index(s[ifEndIdx:], "{{else}}") + ifEndIdx
		endIdx := strings.Index(s[elseIdx:], "{{end}}") + elseIdx
		if ifIdx == 0 {
			log.Printf("ifIdx: %+v", ifIdx)
			log.Printf("ifEndIdx: %+v", ifEndIdx)
			log.Printf("elseIdx: %+v", elseIdx)
			log.Printf("endIdx: %+v", endIdx)
			log.Printf("s: %+v", s)
			o := "ifneq (${" + s[ifIdx+4:ifEndIdx] + "},)\n"
			o += "\t" + s[ifEndIdx+2:elseIdx] + "\n"
			o += "else\n"
			o += "\t" + s[elseIdx+8:endIdx] + "\n"
			o += "endif\n"
			fmt.Fprint(w, o)
			return nil
		}
		// 2023/04/29 19:31:51 ifIdx: 0
		// 2023/04/29 19:31:51 ifEndIdx: 24
		// 2023/04/29 19:31:51 elseIdx: 32
		// 2023/04/29 19:31:51 endIdx: 60
		// 2023/04/29 19:31:51 s: {{if .DRONE_PULL_REQUEST}}danger{{else}}echo Skipping danger{{end}}
		s = s[:ifIdx] + "$(if ${" + s[ifIdx+4:ifEndIdx] + "}," + s[ifEndIdx+2:elseIdx] + "," + s[elseIdx+8:endIdx] + ")" + s[endIdx+7:]
	}
	s = strings.ReplaceAll(s, "\n", ";\\\n\t")
	s = strings.ReplaceAll(s, "{{.", "${")
	s = strings.ReplaceAll(s, "}}", "}")
	fmt.Fprint(w, s)
	return nil
}

type ScalarableMakefileWriter interface {
	WriteMakefile(w io.Writer) error
}

type Scalarable[T ScalarableMakefileWriter] struct {
	Scalar string
	T      T
}

var _ yaml.Unmarshaler = (*Scalarable[ScalarableMakefileWriter])(nil)

func (s *Scalarable[T]) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.ScalarNode:
		n.Value = strings.ReplaceAll(n.Value, "{{.", "${")
		n.Value = strings.ReplaceAll(n.Value, "}}", "}")
		s.Scalar = n.Value
		return nil
	case yaml.MappingNode:
		err := n.Decode(&s.T)
		if err != nil {
			// log.Printf("%+v", n)
			// log.Println(n.Content)
			// for _, c := range n.Content {
			// 	log.Printf("%v", c.Value)
			// }
			// log.Printf("%q", n.Value)
			return fmt.Errorf("unmarshal yaml.MappingNode")
		}
		return nil
	default:
		panic("not implemented")
	}
}

func (s *Scalarable[T]) WriteMakefile(w io.Writer) error {
	switch {
	case s.Scalar != "":
		_, err := fmt.Fprint(w, s.Scalar)
		if err != nil {
			return fmt.Errorf("write scalar: %w", err)
		}
		return nil
	default:
		return s.T.WriteMakefile(w)
	}
}
