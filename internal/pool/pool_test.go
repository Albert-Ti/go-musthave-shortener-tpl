package pool_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/Albert-Ti/go-musthave-shortener-tpl/internal/pool"
	"github.com/stretchr/testify/assert"
)

func ExamplePool() {
	bufPool := pool.New(func() *bytes.Buffer {
		return new(bytes.Buffer)
	})

	b := bufPool.Get()
	b.WriteString("2026-08-01T07:29:21Z")
	b.WriteByte(' ')
	b.WriteString("path")
	b.WriteByte('=')
	b.WriteString("/search?q=flowers")
	os.Stdout.Write(b.Bytes())

	bufPool.Put(b)

	// Output:
	// 2026-08-01T07:29:21Z path=/search?q=flowers
}

type testStruct struct{}

func (t *testStruct) Reset() {}

func TestPool(t *testing.T) {
	t.Run("object nil", func(t *testing.T) {
		newPool := pool.New[*testStruct](nil)
		obj := newPool.Get()

		assert.Nil(t, obj)
	})

	t.Run("creates object", func(t *testing.T) {
		p := pool.New(func() *testStruct {
			return &testStruct{}
		})

		assert.NotNil(t, p.Get())
	})
}

func BenchmarkPool(b *testing.B) {
	fixedTime := "2026-08-01T07:29:21Z"
	b.Run("Pool", func(b *testing.B) {

		bufPool := pool.New(func() *bytes.Buffer {
			return new(bytes.Buffer)
		})

		for b.Loop() {
			obj := bufPool.Get()
			obj.WriteString(fixedTime)
			obj.WriteByte(' ')
			obj.WriteString("path")
			obj.WriteByte('=')
			obj.WriteString("/search?q=flowers")

			bufPool.Put(obj)
		}
	})

	b.Run("No Pool", func(b *testing.B) {
		for b.Loop() {
			obj := new(bytes.Buffer)

			obj.Reset()
			obj.WriteString(fixedTime)
			obj.WriteByte(' ')
			obj.WriteString("path")
			obj.WriteByte('=')
			obj.WriteString("/search?q=flowers")
		}
	})
}
