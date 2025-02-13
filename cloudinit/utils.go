package cloudinit

import (
	"bytes"
	"encoding/base64"
	"text/template"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-cloud/errortypes"
)

type fileData struct {
	Content     string
	Owner       string
	Path        string
	Permissions string
}

type writeFileData struct {
	Files []*fileData
}

const writeFileTmpl = `{{range .Files}}
{{- if eq .Content ""}}
  - content: ""
{{- else}}
  - encoding: base64
    content: {{.Content}}
{{- end}}
    owner: {{.Owner}}
    path: {{.Path}}
    permissions: "{{.Permissions}}"
{{- end}}`

var (
	writeFile = template.Must(template.New("write_file").Parse(writeFileTmpl))
)

func generateWriteFiles(filesData []*fileData) (output string, err error) {
	for _, file := range filesData {
		file.Content = base64.StdEncoding.EncodeToString([]byte(file.Content))
	}

	data := writeFileData{
		Files: filesData,
	}

	outputBuf := &bytes.Buffer{}

	err = writeFile.Execute(outputBuf, data)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "cloudinit: Failed to exec write file template"),
		}
		return
	}

	output = outputBuf.String()

	return
}
