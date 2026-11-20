package horizon

import (
	"fmt"
	"math/big"
	"testing"
)

// FuzzParseAmount checks ParseAmount never panics and that every accepted
// amount is a non-negative integer that survives a format/parse round trip.
func FuzzParseAmount(f *testing.F) {
	for _, s := range []string{
		"", "0", "0.0000000", "1", "12.3400000", ".5", "5.", "-1", "-0.0000001",
		"0.00000001", "922337203685.4775807", "340282366920938463463374607431768211455",
		" 7.5 ", "1e5", "+1", "1_0", "0x10", "1.2.3", "١٢٣",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		n, err := ParseAmount(s)
		if err != nil {
			if n != nil {
				t.Fatalf("ParseAmount(%q) returned a value with error %v", s, err)
			}
			return
		}
		if n == nil || n.Sign() < 0 {
			t.Fatalf("ParseAmount(%q) = %v, want non-negative", s, n)
		}
		// The canonical decimal for n must parse back to n.
		q, r := new(big.Int).QuoRem(n, big.NewInt(1e7), new(big.Int))
		canon := fmt.Sprintf("%s.%07d", q, r.Int64())
		back, err := ParseAmount(canon)
		if err != nil || back.Cmp(n) != 0 {
			t.Fatalf("ParseAmount(%q) = %s but canonical %q parses to %v (%v)", s, n, canon, back, err)
		}
	})
}
