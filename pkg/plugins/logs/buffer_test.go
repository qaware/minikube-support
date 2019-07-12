package logs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuffer_WriteGet(t *testing.T) {
	b := newBuffer()
	var written []interface{}
	next := 1

	t.Run("initial write", func(t *testing.T) {
		written = append(written, next)
		b.Write(next)

		entries := b.GetEntries()
		assert.Equal(t, written, filterNil(entries))
	})

	t.Run("write full", func(t *testing.T) {
		for i := uint64(len(written)); i < size; i++ {
			next = int(i)
			written = append(written, next)
			b.Write(next)
		}

		entries := b.GetEntries()
		assert.Equal(t, written, filterNil(entries))
	})

	t.Run("overwrite", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			next = int(i)
			written = append(written, next)
			b.Write(next)
		}

		entries := b.GetEntries()
		assert.Equal(t, written[5:], filterNil(entries))
	})

}
func filterNil(in []interface{}) []interface{} {
	var result []interface{}
	for _, v := range in {
		if v != nil {
			result = append(result, v)
		}
	}
	return result
}
