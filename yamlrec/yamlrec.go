package yamlrec

import (
	"fmt"
	"io"

	"github.com/SpirentOrion/trace"
	"gopkg.in/yaml.v2"
)

type YAMLRecorder struct {
	writer io.Writer
}

func New(writer io.Writer) (*YAMLRecorder, error) {
	return &YAMLRecorder{writer: writer}, nil
}

func (r *YAMLRecorder) String() string {
	return "yaml"
}

func (r *YAMLRecorder) Start(s *trace.Span) error {
	// Intentionally left unimplemented -- YAMLRecorder only writes new documents as spans finish
	return nil
}

func (r *YAMLRecorder) Finish(s *trace.Span) error {
	buf, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	fmt.Fprintln(r.writer, "---") // document separator
	_, err = r.writer.Write(buf)
	return err
}
