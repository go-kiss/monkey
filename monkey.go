package monkey // import "bou.ke/monkey"

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/huandu/go-tls/g"
)

// patch is an applied patch
// needed to undo a patch
type patch struct {
	originalBytes []byte
	replacement   *reflect.Value
}

var (
	lock = sync.Mutex{}

	patches = make(map[uintptr]patch)

	patches2 = make(map[uintptr]*gPatch)
)

type PatchGuard struct {
	target      reflect.Value
	replacement reflect.Value
}

func (g *PatchGuard) Unpatch() {
	unpatchValue(g.target)
}

func (g *PatchGuard) Restore() {
	patchValue(g.target, g.replacement)
}

// Patch replaces a function with another
func Patch(target, replacement interface{}) *PatchGuard {
	t := reflect.ValueOf(target)
	r := reflect.ValueOf(replacement)
	patchValue(t, r)

	return &PatchGuard{t, r}
}

// PatchInstanceMethod replaces an instance method methodName for the type target with replacement
// Replacement should expect the receiver (of type target) as the first argument
func PatchInstanceMethod(target reflect.Type, methodName string, replacement interface{}) *PatchGuard {
	m, ok := target.MethodByName(methodName)
	if !ok {
		panic(fmt.Sprintf("unknown method %s", methodName))
	}
	r := reflect.ValueOf(replacement)
	patchValue(m.Func, r)

	return &PatchGuard{m.Func, r}
}

func patchValue(target, replacement reflect.Value) {
	lock.Lock()
	defer lock.Unlock()

	if target.Kind() != reflect.Func {
		panic("target has to be a Func")
	}

	if replacement.Kind() != reflect.Func {
		panic("replacement has to be a Func")
	}

	if target.Type() != replacement.Type() {
		panic(fmt.Sprintf("target and replacement have to have the same type %s != %s", target.Type(), replacement.Type()))
	}

	p, ok := patches2[target.Pointer()]
	if !ok {
		p = &gPatch{From: target.Pointer()}
	}
	p.Add(replacement.Pointer())
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
	for target, p := range patches {
		unpatch(target, p)
		delete(patches, target)
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
	unpatch(target.Pointer(), patch)
	delete(patches, target.Pointer())
	return true
}

func unpatch(target uintptr, p patch) {
	copyToLocation(target, p.originalBytes)
}

type gPatch struct {
	From uintptr

	original []byte
	patch    []byte

	// g pointer => patch func pointer
	patches map[uintptr]uintptr

	m sync.Mutex

	prev *gPatch
}

func (p *gPatch) Add(to uintptr) {
	p.m.Lock()
	defer p.m.Unlock()

	if p.patches == nil {
		p.patches = make(map[uintptr]uintptr)
	}

	gid := (uintptr)(g.G())

	if _, ok := p.patches[gid]; ok {
		panic("exists")
	}

	p.patches[gid] = to
}

func (p *gPatch) Apply() {
	p.patch = p.Marshal()

	// dump("apply patch", p.patch)

	v := reflect.ValueOf(p.patch)
	setX(v.Pointer(), len(p.patch))

	jumpData := jmpToFunctionValue(v.Pointer())
	copyToLocation(p.From, jumpData)
}

func (p *gPatch) Marshal() (patch []byte) {
	if p.original == nil {
		p.original = alginPatch(p.From)
	}

	patch = getg()

	for g, to := range p.patches {
		t := jmpTable(g, to)
		patch = append(patch, t...)
	}

	patch = append(patch, p.original...)
	old := jmpToFunctionValue(p.From + uintptr(len(p.original)))
	patch = append(patch, old...)

	return
}
