package goidagif

import (
	"os"
	"testing"
)

func TestGenerateGIF(t *testing.T) {
	outputPath := "storage/test_output.gif"
	text := "Тестовое сообщение"

	// Запускаем функцию
	err := GenerateGIF(outputPath, text)
	if err != nil {
		t.Fatalf("Failed to generate GIF: %v", err)
	}

	// Проверяем, был ли создан файл
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file %s was not created", outputPath)
	}
}
