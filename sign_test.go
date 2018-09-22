package jwt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"math"
	"math/big"
	"testing"
)

func TestECDSASign(t *testing.T) {
	const key, value = "N", ";;Q.;;"

	var c Claims
	c.Set = map[string]interface{}{key: value}
	token, err := c.ECDSASign("ES384", testKeyEC384)
	if err != nil {
		t.Fatal("sign error:", err)
	}

	got, err := ECDSACheck(token, &testKeyEC384.PublicKey)
	if err != nil {
		t.Fatal("check error:", err)
	}
	if s := got.Set[key]; s != value {
		t.Errorf("got value %q for %q, want %q", s, key, value)
	}
}

func TestHMACSign(t *testing.T) {
	var c Claims
	c.Subject = "the world's greatest secret agent"
	got, err := c.HMACSign("HS512", []byte("guest"))
	if err != nil {
		t.Fatal(err)
	}

	want := "eyJhbGciOiJIUzUxMiJ9.eyJzdWIiOiJ0aGUgd29ybGQncyBncmVhdGVzdCBzZWNyZXQgYWdlbnQifQ.6shd8lGY9wOn9NghWeVAwRFtTE9Y-HtYy3PFxPc2ulahSq2HMOR5b8T0OhUCnZzM0svC6VH3hgh8fACD_30ubQ"
	if s := string(got); s != want {
		t.Errorf("got %q, want %q", s, want)
	}
}

func TestRSASign(t *testing.T) {
	c := &Claims{
		Set: map[string]interface{}{
			"iss": "malory",
		},
	}
	got, err := c.RSASign("RS384", testKeyRSA2048)
	if err != nil {
		t.Fatal(err)
	}

	want := "eyJhbGciOiJSUzM4NCJ9.eyJpc3MiOiJtYWxvcnkifQ.KuGs2gecLlfub_m7PcD_6EzQe35DT7MNZwhi2R9bsPgmloi47r3wRdVXEdtABGQeeUz3dOuPOQ20SuWbDTDetW7u6pRjsvjqN14-XKiWQJkjIO1jKkoAUUeIo3k-V65DB6JJHZpNhe4MTv_3JI52wAMH91zjdhP4Aado8Cd-DVW7pdgrHjjA7jfWyXsHcjQmzvzIdBSLOiNtAQsUAaAXeM9s-YCCH0ODMhYO9GMYk195TktbjVKMovjjTW-yC1SbNVGMD8m9-y2u-xX7Nmd2T6ArO4u0HAE6LYTBzn0sknTz_lU7rt3TCK2dCqDAhTXu2cbjrV3cu-1K_rSxHcRVLg"
	if s := string(got); s != want {
		t.Errorf("got %q, want %q", s, want)
	}
}

func TestSignAlgWrong(t *testing.T) {
	_, err := new(Claims).ECDSASign(RS512, testKeyEC256)
	if err != ErrAlgUnk {
		t.Errorf("RSA alg for ECDSA got error %v, want %v", err, ErrAlgUnk)
	}
	_, err = new(Claims).HMACSign(RS512, nil)
	if err != ErrAlgUnk {
		t.Errorf("RSA alg for HMAC got error %v, want %v", err, ErrAlgUnk)
	}
	_, err = new(Claims).RSASign(HS512, testKeyRSA1024)
	if err != ErrAlgUnk {
		t.Errorf("HMAC alg for RSA got error %v, want %v", err, ErrAlgUnk)
	}
}

func TestSignHashNotLinked(t *testing.T) {
	alg := "HB2b256"
	if _, ok := HMACAlgs[alg]; ok {
		t.Fatalf("non-standard alg %q present", alg)
	}
	HMACAlgs[alg] = crypto.BLAKE2b_256
	defer delete(HMACAlgs, alg)

	_, err := new(Claims).HMACSign(alg, nil)
	if err != errHashLink {
		t.Errorf("got error %v, want %v", err, errHashLink)
	}
}

func TestSignBrokenClaims(t *testing.T) {
	// JSON does not allow NaN
	n := NumericTime(math.NaN())

	c := new(Claims)
	c.Issued = &n
	_, err := c.ECDSASign(ES256, testKeyEC256)
	if _, ok := err.(*json.UnsupportedValueError); !ok {
		t.Errorf("HMAC got error %#v, want json.UnsupportedValueError", err)
	}
	_, err = c.HMACSign(HS256, nil)
	if _, ok := err.(*json.UnsupportedValueError); !ok {
		t.Errorf("HMAC got error %#v, want json.UnsupportedValueError", err)
	}

	c = new(Claims)
	c.Set = map[string]interface{}{"iss": n}
	_, err = c.RSASign(RS256, testKeyRSA1024)
	if _, ok := err.(*json.UnsupportedValueError); !ok {
		t.Errorf("RSA got error %#v, want json.UnsupportedValueError", err)
	}
}

func TestECDSAKeyBroken(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	key.Params().N = new(big.Int)
	_, err = new(Claims).ECDSASign(ES512, key)
	if err == nil || err.Error() != "zero parameter" {
		t.Errorf("got error %q, want zero parameter", err)
	}
}

func TestRSASignKeyTooSmall(t *testing.T) {
	// can't sign 512 bits with a 512-bit RSA key
	key := mustParseRSAKey(`-----BEGIN RSA PRIVATE KEY-----
MIIBPAIBAAJBAIRyPiLmS7ta5bS6eEqUb1IZhxYJ/sB+Hq/uV3xIEcu075uE0mr9
xSUHcztAcHwEYE/JF0Zc5HS++ALadTK2qZUCAwEAAQJARsqdRaAcOG70Oi4034AJ
JDO6zV/YR2Dh3B0jq60FvgAVLYKDJ7klDpeqmLB64q2IXfnoVtJDjXSTpA6qyNvG
/QIhAOWfFSLq07Ock4gjGy7qeT3Tpa/uYmRuqk90jEfn2/oDAiEAk6lYDZ1DdXmY
4cnQu3Q8A/ZHW52uFR76mLi8FihzRocCIQCWGgT+G1WibvM+JfzKEXqKAQWpWQK2
tmTcpcph4t44swIhAIzdu8PZKHbUlvWnqzp5S5vYAgEzrtQ1Zon1inF1C2vXAiEA
uRVZaJLTfpQ+n88IcdG4WPKnRZqxGnrq3DjtIvFrBlM=
-----END RSA PRIVATE KEY-----`)

	_, err := new(Claims).RSASign(RS512, key)
	if err != rsa.ErrMessageTooLong {
		t.Errorf("got error %q, want %q", err, rsa.ErrMessageTooLong)
	}
}

func TestUseAlg(t *testing.T) {
	/// test all standard algorithms
	algs := make(map[string]crypto.Hash)
	for alg, hash := range ECDSAAlgs {
		algs[alg] = hash
	}
	for alg, hash := range HMACAlgs {
		algs[alg] = hash
	}
	for alg, hash := range RSAAlgs {
		algs[alg] = hash
	}

	for alg, wantHash := range algs {
		header, hash, err := useAlg(alg, algs)
		if err != nil {
			t.Errorf("error for %q: %s", alg, err)
			continue
		}

		if hash != wantHash {
			t.Errorf("wrong hash for %q", alg)
		}

		headerJSON, err := encoding.DecodeString(header)
		if err != nil {
			t.Errorf("malformed header for %q: %s", alg, err)
			continue
		}
		m := make(map[string]interface{})
		if err := json.Unmarshal(headerJSON, &m); err != nil {
			t.Errorf("malformed header for %q: %s", alg, err)
			continue
		}
		if s, ok := m["alg"].(string); !ok || s != alg {
			t.Errorf("got alg %#v for %q", m["alg"], alg)
		}
	}
}
