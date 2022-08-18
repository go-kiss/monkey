package monkey_test

import (
	"math"
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/go-kiss/monkey"
	"github.com/go-kiss/monkey/demo"
)

func no() bool  { return false }
func yes() bool { return true }

func TestTimePatch(t *testing.T) {
	before := time.Now()
	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	})
	during := time.Now()
	assert(t, monkey.Unpatch(time.Now))
	after := time.Now()

	assert(t, time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC) == during)
	assert(t, before != during)
	assert(t, during != after)
}

func TestGC(t *testing.T) {
	value := true
	monkey.Patch(no, func() bool {
		return value
	})
	defer monkey.UnpatchAll()
	runtime.GC()
	assert(t, no())
}

func TestSimple(t *testing.T) {
	assert(t, !no())
	monkey.Patch(no, yes)
	assert(t, no())
	assert(t, monkey.Unpatch(no))
	assert(t, !no())
	assert(t, !monkey.Unpatch(no))
}

func TestGuard(t *testing.T) {
	var guard *monkey.PatchGuard
	guard = monkey.Patch(no, func() bool {
		guard.Unpatch()
		defer guard.Restore()
		return !no()
	})
	for i := 0; i < 100; i++ {
		assert(t, no())
	}
	monkey.Unpatch(no)
}

func TestUnpatchAll(t *testing.T) {
	assert(t, !no())
	monkey.Patch(no, yes)
	assert(t, no())
	monkey.UnpatchAll()
	assert(t, !no())
}

type s struct{}

func (s *s) yes() bool { return true }

func TestWithInstanceMethod(t *testing.T) {
	i := &s{}

	assert(t, !no())
	monkey.Patch(no, i.yes)
	assert(t, no())
	monkey.Unpatch(no)
	assert(t, !no())
}

type f struct{}

func (f *f) No() bool { return false }

func TestOnInstanceMethod(t *testing.T) {
	i := &f{}
	assert(t, !i.No())
	monkey.Patch((*f).No, func(_ *f) bool { return true })
	assert(t, i.No())
	assert(t, monkey.UnpatchInstanceMethod(reflect.TypeOf(i), "No"))
	assert(t, !i.No())
}

func TestNotFunction(t *testing.T) {
	panics(t, func() {
		monkey.Patch(no, 1)
	})
	panics(t, func() {
		monkey.Patch(1, yes)
	})
}

func TestNotCompatible(t *testing.T) {
	panics(t, func() {
		monkey.Patch(no, func() {})
	})
}

func assert(t *testing.T, b bool, args ...interface{}) {
	t.Helper()
	if !b {
		t.Fatal(append([]interface{}{"assertion failed"}, args...))
	}
}

func panics(t *testing.T, f func()) {
	t.Helper()
	defer func() {
		t.Helper()
		if v := recover(); v == nil {
			t.Fatal("expected panic")
		}
	}()
	f()
}

//go:noinline
func foo(a, b int) int {
	return a + b
}

//go:noinline
func bar(a, b int) int {
	return a - b
}

func TestEmpty(t *testing.T) {
	monkey.PatchEmpty(foo)

	stop := make(chan int)
	var wg sync.WaitGroup
	for i := 1; i < runtime.NumCPU()-1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			j := 0
			for {
				f := func(j int) func(a, b int) int {
					return func(a, b int) int {
						return a + b + j
					}
				}(j)
				monkey.Patch(foo, f)
				assert(t, 1+1+j == foo(1, 1))
				monkey.Unpatch(foo)
				select {
				case <-stop:
					return
				default:
					j++
				}
			}
		}(i)
	}

	for i := 0; i < 2000000; i++ {
		assert(t, i+1 == foo(i, 1))
	}
	close(stop)
	wg.Wait()
}

func TestG(t *testing.T) {
	monkey.Patch(foo, bar)

	assert(t, -1 == foo(1, 2))
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		assert(t, 3 == foo(1, 2))
	}()
	go func() {
		monkey.Patch(foo, func(a, b int) int {
			return a * b
		})
		defer wg.Done()
		assert(t, 2 == foo(1, 2))
	}()
	wg.Wait()
}

func TestGlobal(t *testing.T) {
	monkey.Patch(math.Abs, func(a float64) float64 {
		return a + 1
	}, monkey.OptGlobal)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert(t, 2 == math.Abs(1))
	}()
	wg.Wait()
}

func add2[T int | float64](i, j T) T {
	return i - j
}

type S2__monkey__[T int | float64] struct{ demo.S2[T] }

func (s *S2__monkey__[T]) Foo() T { return s.I * 2 }

func TestGeneric(t *testing.T) {
	g1 := monkey.Patch(demo.Add[int], add2[int], monkey.OptGeneric)
	assert(t, -1 == demo.Add(1, 2))
	g1.Unpatch()
	assert(t, 3 == demo.Add(1, 2))

	g2 := monkey.Patch((*demo.S2[int]).Foo, (*S2__monkey__[int]).Foo, monkey.OptGeneric)
	s := demo.S2[int]{I: 2}
	assert(t, 4 == s.Foo())
	g2.Unpatch()
	assert(t, 2 == s.Foo())
}
