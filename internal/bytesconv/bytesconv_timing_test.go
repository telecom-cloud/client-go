package bytesconv

import (
	"strings"
	"testing"
)

// For test only, but it will import golang.org/x/net/http.
// So comment out all this code. Keep this for the full context.
//func BenchmarkValidHeaderFiledValueTable(b *testing.B) {
//	// Test all characters
//	allBytes := make([]string, 0)
//	for i := 0; i < 256; i++ {
//		allBytes = append(allBytes, string([]byte{byte(i)}))
//	}
//
//	for i := 0; i < b.N; i++ {
//		for _, s := range allBytes {
//			_ = httpguts.ValidHeaderFieldValue(s)
//		}
//	}
//}

func BenchmarkValidHeaderFiledValueTableCrafter(b *testing.B) {
	// Test all characters
	allBytes := make([]byte, 0)
	for i := 0; i < 256; i++ {
		allBytes = append(allBytes, byte(i))
	}

	for i := 0; i < b.N; i++ {
		for _, s := range allBytes {
			_ = func() bool {
				return ValidHeaderFieldValueTable[s] != 0
			}()
		}
	}
}

func BenchmarkNewlineToSpace(b *testing.B) {
	// Test all characters
	allBytes := make([]byte, 0)
	for i := 0; i < 256; i++ {
		allBytes = append(allBytes, byte(i))
	}
	headerNewlineToSpace := strings.NewReplacer("\n", " ", "\r", " ")

	for i := 0; i < b.N; i++ {
		_ = headerNewlineToSpace.Replace(string(allBytes))
	}
}

func BenchmarkNewlineToSpaceCrafter01(b *testing.B) {
	// Test all characters
	allBytes := make([]byte, 0)
	for i := 0; i < 256; i++ {
		allBytes = append(allBytes, byte(i))
	}

	for i := 0; i < b.N; i++ {
		filteredVal := make([]byte, 0, len(allBytes))
		for i := 0; i < len(allBytes); i++ {
			filteredVal = append(filteredVal, NewlineToSpaceTable[allBytes[i]])
		}
		_ = filteredVal
	}
}

func BenchmarkNewlineToSpaceCrafter02(b *testing.B) {
	// Test all characters
	allBytes := make([]byte, 0)
	for i := 0; i < 256; i++ {
		allBytes = append(allBytes, byte(i))
	}

	for i := 0; i < b.N; i++ {
		filteredVal := make([]byte, len(allBytes))
		copy(filteredVal, allBytes)
		for ii := 0; ii < len(allBytes); ii++ {
			filteredVal[ii] = NewlineToSpaceTable[filteredVal[ii]]
		}
	}
}
