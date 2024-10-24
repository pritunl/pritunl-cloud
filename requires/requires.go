// Init system with before and after constraints.
package requires

import (
	"container/list"
	"fmt"
	"os"

	"github.com/dropbox/godropbox/container/set"
)

var (
	modules = list.New()
)

type Module struct {
	name    string
	before  set.Set
	after   set.Set
	Handler func() (err error)
}

func (m *Module) Before(name string) {
	m.before.Add(name)
}

func (m *Module) After(name string) {
	m.after.Add(name)
}

func New(name string) (module *Module) {
	module = &Module{
		name:   name,
		before: set.NewSet(),
		after:  set.NewSet(),
	}
	modules.PushBack(module)
	return
}

func Init(ignore []string) {
	loaded := false
	ignoreSet := set.NewSet()

	if ignore != nil {
		for _, name := range ignore {
			ignoreSet.Add(name)
		}
	}

Loop:
	for count := 0; count < 100; count += 1 {
		i := modules.Front()
		for i != nil {
			module := i.Value.(*Module)

			j := i.Prev()
			for j != nil {
				if module.before.Contains(j.Value.(*Module).name) {
					modules.MoveBefore(i, j)
					continue Loop
				}
				j = j.Prev()
			}

			j = i.Next()
			for j != nil {
				if module.after.Contains(j.Value.(*Module).name) {
					modules.MoveAfter(i, j)
					continue Loop
				}
				j = j.Next()
			}

			i = i.Next()
		}

		loaded = true
		break Loop
	}

	if !loaded {
		fmt.Fprint(os.Stderr, "Requires failed to satisfy constraints\n")

		i := modules.Front()
		for i != nil {
			module := i.Value.(*Module)
			line := module.name

			for val := range module.before.Iter() {
				line += fmt.Sprintf("   before: %s", val.(string))
			}
			for val := range module.after.Iter() {
				line += fmt.Sprintf("   after: %s", val.(string))
			}

			fmt.Fprint(os.Stderr, line+"\n")
			i = i.Next()
		}

		i = modules.Front()
	Loop2:
		for i != nil {
			module := i.Value.(*Module)

			j := i.Prev()
			for j != nil {
				val := j.Value.(*Module).name
				if module.before.Contains(val) {
					fmt.Fprintf(os.Stderr, "'%s' not before '%s'\n",
						module.name, val)
					break Loop2
				}
				j = j.Prev()
			}

			j = i.Next()
			for j != nil {
				val := j.Value.(*Module).name
				if module.after.Contains(val) {
					fmt.Fprintf(os.Stderr, "'%s' not after '%s'\n",
						module.name, val)
					break Loop2
				}
				j = j.Next()
			}

			i = i.Next()
		}

		os.Exit(1)
	}

	i := modules.Front()
	for i != nil {
		if !ignoreSet.Contains(i.Value.(*Module).name) {
			err := i.Value.(*Module).Handler()
			if err != nil {
				panic(err)
			}
		}

		i = i.Next()
	}
}
