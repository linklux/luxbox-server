package component

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"

	"github.com/google/uuid"
)

type StringGenerator struct{}

func (this *StringGenerator) GenerateString(length int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	ret := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))

		if err != nil {
			return "", err
		}

		ret[i] = letters[num.Int64()]
	}

	return string(ret), nil
}

func (this *StringGenerator) GenerateBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)

	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (this *StringGenerator) GenerateUuid4() string {
	uuid := uuid.New()
	str := hex.EncodeToString(uuid[:])

	// TODO There must be a better way to achieve this...
	indices := []int{8, 12 + 1, 16 + 2, 20 + 3} // format as 8-4-4-4-12

	for _, index := range indices {
		tmp := str[:index] + "-" + str[index:]
		str = tmp
	}

	return str
}
