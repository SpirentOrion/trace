package yamlrec

import (
	"fmt"
	"io"

	"github.com/SpirentOrion/trace"
	"gopkg.in/yaml.v2"
)

type YAMLRecorder struct {
	io.Writer
}

var _ trace.Recorder = &YAMLRecorder{}

func New(writer io.Writer) (*YAMLRecorder, error) {
	return &YAMLRecorder{writer}, nil
}

func (r *YAMLRecorder) String() string {
	return "yaml"
}

func (r *YAMLRecorder) Record(s *trace.Span) error {
	buf, err := yaml.Marshal(s)
	if err != nil {
		return err
	}

	fmt.Fprintln(r, "---") // document separator
	_, err = r.Write(buf)
	return err
}
