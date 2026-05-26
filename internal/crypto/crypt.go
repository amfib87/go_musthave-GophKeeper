package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

// EncryptData шифрует данные с использованием мастер‑пароля.
// Параметры:
//
//	data — данные для шифрования.
//	masterPassword — мастер‑пароль пользователя.
//
// Возвращает:
//
//	[]byte — зашифрованные данные.
//	error — ошибка при шифровании (nil при успехе).
func EncryptData(data, masterPassword []byte) ([]byte, error) {
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}

	key := pbkdf2.Key(masterPassword, salt, 4096, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Вычисляем хеш исходных данных
	hash := sha256.Sum256(data)

	// Создаём слайс с достаточным местом для хеша и данных
	dataWithHash := make([]byte, len(hash)+len(data))

	// Преобразовываем массив [32]byte в слайс []byte через [:], чтобы можно было использовать copy()
	copy(dataWithHash, hash[:]) // <-- ключевое исправление
	copy(dataWithHash[len(hash):], data)

	ciphertext := make([]byte, aes.BlockSize+len(dataWithHash))
	iv := ciphertext[:aes.BlockSize]
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], dataWithHash)

	result := make([]byte, len(salt)+len(ciphertext))
	copy(result, salt)
	copy(result[len(salt):], ciphertext)
	return result, nil
}

// DecryptData дешифрует данные с использованием мастер‑пароля.
// Параметры:
//
//	encryptedData — зашифрованные данные.
//	masterPassword — мастер‑пароль пользователя.
//
// Возвращает:
//
//	[]byte — дешифрованные данные.
//	error — ошибка при дешифровании (nil при успехе).
func DecryptData(encryptedData, masterPassword []byte) ([]byte, error) {
	if len(encryptedData) < 8+aes.BlockSize {
		return nil, fmt.Errorf("invalid encrypted data")
	}

	salt := encryptedData[:8]
	ciphertext := encryptedData[8:]

	key := pbkdf2.Key(masterPassword, salt, 4096, 32, sha256.New)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	encrypted := ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	plaintextWithHash := make([]byte, len(encrypted))
	stream.XORKeyStream(plaintextWithHash, encrypted)

	// Проверяем целостность: первые 32 байта — хеш исходных данных
	if len(plaintextWithHash) < 32 {
		return nil, fmt.Errorf("decrypted data too short for hash")
	}

	expectedHash := plaintextWithHash[:32]
	originalData := plaintextWithHash[32:]
	actualHash := sha256.Sum256(originalData)

	// Сравниваем массивы [32]byte напрямую
	if !bytes.Equal(expectedHash, actualHash[:]) { // <-- преобразуем actualHash в слайс для сравнения
		return nil, fmt.Errorf("integrity check failed: invalid password or corrupted data")
	}

	return originalData, nil
}
