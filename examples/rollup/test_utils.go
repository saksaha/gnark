/*
Copyright © 2020 ConsenSys

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rollup

import (
	"hash"
	"math/rand"
	"testing"
	"fmt"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
)

func createAccount(i int) (Account, eddsa.PrivateKey) {
	var acc Account
	var rnd fr.Element
	var privkey eddsa.PrivateKey

	// create account, the i-th account has an balance of 20+i
	acc.index = uint64(i)
	acc.nonce = uint64(i)
	acc.balance.SetUint64(uint64(i) + 20)
	rnd.SetUint64(uint64(i))
	src := rand.NewSource(int64(i))
	r := rand.New(src)

	// TODO handle error
	privkey, _ = eddsa.GenerateKey(r)
	acc.pubKey = privkey.PublicKey
	fmt.Printf("createAccount(): i: %d(%d), pubKey: %v, privkey: %v, accountBalance: %v, nonce: %d\n",
		i, uint64(i), privkey, acc.pubKey, acc.balance, acc.nonce)

	return acc, privkey
}

// Returns a newly created operator and tha private keys of the associated accounts
func createOperator(nbAccounts int) (Operator, []eddsa.PrivateKey) {

	operator := NewOperator(nbAccounts)

	userAccounts := make([]eddsa.PrivateKey, nbAccounts)

	// randomly fill the accounts
	for i := 0; i < nbAccounts; i++ {

		acc, privkey := createAccount(i)

		// fill the index map of the operator
		b := acc.pubKey.A.X.Bytes()
		operator.AccountMap[string(b[:])] = acc.index

		// fill user accounts list
		userAccounts[i] = privkey
		baccount := acc.Serialize()

		copy(operator.State[SizeAccount*i:], baccount)

		// create the list of hashes of account
		operator.h.Reset()
		operator.h.Write(acc.Serialize())
		buf := operator.h.Sum([]byte{})
		copy(operator.HashState[operator.h.Size()*i:], buf)
	}

	fmt.Printf("\ncreateOperator(): state(%d): %d\nhashSate(%d): %d\n",
		len(operator.State), operator.State, len(operator.HashState), operator.HashState,)
	return operator, userAccounts

}

func compareAccount(t *testing.T, acc1, acc2 Account) {

	if acc1.index != acc2.index {
		t.Fatal("Incorrect index")
	}
	if acc1.nonce != acc2.nonce {
		t.Fatal("Incorrect nonce")
	}
	if !acc1.balance.Equal(&acc2.balance) {
		t.Fatal("Incorrect balance")
	}
	if !acc1.pubKey.A.X.Equal(&acc2.pubKey.A.X) {
		t.Fatal("Incorrect public key (X)")
	}
	if !acc1.pubKey.A.Y.Equal(&acc2.pubKey.A.Y) {
		t.Fatal("Incorrect public key (Y)")
	}

}

func compareHashAccount(t *testing.T, h []byte, acc Account, hFunc hash.Hash) {

	hFunc.Reset()
	_, err := hFunc.Write(acc.Serialize())
	if err != nil {
		t.Fatal(err)
	}
	res := hFunc.Sum([]byte{})
	if len(res) != len(h) {
		t.Fatal("Error comparing hashes (different lengths)")
	}
	for i := 0; i < len(res); i++ {
		if res[i] != h[i] {
			t.Fatal("Error comparing hashes (different content)")
		}
	}
}
