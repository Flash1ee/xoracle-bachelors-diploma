package command

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/fatih/color"

	"xoracle/cmd/config"
)

var InfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Verify file checksums",
	RunE:  infoCommand,
}

var dirFlag string

func init() {
	RootCmd.AddCommand(InfoCmd)
	InfoCmd.Flags().StringP("dir", "d", "./src", "Directory to save the file blocks")
}

func infoCommand(cmd *cobra.Command, args []string) error {
	var err error
	cfg, err = loadConfig()
	if err != nil {
		log.Fatal("error load config")
	}
	//dir := args[0]
	dir := "./src"
	if err := info(&cfg, dir); err != nil {
		return fmt.Errorf("backup failed: %v", err)
	}

	fmt.Println("Backup completed successfully.")
	return nil
}

func info(cfg *config.Config, dir string) error {
	// Получение списка файлов в указанной папке
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Printf("Failed to read directory: %v\n", err)
		os.Exit(1)
	}

	// Пробегаемся по каждому файлу в указанной папке
	for _, file := range files {
		// Проверка контрольной суммы файла
		err := verifyChecksum(cfg, dir, file.Name())
		if err != nil {
			red := color.New(color.FgRed).SprintFunc()
			fmt.Printf("Checksum mismatch for file %s: %v\n", file.Name(), red(err))
		} else {
			green := color.New(color.FgGreen).SprintFunc()
			fmt.Printf("Checksum match for file %s\n", green(file.Name()))
		}
	}
	return nil
}

func verifyChecksum(cfg *config.Config, dir, fileName string) error {
	filePath := filepath.Join(dir, fileName)

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Вычисляем контрольную сумму файла
	actualChecksum, err := fileChecksum(file)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum for file: %v", err)
	}

	// Получение ожидаемой контрольной суммы из конфигурации или другого источника
	expectedChecksum := getExpectedChecksum(fileName, cfg)

	// Сравниваем контрольные суммы
	if expectedChecksum != actualChecksum {
		return fmt.Errorf("checksum mismatch: expected = %s, actual = %s", expectedChecksum, actualChecksum)
	}

	return nil
}

func getExpectedChecksum(filePath string, cfg *config.Config) string {
	var res string
	for _, val := range cfg.Files {
		if filePath == val.Filename {
			res = val.Checksum
			return res
		}
	}
	return res
}
