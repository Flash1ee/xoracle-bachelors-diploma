func restoreFiles(cmd *cobra.Command, args []string) error {
	var err error
	cfg, err = loadConfig()
	if err != nil {
		log.Fatal("error load config")
	}
	if err := restore(&cfg); err != nil {
		return fmt.Errorf("backup failed: %v", err)
	}

	fmt.Println("Restore completed successfully.")
	return nil
}
func restore(config *config.Config) error {
	for _, superblock := range config.SuperBlocks {
		// Read superblock from file
		superblockData, err := readSuperblock(superblock.Location)
		if err != nil {
			return fmt.Errorf("failed to read superblock %s: %v", superblock.Location, err)
		}
		// Get filename from superblock location
		fileName := getLastWordWithoutExtension(superblock.Location)

		// Create restore directory if not exists
		restoreDir := "./restore"
		err = os.MkdirAll(restoreDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create restore directory: %v", err)
		}

		// Create destination file
		restorePath := filepath.Join(restoreDir, fileName)
		file, err := os.Create(restorePath)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to create file %q: %v", restorePath, err))
		}
		prevBlock := make([]byte, superblock.BlockSize)
		for i := 0; i < len(superblock.Blocks); i++ {
			if i == len(superblock.Blocks)-1 {
				//fmt.Println("block  restore ", sb.Blocks[i-1])
				fmt.Println("check padding")
				xorBlock(superblockData.Blocks[i], prevBlock)
				superblockData.Blocks[i] = superblockData.Blocks[i][:superblock.Blocks[len(superblock.Blocks)-1].Size]
			} else {
				log.Println(fmt.Sprintf("block num = %v", i))
				xorBlock(superblockData.Blocks[i], prevBlock)
				prevBlock = superblockData.Blocks[i]
			}
			_, err = file.Write(superblockData.Blocks[i])
			if err != nil {
				file.Close()
				os.Remove(restorePath)
				log.Fatal(fmt.Errorf("failed to write block to file %q: %v", restorePath, err))
			}
		}
	}
	return nil
}
func xorBlock(dst, src []byte) {
	for i := 0; i < len(dst); i++ {
		dst[i] ^= src[i]
	}
}

func removePadding(block []byte) []byte {
	cnt := 0
	i := len(block) - 1
	for ; i >= 0; i-- {
		if block[i] != 0 {
			break
		}
		cnt += 1
	}
	return block[:i+1]
}

func getLastWordWithoutExtension(filePath string) string {
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	words := strings.Fields(name)
	if len(words) > 0 {
		return words[len(words)-1]
	}
	return ""
}


func readSuperblock(filePath string) (*superBlock, error) {
	// Открываем файл суперблока для чтения
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open superblock file %s: %v", filePath, err)
	}
	defer file.Close()

	// Читаем данные из файла
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read superblock file %s: %v", filePath, err)
	}

	// Декодируем данные из JSON
	var sb superBlock
	err = json.Unmarshal(data, &sb)
	if err != nil {
		return nil, fmt.Errorf("failed to decode superblock from file %s: %v", filePath, err)
	}

	return &sb, nil
}

func saveRestoredBlocks(restoreDir, filename string, blocks []byte) error {
	filePath := filepath.Join(restoreDir, filename)
	err := ioutil.WriteFile(filePath, blocks, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to save restored blocks to file %s: %v", filePath, err)
	}
	return nil
}
