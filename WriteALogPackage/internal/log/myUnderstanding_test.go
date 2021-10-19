package log

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMyUnderstanding(t *testing.T) {
	f, err := os.Create("mystore.txt")
	require.NoError(t, err)
	myStore, err := newStore(f)
	require.NoError(t, err)

	message := []byte("hello from records")
	_, pos, err := myStore.Append(message)
	require.NoError(t, err)

	readValue, err := myStore.Read(pos)
	require.NoError(t, err)

	message2 := []byte("this is another message")
	_, pos, err = myStore.Append(message2)
	require.NoError(t, err)

	readValue2, err := myStore.Read(pos)
	require.NoError(t, err)

	fmt.Printf("%s\n", string(readValue))
	fmt.Printf("%s\n", string(readValue2))
	require.NoError(t, err)
}
