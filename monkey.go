package monkey

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/huandu/go-tls/g"
)

var (
	lock = sync.Mutex{}

	patches = make(map[uintptr]*patch)
)

type PatchGuard struct {
	target      reflect.Value
	replacement reflect.Value

	global  bool
	generic bool
}

func (g *PatchGuard) Unpatch() {
	unpatchValue(g.target)
}

func (g *PatchGuard) Restore() {
	patchValue(g.target, g.replacement, g.global, g.generic)
}

// Patch replaces a function with another for current goroutine only.
//
// Usage examples:
//   Patch(math.Abs, func(n float64) { return 0 })
//   Patch((*net.Dialer).Dial, func(_ *net.Dialer, _, _ string) (net.Conn, error) {})
func Patch(target, replacement interface{}, opts ...Option) *PatchGuard {
	t := reflect.ValueOf(target)
	r := reflect.ValueOf(replacement)

	o := opt{}

	for _, opt := range opts {
		opt.apply(&o)
	}

	patchValue(t, r, o.global, o.generic)

	return &PatchGuard{t, r, o.global, o.generic}
}

// See reflect.Value
type value struct {
	_   uintptr
	ptr unsafe.Pointer
}

func getPtr(v reflect.Value) unsafe.Pointer {
	return (*value)(unsafe.Pointer(&v)).ptr
}

func checkStructMonkeyType(a, b reflect.Type) bool {
	if a.NumIn() != b.NumIn() {
		return false
	}

	if a.NumIn() == 0 {
		return false
	}

	for i := 1; i < a.NumIn(); i++ {
		if a.In(i) != b.In(i) {
			return false
		}
	}

	t1 := a.In(0).String()
	t2 := b.In(0).String()

	if strings.Index(t2, "__monkey__") == -1 {
		return false
	}

	t1 = t1[strings.LastIndex(t1, "."):]
	t2 = t2[strings.LastIndex(t2, "."):]

	t2 = strings.Replace(t2, "__monkey__", "", 1)

	return t1 == t2
}

func patchValue(target, replacement reflect.Value, global, generic bool) {
	lock.Lock()
	defer lock.Unlock()

	if target.Kind() != reflect.Func {
		panic("target has to be a Func")
	}

	if replacement.Kind() != reflect.Func {
		panic("replacement has to be a Func")
	}

	if replacement.IsNil() {
		panic("replacement must not to be nil")
	}

	if target.Type() != replacement.Type() {
		if checkStructMonkeyType(target.Type(), replacement.Type()) {
			goto valid
		}

		panic(fmt.Sprintf(
			"target and replacement have to have the same type %s != %s",
			target.Type(), replacement.Type()))
	}

valid:

	if global {
		jumpData := jmpToGoFn((uintptr)(getPtr(replacement)))
		copyToLocation(target.Pointer(), jumpData)
		return
	}

	p, ok := patches[target.Pointer()]
	if !ok {
		p = &patch{from: target.Pointer(), generic: generic}
		patches[target.Pointer()] = p
	}

	if p.generic {
		p.Add(getFirstCallFunc(replacement.Pointer()))
	} else {
		p.Add((uintptr)(getPtr(replacement)))
	}

	p.Apply()
}

// PatchEmpty patches target with empty patch.
// Call the target will run the original func.
func PatchEmpty(target interface{}) {
	lock.Lock()
	defer lock.Unlock()

	t := reflect.ValueOf(target).Pointer()

	p, ok := patches[t]
	if ok {
		return
	}

	p = &patch{from: t}
	patches[t] = p
	p.Apply()
}

// Unpatch removes any monkey patches on target
// returns whether target was patched in the first place
func Unpatch(target interface{}) bool {
	return unpatchValue(reflect.ValueOf(target))
}

// UnpatchInstanceMethod removes the patch on methodName of the target
// returns whether it was patched in the first place
func UnpatchInstanceMethod(target reflect.Type, methodName string) bool {
	m, ok := target.MethodByName(methodName)
	if !ok {
		panic(fmt.Sprintf("unknown method %s", methodName))
	}
	return unpatchValue(m.Func)
}

// UnpatchAll removes all applied monkeypatches
func UnpatchAll() {
	lock.Lock()
	defer lock.Unlock()
	for _, p := range patches {
		p.patches = nil
		p.Apply()
	}
}

// Unpatch removes a monkeypatch from the specified function
// returns whether the function was patched in the first place
func unpatchValue(target reflect.Value) bool {
	lock.Lock()
	defer lock.Unlock()
	patch, ok := patches[target.Pointer()]
	if !ok {
		return false
	}

	return patch.Del()
}

func unpatch(target uintptr, p *patch) {
	copyToLocation(target, p.original)
}

type patch struct {
	from     uintptr
	realFrom uintptr

	original []byte
	patch    []byte

	patched bool
	generic bool

	// g pointer => patch func pointer
	patches map[uintptr]uintptr
}

func (p *patch) getFrom() uintptr {
	if !p.generic {
		return p.from
	}
	if p.realFrom == 0 {
		p.realFrom = getFirstCallFunc(p.from)
	}
	return p.realFrom
}

func (p *patch) Add(to uintptr) {
	if p.patches == nil {
		p.patches = make(map[uintptr]uintptr)
	}

	gid := (uintptr)(g.G())

	p.patches[gid] = to
}

func (p *patch) Del() bool {
	if p.patches == nil {
		return false
	}

	gid := (uintptr)(g.G())
	if _, ok := p.patches[gid]; !ok {
		return false
	}
	delete(p.patches, gid)
	p.Apply()
	return true
}

func (p *patch) Apply() {
	p.patch = p.Marshal()

	v := reflect.ValueOf(p.patch)
	allowExec(v.Pointer(), len(p.patch))

	if p.patched {
		data := littleEndian(v.Pointer())
		copyToLocation(p.getFrom()+2, data)
	} else {
		jumpData := jmpToFunctionValue(v.Pointer())
		copyToLocation(p.getFrom(), jumpData)
		p.patched = true
	}
}

func (p *patch) Marshal() (patch []byte) {
	if p.original == nil {
		p.original = alginPatch(p.getFrom())
	}

	patch = getg()

	for g, to := range p.patches {
		t := jmpTable(g, to, !p.generic)
		patch = append(patch, t...)
	}

	patch = append(patch, p.original...)
	old := jmpToFunctionValue(p.getFrom() + uintptr(len(p.original)))
	patch = append(patch, old...)

	return
}
