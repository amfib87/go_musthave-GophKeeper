package crypto_test

import (
	"fmt"
	"testing"

	"github.com/amfib87/go_musthave-GophKeeper/internal/crypto"
)

func TestEncryptData(t *testing.T) {
	tests := []struct {
		name           string
		data           []byte
		masterPassword []byte
		wantErr        bool
	}{
		{
			name:           "normal case",
			data:           []byte("hello world"),
			masterPassword: []byte("mysecretpassword"),
			wantErr:        false,
		},
		{
			name:           "empty data",
			data:           []byte(""),
			masterPassword: []byte("password"),
			wantErr:        false,
		},
		{
			name:           "long data",
			data:           []byte("this is a very long string that should be encrypted properly"),
			masterPassword: []byte("anotherpassword"),
			wantErr:        false,
		},
		{
			name:           "nil data",
			data:           nil,
			masterPassword: []byte("password"),
			wantErr:        false,
		},
		{
			name:           "empty password",
			data:           []byte("data"),
			masterPassword: []byte(""),
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := crypto.EncryptData(tt.data, tt.masterPassword)

			if tt.wantErr {
				if err == nil {
					t.Errorf("EncryptData() error = nil, wantErr = true")
				}
				return
			} else {
				if err != nil {
					t.Fatalf("EncryptData() error = %v, wantErr = false", err)
				}

				// Проверяем, что зашифрованные данные не пустые и длиннее исходных
				if len(encrypted) == 0 {
					t.Error("EncryptData() returned empty encrypted data")
				}
				if len(encrypted) <= len(tt.data) {
					t.Error("Encrypted data should be longer than original data")
				}

				// Дешифруем и проверяем, что получили исходные данные
				decrypted, err := crypto.DecryptData(encrypted, tt.masterPassword)
				if err != nil {
					t.Errorf("DecryptData() error = %v", err)
				}

				if !equalBytes(decrypted, tt.data) {
					t.Errorf("Decrypted data %v does not match original %v", decrypted, tt.data)
				}
			}
		})
	}
}

func TestDecryptData(t *testing.T) {
	tests := []struct {
		name           string
		encryptedData  []byte
		masterPassword []byte
		wantErr        bool
		expectedData   []byte
	}{
		{
			name:           "valid decryption",
			encryptedData:  nil, // Будет заполнен в тесте
			masterPassword: []byte("mysecretpassword"),
			wantErr:        false,
			expectedData:   []byte("test data"),
		},
		{
			name:           "invalid data length",
			encryptedData:  []byte{1, 2, 3}, // слишком короткие данные
			masterPassword: []byte("password"),
			wantErr:        true,
			expectedData:   nil,
		},
		{
			name:           "wrong password",
			encryptedData:  nil, // Будет заполнен в тесте
			masterPassword: []byte("wrongpassword"),
			wantErr:        true,
			expectedData:   []byte("test data"),
		},
		{
			name:           "nil encrypted data",
			encryptedData:  nil,
			masterPassword: []byte("password"),
			wantErr:        true,
			expectedData:   nil,
		},
	}

	// Предварительно шифруем данные для тестов, где это нужно
	for i := range tests {
		if tests[i].name == "valid decryption" {
			encrypted, err := crypto.EncryptData(tests[i].expectedData, tests[i].masterPassword)
			if err != nil {
				t.Fatalf("failed to encrypt test data: %v", err)
			}

			tests[i].encryptedData = encrypted

		} else if tests[i].name == "wrong password" {
			// Шифруем с правильным паролем
			encrypted, err := crypto.EncryptData(tests[i].expectedData, []byte("mysecretpassword"))
			if err != nil {
				t.Fatalf("failed to encrypt test data: %v", err)
			}
			tests[i].encryptedData = encrypted
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decrypted, err := crypto.DecryptData(tt.encryptedData, tt.masterPassword)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DecryptData() error = nil, wantErr = true")
				}
				return
			}

			if err != nil {
				t.Fatalf("DecryptData() error = %v, wantErr = false", err)
			}

			if !equalBytes(decrypted, tt.expectedData) {
				t.Errorf("Decrypted data %v does not match expected %v", decrypted, tt.expectedData)
			}
		})
	}
}

func TestEncryptDecryptConsistency(t *testing.T) {
	testCases := []struct {
		data           []byte
		masterPassword []byte
	}{
		{data: []byte("simple text"), masterPassword: []byte("password123")},
		{data: []byte(""), masterPassword: []byte("empty")},
		{data: []byte{0x00, 0x01, 0xFF, 0x80}, masterPassword: []byte("binary")},
		{data: make([]byte, 1000), masterPassword: []byte("longpassword")},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("case_%d", i), func(t *testing.T) {
			encrypted, err := crypto.EncryptData(tc.data, tc.masterPassword)
			if err != nil {
				t.Fatalf("EncryptData failed: %v", err)
			}

			decrypted, err := crypto.DecryptData(encrypted, tc.masterPassword)
			if err != nil {
				t.Fatalf("DecryptData failed: %v", err)
			}

			if !equalBytes(tc.data, decrypted) {
				t.Errorf("Data mismatch after encrypt/decrypt cycle. Original: %v, Decrypted: %v", tc.data, decrypted)
			}
		})
	}
}

// Вспомогательная функция для сравнения байтовых массивов
func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
