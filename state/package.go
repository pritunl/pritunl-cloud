package state

import (
	"reflect"
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/pritunl/pritunl-cloud/database"
)

var (
	refCounter = 0
	registry   = map[PackageHandler]*Package{}
)

type PackageHandler interface {
	Refresh(pkg *Package, db *database.Database) (err error)
	Apply(st *State)
}

type Package struct {
	ref     int
	after   set.Set
	ttl     time.Duration
	handler PackageHandler
}

func (p *Package) Cache(d time.Duration) *Package {
	p.ttl = d
	return p
}

func (p *Package) Evict() *Package {
	p.ttl = 0
	return p
}

func (p *Package) After(handler PackageHandler) *Package {
	p.after.Add(registry[handler].ref)
	return p
}

func NewPackage(handler PackageHandler) *Package {
	refCounter += 1

	pkg := &Package{
		ref:     refCounter,
		handler: handler,
		after:   set.NewSet(),
	}
	registry[handler] = pkg

	return pkg
}

func RefreshAll(db *database.Database, runtimes *Runtimes) (err error) {
	inDegree := make(map[int]int)
	dependents := make(map[int][]*Package)
	refToPackage := make(map[int]*Package)

	for _, pkg := range registry {
		inDegree[pkg.ref] = 0
		dependents[pkg.ref] = []*Package{}
		refToPackage[pkg.ref] = pkg
	}

	for _, pkg := range registry {
		for afterRef := range pkg.after.Iter() {
			afterRefInt := afterRef.(int)
			inDegree[pkg.ref]++
			dependents[afterRefInt] = append(dependents[afterRefInt], pkg)
		}
	}

	ready := make(chan *Package, len(registry))
	for ref, degree := range inDegree {
		if degree == 0 {
			ready <- refToPackage[ref]
		}
	}

	done := make(chan *Package, len(registry))
	errors := make(chan error, len(registry))
	completed := make(map[int]bool)

	processPackage := func(pkg *Package) {
		go func() {
			defer func() {
				done <- pkg
			}()

			start := time.Now()
			refreshErr := pkg.handler.Refresh(pkg, db)
			dur := time.Since(start)
			if refreshErr != nil {
				errors <- refreshErr
			}

			structName := reflect.TypeOf(pkg.handler).Elem().Name()
			runtimes.SetState(structName, dur)
		}()
	}

	processed := 0
	totalPackages := len(registry)

	for processed < totalPackages {
		select {
		case pkg := <-ready:
			processPackage(pkg)

		case completedPkg := <-done:
			processed++
			completed[completedPkg.ref] = true

			for _, dependent := range dependents[completedPkg.ref] {
				if !completed[dependent.ref] {
					allDepsCompleted := true

					for afterRef := range dependent.after.Iter() {
						if !completed[afterRef.(int)] {
							allDepsCompleted = false
						}
					}

					if allDepsCompleted {
						ready <- dependent
					}
				}
			}

		case refreshErr := <-errors:
			if err == nil {
				err = refreshErr
			}
		}
	}

	select {
	case refreshErr := <-errors:
		if err == nil {
			err = refreshErr
		}
	default:
	}

	return
}

func ApplyAll(st *State) {
	for _, pkg := range registry {
		pkg.handler.Apply(st)
	}
}
