package crypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIzzyWizzy(t *testing.T) {
	key, _ := hex.DecodeString("2b7e151628aed2a6abf7158809cf4f3c")
	plainText, _ := hex.DecodeString("6bc1bee22e409f96e93d7e117393172aae2d8a571e03ac9c9eb76fac45" +
		"af8e5130c81c46a35ce411e5fbc1191a0a52eff69f2445df4f9b17ad2b417be66c3710")
	expected, _ := hex.DecodeString("3ad77bb40d7a3660a89ecaf32466ef97f5d3d58503b9699de785895a96fdb" +
		"aaf43b1cd7f598ece23881b00e3ed0306887b0c785e27e8ad3f8223207104725dd4")

	input := []struct {
		plainT  []byte
		cipherT []byte
		iw      IzzyWizzy
	}{
		{plainText, expected, IzzyWizzy{Key: key}},
	}

	for _, tt := range input {
		t.Run("", func(t *testing.T) {
			tt := tt
			actual, err := tt.iw.Encrypt(tt.plainT)
			require.NoError(t, err)
			assert.EqualValues(t, actual[:len(actual)-16], tt.cipherT)
			decr, err := tt.iw.Decrypt(actual)
			require.NoError(t, err)
			assert.EqualValues(t, decr, tt.plainT)
		})
	}
}
