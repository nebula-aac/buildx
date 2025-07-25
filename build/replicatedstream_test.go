package build

import (
	"bytes"
	"crypto/rand"
	"io"
	mathrand "math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func generateRandomData(size int) []byte {
	data := make([]byte, size)
	rand.Read(data)
	return data
}
func TestSyncMultiReaderParallel(t *testing.T) {
	data := generateRandomData(1024 * 1024)
	source := bytes.NewReader(data)
	mr := NewSyncMultiReader(source)

	var wg sync.WaitGroup
	numReaders := 10
	bufferSize := 4096 * 4

	readers := make([]io.ReadCloser, numReaders)

	for i := range numReaders {
		readers[i] = mr.NewReadCloser()
	}

	for i := range numReaders {
		wg.Add(1)
		go func(readerId int) {
			defer wg.Done()
			reader := readers[readerId]
			defer reader.Close()

			totalRead := 0
			buf := make([]byte, bufferSize)
			for totalRead < len(data) {
				// Simulate random read sizes
				readSize := mathrand.Intn(bufferSize) // #nosec G404 -- ignore "Use of weak random number generator (math/rand instead of crypto/rand)"
				n, err := reader.Read(buf[:readSize])

				if n > 0 {
					assert.Equal(t, data[totalRead:totalRead+n], buf[:n], "Reader %d mismatch", readerId)
					totalRead += n
				}

				if err == io.EOF {
					assert.Equal(t, len(data), totalRead, "Reader %d EOF mismatch", readerId)
					return
				}

				assert.NoError(t, err, "Reader %d error", readerId)

				// #nosec G404 -- ignore "Use of weak random number generator (math/rand instead of crypto/rand)"
				if mathrand.Intn(1000) == 0 {
					t.Logf("Reader %d closing", readerId)
					// Simulate random close
					return
				}

				// Simulate random timing between reads
				time.Sleep(time.Millisecond * time.Duration(mathrand.Intn(5))) // #nosec G404 -- ignore "Use of weak random number generator (math/rand instead of crypto/rand)"
			}

			assert.Equal(t, len(data), totalRead, "Reader %d total read mismatch", readerId)
		}(i)
	}

	wg.Wait()
}
