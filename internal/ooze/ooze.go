package ooze

import (
	"github.com/gtramontina/ooze/internal/goinfectedfile"
	"github.com/gtramontina/ooze/internal/gomutatedfile"
	"github.com/gtramontina/ooze/internal/gosourcefile"
	"github.com/gtramontina/ooze/internal/result"
	"github.com/gtramontina/ooze/internal/viruses"
)

type Logger interface {
	Logf(message string, args ...any)
}

type Repository interface {
	ListGoSourceFiles() []*gosourcefile.GoSourceFile
	LinkAllToTemporaryRepository(temporaryPath string) TemporaryRepository
}

type TemporaryRepository interface {
	Root() string
	Overwrite(filePath string, data []byte)
	Remove()
}

type Laboratory interface {
	Test(repository Repository, file *gomutatedfile.GoMutatedFile) <-chan result.Result[string]
}

type Diagnostic struct {
	res  <-chan result.Result[string]
	file *gomutatedfile.GoMutatedFile
}

func (d *Diagnostic) IsOk() bool {
	return (<-d.res).IsOk()
}

func NewDiagnostic(res <-chan result.Result[string], file *gomutatedfile.GoMutatedFile) *Diagnostic {
	return &Diagnostic{
		res:  res,
		file: file,
	}
}

type Reporter interface {
	AddDiagnostic(diagnostic *Diagnostic)
	Summarize()
}

type Ooze struct {
	repository Repository
	laboratory Laboratory
	reporter   Reporter
}

func New(repository Repository, laboratory Laboratory, reporter Reporter) *Ooze {
	return &Ooze{
		repository: repository,
		laboratory: laboratory,
		reporter:   reporter,
	}
}

func (o *Ooze) Release(viri ...viruses.Virus) {
	sources := o.repository.ListGoSourceFiles()

	var incubated []*goinfectedfile.GoInfectedFile

	for _, source := range sources {
		for _, virus := range viri {
			incubated = append(incubated, source.Incubate(virus)...)
		}
	}

	if len(incubated) == 0 {
		return
	}

	for _, infectedFile := range incubated {
		mutatedFile := infectedFile.Mutate()
		res := o.laboratory.Test(o.repository, mutatedFile)
		o.reporter.AddDiagnostic(NewDiagnostic(res, mutatedFile))
	}
}
