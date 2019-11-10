package crypto

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKeys(t *testing.T) {
	privateA, publicA, err := GenerateKeys()
	assert.NoError(t, err)
	privateB, publicB, err := GenerateKeys()
	assert.NoError(t, err)

	pkA := GeneratePublicKey(publicA.Bytes())
	fmt.Println(publicA.Bytes())
	pkB := GeneratePublicKey(publicB.Bytes())

	secretA, err := ComputeSecret(privateA, pkB)
	assert.NoError(t, err)
	secretB, err := ComputeSecret(privateB, pkA)
	fmt.Println(len(secretA))
	fmt.Println(secretA)
	fmt.Println(len(secretB))
	fmt.Println(secretB)
	fmt.Println(bytes.Equal(secretA, secretB))
}
